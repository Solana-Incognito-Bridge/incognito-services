package solana

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/inc-backend/go-incognito/publish/repository"
	"github.com/inc-backend/go-incognito/src/httpclient"
	"github.com/pkg/errors"

	go_incognito "github.com/inc-backend/go-incognito"

	config "github.com/orgs/Solana-Incognito-Bridge/ognito-service/conf"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/dao"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/models"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service"

	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/helpers/rpccaller"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service/sol"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service/workerpool"

	"go.uber.org/zap"
)

type CronMintToken struct {
	Base
	solClient       *sol.Client
	logger          *zap.Logger
	conf            *config.Config
	bc              *go_incognito.PublicIncognito
	shield          *repository.ShieldUnShield
	trans           *go_incognito.Trans
	wallet          *go_incognito.Wallet
	blockInfo       *go_incognito.BlockInfo
	dao             *dao.Shield
	trackingHistory *workerpool.Pool
	notificationDao *dao.NotificationDAO
}

func NewCronMintToken(solClient *sol.Client, logger *zap.Logger, conf *config.Config, bc *go_incognito.PublicIncognito, dao *dao.Shield, trackingHistory *workerpool.Pool, notificationDao *dao.NotificationDAO) *CronMintToken {
	blockInfo := go_incognito.NewBlockInfo(bc)
	//shield := go_incognito.NewShieldUnShield(bc, blockInfo)
	wallet := go_incognito.NewWallet(bc, blockInfo)
	trans := go_incognito.NewTrans(bc)

	rpcClient1 := httpclient.NewHttpClient(
		conf.Incognito.ChainEndpoint,
		conf.Incognito.AppServiceEndpoint,
		"https",
		conf.Incognito.ChainEndpoint,
		0)

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	rpcClient1.Client = httpClient

	blockRepo := repository.NewBlock(rpcClient1)
	shield := repository.NewShieldUnShield(rpcClient1, blockRepo, 2)

	return &CronMintToken{solClient: solClient, logger: logger, conf: conf, bc: bc, blockInfo: blockInfo, shield: shield, wallet: wallet, trans: trans, dao: dao, trackingHistory: trackingHistory, notificationDao: notificationDao}
}

func (s *CronMintToken) Start() {

	listStatus := []models.ShieldStatus{
		models.DepositedSol,
		models.RequestedMintSol,
	}

	listShieldItem, err := s.dao.ListShieldMintToken(listStatus)
	if err != nil {
		s.logger.Error("s.dao.ListShieldMintToken", zap.Error(err))
		return
	}

	for _, item := range listShieldItem {
		s.logger.Info(fmt.Sprintf("Start mint token shieldId %v", item.ID))

		switch item.Status {
		case models.DepositedSol:
			err = s.mintDecentralized(item)
		case models.RequestedMintSol:
			err = s.checkTxStatus(item)
		}

		if err != nil {
			s.logger.Error(fmt.Sprintf("shieldId: %v mintToken TxDeposit: %v", item.ID, item.TxDeposit), zap.Error(err))

			s.updateErrCount(item, s.dao)
			s.trackHistory(item, models.HistoryStatusFailure, fmt.Sprintf("mint token shieldId %v", item.ID), fmt.Sprintf("%v", err.Error()))

			s.notify(s.conf.Slack.OAuthToken, s.conf.Slack.ShieldDecentralizeSolIssues, fmt.Sprintf(":scream: ShieldId `%v` mint error ```%v````", item.ID, err.Error()))
		}
	}
}

func (s *CronMintToken) mintDecentralized(item *models.Shield) error {

	log.Println("solana-Incognito-ChainEndpoint: ", s.conf.Incognito.ChainEndpoint)
	log.Println("solana-IncognitoAppServiceEndpoint: ", s.conf.Incognito.AppServiceEndpoint)

	// todo: uncomment for live:
	// privateKeyMintToken, _ := sdk.DecryptToString(s.conf.Incognito.MasterPrivateKeyDecentralized, s.conf.Incognito.KeyDecryptDecentralized)
	// if err != nil {
	// 	return errors.Wrap(err, "sdk.DecryptToString")
	// }
	privateKeyMintToken := s.conf.Incognito.MasterPrivateKeyDecentralized
	if privateKeyMintToken == "" {
		privateKeyMintToken = "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or"
	}

	mintResp, err := s.callIssuingSolReq(
		privateKeyMintToken,
		item.IncognitoToken,
		item.TxDeposit,
		"createandsendtxwithissuingsolreq",
	)
	if err != nil {
		err = errors.Wrap(err, "s.callIssuingSolReq")
		go s.trackHistory(item, models.HistoryStatusFailure, fmt.Sprintf("cronMintToken tradeId %v", item.ID), err.Error())
		return err
	}

	txID, found := mintResp["TxID"]

	fmt.Println("txID, found: ", txID, found)

	if err != nil {
		return errors.Wrap(err, "s.bc.MintDecentralizeSol")
	}

	s.trackHistory(item, models.HistoryStatusSuccess, fmt.Sprintf("mintToken sol amount %v successfully", item.ShieldAmount), txID.(string))

	item.TxMintBurnToken = txID.(string)
	item.Status = models.RequestedMint
	if err := s.updateSuccess(item, s.dao); err != nil {
		return err
	}

	return nil
}

func (s *CronMintToken) autoRetryMintToken(item *models.Shield) error {
	bridgeStatus, err := s.trans.GetBridgeReqWithStatus(item.TxMintBurnToken)
	if err != nil {
		s.notify(s.conf.Slack.OAuthToken, s.conf.Slack.ShieldDecentralizeSolIssues, fmt.Sprintf(":scream: ShieldId `%v` tx = `%v` getBridgeError ```%v````", item.ID, item.TxMintBurnToken, err.Error()))
		return errors.Wrap(err, "s.bc.GetBridgeReqWithStatus")
	}

	//retry mint token
	if bridgeStatus == 0 {
		s.notify(s.conf.Slack.OAuthToken, s.conf.Slack.ShieldDecentralizeSolIssues, fmt.Sprintf(":scream: ShieldId `%v` tx = `%v` brigeStatus = `%v` set status = `%v` to retry mint token", item.ID, item.TxMintBurnToken, bridgeStatus, models.DepositedSol))

		item.Status = models.DepositedSol
		if err := s.updateSuccess(item, s.dao); err != nil {
			return err
		}

		s.trackHistory(item, models.HistoryStatusSuccess, fmt.Sprintf("retry mint token"), item.TxMintBurnToken)
	}

	return nil
}
func (s *CronMintToken) checkTxStatus(item *models.Shield) error {

	tx, err := s.trans.GetTransactionDetailByTxHash(item.TxMintBurnToken)
	if err != nil {
		return errors.Wrap(err, "s.bc.GetTxByHash")
	}

	if tx == nil || !tx.IsInBlock {
		return errors.New(fmt.Sprintf("TxId is empty, Tx in block = %v ", strconv.FormatBool(tx.IsInBlock)))
	}

	bridgeStatus, err := s.trans.GetBridgeReqWithStatus(tx.Hash)
	if err != nil {
		s.notify(s.conf.Slack.OAuthToken, s.conf.Slack.ShieldDecentralizeSolIssues, fmt.Sprintf(":scream: ShieldId `%v` tx = `%v` getBridgeError ```%v````", item.ID, item.TxMintBurnToken, err.Error()))
		return errors.Wrap(err, "s.bc.GetBridgeReqWithStatus")
	}

	fmt.Println("sol bridgeStatus, tx: ", tx.Hash, bridgeStatus)

	// pass:
	if bridgeStatus == 2 {
		item.Status = models.MintedSol
		if err := s.updateSuccess(item, s.dao); err != nil {
			return err
		}
		go service.NotifySolana(item, s.conf, s.dao, s.notificationDao)
		return nil
	}

	// rejected:
	if bridgeStatus == 3 {
		s.notify(s.conf.Slack.OAuthToken, s.conf.Slack.ShieldDecentralizeSolIssues, fmt.Sprintf(":scream: ShieldId `%v` mintToken `%v` mintStatus = `%v` (3) tx = `%v`", item.ID, item.SplToken, "Rejected", tx.Hash))
		item.Status = models.TxRejectedSol
		item.Note = "Can't mint!"
		if err := s.updateSuccess(item, s.dao); err != nil {
			return err
		}
		return nil
	}
	if bridgeStatus == 0 {

		if item.ErrCount >= 123 {
			s.notify(s.conf.Slack.OAuthToken, s.conf.Slack.ShieldDecentralizeSolIssues, fmt.Sprintf(":scream: ShieldId `%v` mintToken `%v` mintStatus = `%v` (0) tx = `%v`", item.ID, item.SplToken, "Rejected", tx.Hash))
			// retry
			item.Status = models.TxRejectedSol
			if err := s.updateSuccess(item, s.dao); err != nil {
				return err
			}
			err := s.autoRetryMintToken(item)
			if err != nil {
				return errors.Wrap(err, "s.autoRetryMintToken")
			}
		}
		s.updateErrCount(item, s.dao)

		return nil
	}

	s.logger.Error(fmt.Sprintf("bridge status = %v, is not completed", bridgeStatus))
	return nil

}

func (s *CronMintToken) trackHistory(item *models.Shield, status models.HistoryStatus, requestMsg string, responseMsg string) {
	trackData := &models.ShieldHistory{
		JobId:         item.ID,
		JobStatus:     int(item.Status),
		JobStatusName: s.getStatusName(int(item.Status)),
		Status:        status,
		RequestMsg:    requestMsg,
		ResponseMsg:   responseMsg,
	}

	np := NewHistoryTask(trackData, s.dao)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		// Submit the task to be worked on. When RunTask
		// returns we know it is being handled.
		s.trackingHistory.Run(np)
		wg.Done()
	}()
	wg.Wait()

	if item.ErrCount >= MaxErr {
		s.notify(s.conf.Slack.OAuthToken, s.conf.Slack.ShieldDecentralizeSolIssues, fmt.Sprintf(":scream: ShieldId `%v` address = `%v` statusName = `%v` errorCount `%v`, please check manual", item.ID, item.Address, s.getStatusName(int(item.Status)), item.ErrCount))
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
