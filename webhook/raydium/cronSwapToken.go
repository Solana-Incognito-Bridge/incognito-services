package raydium

import (
	"fmt"
	"strconv"

	go_incognito "github.com/inc-backend/go-incognito"
	config "github.com/orgs/Solana-Incognito-Bridge/ognito-service/conf"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/dao"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/models"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service/sol"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	PLG_EXECUTE_PREFIX      = 4
	PLG_REQ_WITHDRAW_PREFIX = 5
)

const (
	LOW    = 500
	MEDIUM = 3000
	HIGH   = 10000
)

const MAX_PERCENT = 10000

type cronSwapToken struct {
	bc        *go_incognito.PublicIncognito
	wallet    *go_incognito.Wallet
	trans     *go_incognito.Trans
	solClient *sol.Client
	dao       *dao.Raydium
	conf      *config.Config
	logger    *zap.Logger
}

func NewCronSwapToken(bc *go_incognito.PublicIncognito, solClient *sol.Client, dao *dao.Raydium, conf *config.Config, logger *zap.Logger) *cronSwapToken {
	blockInfo := go_incognito.NewBlockInfo(bc)
	wallet := go_incognito.NewWallet(bc, blockInfo)
	trans := go_incognito.NewTrans(bc)

	return &cronSwapToken{bc: bc, wallet: wallet, trans: trans, solClient: solClient, dao: dao, conf: conf, logger: logger}
}

func (s *cronSwapToken) Start() {
	logs, _, _ := s.dao.List(0, 99999,
		map[string]string{
			"trade_type": fmt.Sprintf("%v", int(models.Raydium)),
			"status.in": fmt.Sprintf("%v, %v",
				int(models.SubmitedProofToSCSucceed),
				int(models.SwapedCoin)),
			"err_count.lessthan": strconv.Itoa(s.conf.PencakeSwapProtocol.RetryTimes),
		},
	)
	for _, log := range logs {
		s.process(log)
	}
}

func (s *cronSwapToken) process(log *models.Raydium) {
	if log.Status == models.SubmitedProofToSCSucceed {
		s.processSwap(log)
	} else if log.Status == models.SwapedCoin {
		s.checkTxSwap(log)
	}
}

func (s *cronSwapToken) processSwap(log *models.Raydium) {

	signersSwapPrivateKey := s.conf.RaydiumProtocol.MasterPrivateKey
	privKeyFee := s.conf.Solana.MasterShieldPrivateKey

	amountIn, _ := strconv.ParseUint(log.SrcQties, 10, 64)
	amountOut, _ := strconv.ParseUint(log.ExpectedOutputAmount, 10, 64)

	// begin swap now ==========================================================================================================================
	txHash, err := s.solClient.Swap(privKeyFee, signersSwapPrivateKey, log.SrcContractAddress, log.DestContractAddress, amountIn, amountOut)

	if err != nil {
		fmt.Println("Raydium trade", err.Error())
		log.ErrCount += 1
		s.dao.Update(log)

		err = errors.Wrap(err, "c.Swap()")
		go s.dao.TrackHistory(log, models.HistoryStatusFailure, fmt.Sprintf("cronSwapToken tradeId %v", log.ID), err.Error())
		return
	}

	log.ErrCount = 0
	log.ExecuteSwapTx = txHash

	fmt.Printf("pRaydium trade executed tx: %v\n", log.ExecuteSwapTx)
	go s.dao.TrackHistory(log, models.HistoryStatusSuccess, fmt.Sprintf("cronSwapToken tradeId %v tx = %v", log.ID, log.ExecuteSwapTx), "pRaydium trade executed!")

	log.Status = models.SwapedCoin
	log.ErrCount = 0
	s.dao.Update(log)
}

func (s *cronSwapToken) checkTxSwap(log *models.Raydium) {

	ok, err := s.solClient.CheckTxStatus(log.ExecuteSwapTx)

	if err != nil {
		//error
		go s.dao.TrackHistory(log, models.HistoryStatusFailure, fmt.Sprintf("TransactionReceipt tx %v", log.ExecuteSwapTx), err.Error())
		log.ErrCount += 1
		if log.ErrCount == models.MaxErr+10 {
			log.Status = models.SwapedCoinFailed
		}

		s.dao.Update(log)

		return
	}

	if ok {
		go s.dao.TrackHistory(log, models.HistoryStatusSuccess, fmt.Sprintf("cronSwapToken tradeId %v", log.ID), "Tx swap trade succeed!")

		// pass:
		log.Status = models.SwapedCoinSucceed
		log.ErrCount = 0
		s.dao.Update(log)
		return
	} else {
		go s.dao.TrackHistory(log, models.HistoryStatusSuccess, fmt.Sprintf("cronSwapToken tradeId %v", log.ID), "Check tx false!")
	}
}
