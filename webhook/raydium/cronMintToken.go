package raydium

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/inc-backend/3rd-libs/3rd/slack"

	go_incognito "github.com/inc-backend/go-incognito"
	config "github.com/orgs/Solana-Incognito-Bridge/ognito-service/conf"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/dao"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/helpers/rpccaller"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/models"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service/sol"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type CronMintToken struct {
	bc        *go_incognito.PublicIncognito
	solClient *sol.Client
	logger    *zap.Logger
	conf      *config.Config

	shield    *go_incognito.ShieldUnShield
	trans     *go_incognito.Trans
	wallet    *go_incognito.Wallet
	blockInfo *go_incognito.BlockInfo
	dao       *dao.Raydium
}

func NewCronMintToken(
	bc *go_incognito.PublicIncognito,
	solClient *sol.Client,
	dao *dao.Raydium,
	conf *config.Config,
	logger *zap.Logger,

) *CronMintToken {
	blockInfo := go_incognito.NewBlockInfo(bc)
	shield := go_incognito.NewShieldUnShield(bc, blockInfo)
	wallet := go_incognito.NewWallet(bc, blockInfo)
	trans := go_incognito.NewTrans(bc)

	return &CronMintToken{solClient: solClient, logger: logger, conf: conf, bc: bc, blockInfo: blockInfo, shield: shield, wallet: wallet, trans: trans, dao: dao}
}

func (s *CronMintToken) Start() {
	logs, _, _ := s.dao.List(0, 99999,
		map[string]string{
			"trade_type": fmt.Sprintf("%v", int(models.Raydium)),
			"status.in": fmt.Sprintf("%v, %v",
				int(models.WithdrawSuccess),
				int(models.MintRequested)),
			"err_count.lessthan": strconv.Itoa(s.conf.PencakeSwapProtocol.RetryTimes),
		},
	)
	for _, log := range logs {
		s.process(log)
	}
}

func (s *CronMintToken) process(log *models.Raydium) {
	if log.Status == models.WithdrawSuccess {
		s.mintToken(log)
	} else if log.Status == models.MintRequested {
		s.checkTxStatus(log)
	}
}

func (s *CronMintToken) mintToken(log *models.Raydium) {

	mintResp, err := s.callIssuingSolReq(
		s.conf.Incognito.MasterMintPrivateKey,
		log.DestTokens,
		log.WithdrawTx,
		"createandsendtxwithissuingsolreq",
	)
	if err != nil {
		err = errors.Wrap(err, "s.callIssuingSolReq")
		go s.dao.TrackHistory(log, models.HistoryStatusFailure, fmt.Sprintf("CronMintToken tradeId %v", log.ID), err.Error())
		return
	}

	txID, found := mintResp["TxID"]

	fmt.Println("txID, found: ", txID, found)

	go s.dao.TrackHistory(log, models.HistoryStatusSuccess, fmt.Sprintf("CronMintToken tradeId %v, tx: %v", log.ID, txID), "submit mint successful!")

	log.MintTx = txID.(string)
	log.Status = models.MintRequested

	s.dao.Update(log)

}

func (s *CronMintToken) checkTxStatus(log *models.Raydium) {
	tx, err := s.trans.GetTransactionDetailByTxHash(log.MintTx)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("checkTxStatus(%s)", log.MintTx))
		go s.dao.TrackHistory(log, models.HistoryStatusFailure, fmt.Sprintf("cronMintToken tradeId %v", log.ID), err.Error())

		log.ErrCount += 1
		if log.ErrCount == models.MaxErr+10 {
			log.Status = models.MintFailed
		}

		s.dao.Update(log)
		return
	}

	if !tx.IsInBlock {
		err = errors.New(fmt.Sprintf("TxId is empty, Tx in block = %v ", strconv.FormatBool(tx.IsInBlock)))
		err = errors.Wrap(err, fmt.Sprintf("tx.IsInBlock: %v", tx.IsInBlock))
		go s.dao.TrackHistory(log, models.HistoryStatusFailure, fmt.Sprintf("cronMintToken tradeId %v", log.ID), err.Error())

		log.ErrCount += 1
		if log.ErrCount == models.MaxErr+10 {
			log.Status = models.MintFailed
		}
		s.dao.Update(log)
		return
	}

	bridgeStatus, err := s.trans.GetBridgeReqWithStatus(tx.Hash)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("GetBridgeReqWithStatus(%v)", tx.Hash))
		go s.dao.TrackHistory(log, models.HistoryStatusFailure, fmt.Sprintf("cronMintToken tradeId %v", log.ID), err.Error())

		log.ErrCount += 1
		if log.ErrCount == models.MaxErr+10 {
			log.Status = models.MintFailed
		}

		s.dao.Update(log)
		return
	}

	fmt.Println("GetBridgeReqWithStatus, tx: ", tx.Hash, bridgeStatus)

	// pass:
	if bridgeStatus == 2 {
		go s.dao.TrackHistory(log, models.HistoryStatusSuccess, fmt.Sprintf("cronMintToken tradeId %v", log.ID), "uniswap trade successful!")

		log.Status = models.Minted
		log.ErrCount = 0
		s.dao.Update(log)

		msgForSlack := fmt.Sprintf(":rocket: Raydium swap successfully `%v` from `%v %v` to `%v %v` ", log.ID, log.SrcQties, log.SrcSymbol, log.OutputAmount, log.DestSymbol)

		s := slack.InitSlack(s.conf.Slack.OAuthToken, s.conf.Slack.RaydiumProtocol)
		go s.PostMsg(msgForSlack)

		return
	}

	// rejected:
	if bridgeStatus == 3 {
		go s.dao.TrackHistory(log, models.HistoryStatusFailure, fmt.Sprintf("cronMintToken tradeId %v, %v", log.ID, log.MintTx), "Tx rejected")

		log.Status = models.MintTxRejected
		log.ErrCount = 0

		s.dao.Update(log)
		//check mint result
		return
	}

	if bridgeStatus == 0 {
		go s.dao.TrackHistory(log, models.HistoryStatusFailure, fmt.Sprintf("cronMintToken tradeId %v, %v", log.ID, log.MintTx), "Tx rejected")

		log.ErrCount += 1
		if log.ErrCount == models.MaxErr+10 {
			log.Status = models.MintFailed
		}
		s.dao.Update(log)
		return
	}
}

func (s *CronMintToken) callIssuingSolReq(
	privKey string,
	incTokenIDStr string,
	TxSig string,
	method string,
) (map[string]interface{}, error) {

	type IssuingETHRes struct {
		rpccaller.RPCBaseRes
		Result interface{} `json:"Result"`
	}

	rpcClient := rpccaller.NewRPCClient()
	meta := map[string]interface{}{
		"IncTokenID": incTokenIDStr,
		"TxSig":      TxSig,
	}
	params := []interface{}{
		privKey,
		nil,
		5,
		-1,
		meta,
	}
	var res IssuingETHRes
	err := rpcClient.RPCCall(
		"",
		s.conf.Incognito.ChainEndpoint,
		"",
		method,
		params,
		&res,
	)
	if err != nil {
		return nil, err
	}

	resp, err := json.Marshal(res)

	if err != nil {
		return nil, err
	}
	fmt.Println("resp", string(resp))

	if res.RPCError != nil {
		return nil, errors.New(res.RPCError.Message)
	}
	return res.Result.(map[string]interface{}), nil
}
