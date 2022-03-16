package raydium

import (
	"fmt"
	"strconv"
	"time"

	go_incognito "github.com/inc-backend/go-incognito"
	config "github.com/orgs/Solana-Incognito-Bridge/ognito-service/conf"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/dao"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/models"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service/sol"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	GetAndSubmitProofDuration   = time.Minute * 90
	GetWithdrawTxStatusDuration = time.Minute * 30
)

type cronSubmitProof struct {
	bc        *go_incognito.PublicIncognito
	wallet    *go_incognito.Wallet
	trans     *go_incognito.Trans
	solClient *sol.Client
	dao       *dao.Raydium
	conf      *config.Config
	logger    *zap.Logger
}

func NewCronSubmitProof(bc *go_incognito.PublicIncognito, solClient *sol.Client, dao *dao.Raydium, conf *config.Config, logger *zap.Logger) *cronSubmitProof {
	blockInfo := go_incognito.NewBlockInfo(bc)
	wallet := go_incognito.NewWallet(bc, blockInfo)
	trans := go_incognito.NewTrans(bc)
	return &cronSubmitProof{bc: bc, wallet: wallet, trans: trans, solClient: solClient, dao: dao, conf: conf, logger: logger}
}

func (s *cronSubmitProof) Start() {
	logs, _, _ := s.dao.List(0, 99999,
		map[string]string{
			"trade_type": fmt.Sprintf("%v", int(models.Raydium)),
			"status.in": fmt.Sprintf("%v, %v",
				int(models.SendFeeSuccess),
				int(models.SubmitedProofToSC)),
			"err_count.lessthan": strconv.Itoa(s.conf.PencakeSwapProtocol.RetryTimes),
		},
	)

	for _, log := range logs {
		s.process(log)
	}
}

func (s *cronSubmitProof) process(log *models.Raydium) {
	if log.Status == models.SendFeeSuccess {
		s.processSubmitProof(log)
	} else if log.Status == models.SubmitedProofToSC {
		s.checkTxSubmmitPrrof(log)
	}
}

func (s *cronSubmitProof) processSubmitProof(log *models.Raydium) {

	signersSwapPrivateKey := s.conf.RaydiumProtocol.MasterPrivateKey
	privKeyFee := s.conf.Solana.MasterShieldPrivateKey

	txHash, err := s.solClient.SubmitProofTx(privKeyFee, signersSwapPrivateKey, log.BurnTx, log.SrcContractAddress)

	if err != nil {
		err = errors.Wrap(err, "SubmitBurnProof")
		go s.dao.TrackHistory(log, models.HistoryStatusFailure, fmt.Sprintf("cronSubmitProof tradeId %v", log.ID), err.Error())

		log.ErrCount += 1
		if log.ErrCount == models.MaxErr+10 {
			log.Status = models.SubmitedFroofToSCFailed
		}
		s.dao.Update(log)
		return
	}

	log.ErrCount = 0
	log.SubmitProofTx = txHash

	fmt.Printf("SubmitBurnProof, txHash: %x\n", log.SubmitProofTx)
	go s.dao.TrackHistory(log, models.HistoryStatusSuccess, fmt.Sprintf("cronSubmitProof tx %v", log.SubmitProofTx), "Submited proof trade!")

	log.Status = models.SubmitedProofToSC
	log.ErrCount = 0
	s.dao.Update(log)

}

func (s *cronSubmitProof) checkTxSubmmitPrrof(log *models.Raydium) {

	ok, err := s.solClient.CheckTxStatus(log.SubmitProofTx)

	if err != nil {
		//error
		go s.dao.TrackHistory(log, models.HistoryStatusFailure, fmt.Sprintf("CheckTxStatus tx %v", log.SubmitProofTx), err.Error())
		log.ErrCount += 1
		if log.ErrCount == models.MaxErr+10 {
			log.Status = models.SubmitedFroofToSCFailed
		}
		s.dao.Update(log)

		return
	}

	if ok {
		go s.dao.TrackHistory(log, models.HistoryStatusSuccess, fmt.Sprintf("cronSubmitProof tradeId %v", log.ID), "Tx submit proof succeed!")

		// pass:
		log.Status = models.SubmitedProofToSCSucceed
		log.ErrCount = 0
		s.dao.Update(log)
		return
	} else {
		go s.dao.TrackHistory(log, models.HistoryStatusSuccess, fmt.Sprintf("cronSwapToken tradeId %v", log.ID), "Check tx submit false!")
	}

}
