package raydium

import (
	"fmt"
	"strconv"

	"github.com/inc-backend/crypto-libs/helper"

	go_incognito "github.com/inc-backend/go-incognito"
	config "github.com/orgs/Solana-Incognito-Bridge/ognito-service/conf"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/dao"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/models"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service/sol"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type cronWithdraw struct {
	bc        *go_incognito.PublicIncognito
	wallet    *go_incognito.Wallet
	trans     *go_incognito.Trans
	solClient *sol.Client
	dao       *dao.Raydium
	conf      *config.Config
	logger    *zap.Logger
}

func NewcronWithdraw(bc *go_incognito.PublicIncognito, solClient *sol.Client, dao *dao.Raydium, conf *config.Config, logger *zap.Logger) *cronWithdraw {
	blockInfo := go_incognito.NewBlockInfo(bc)
	wallet := go_incognito.NewWallet(bc, blockInfo)
	trans := go_incognito.NewTrans(bc)

	return &cronWithdraw{bc: bc, wallet: wallet, trans: trans, solClient: solClient, dao: dao, conf: conf, logger: logger}

}

func (s *cronWithdraw) Start() {
	logs, _, _ := s.dao.List(0, 99999,
		map[string]string{
			"trade_type": fmt.Sprintf("%v", int(models.Raydium)),
			"status.in": fmt.Sprintf("%v, %v, %v, %v",
				int(models.SwapedCoinSucceed),
				int(models.RequestedWithdraw),
				int(models.NeedToRefundRequest),
				int(models.RefundRequestedWithdraw),
			),
			"err_count.lessthan": strconv.Itoa(s.conf.PencakeSwapProtocol.RetryTimes),
		},
	)
	for _, log := range logs {
		s.process(log)
	}
}

func (s *cronWithdraw) process(log *models.Raydium) {
	if log.Status == models.SwapedCoinSucceed {
		s.processWithdrawBuyToken(log)
	} else if log.Status == models.NeedToRefundRequest {

	} else if log.Status == models.RequestedWithdraw {
		s.checkTx(log)
	} else if log.Status == models.RefundRequestedWithdraw {
		// s.checkTxRefund(log)
	}
}

func (s *cronWithdraw) processWithdrawBuyToken(log *models.Raydium) {

	buyToken, _ := s.dao.GetPTokenByID(log.DestTokens)

	signersSwapPrivateKey := s.conf.RaydiumProtocol.MasterPrivateKey
	privKeyFee := s.conf.Solana.MasterShieldPrivateKey

	amountOut, _ := strconv.ParseUint(log.ExpectedOutputAmount, 10, 64)

	// begin swap now ==========================================================================================================================
	txHash, err := s.solClient.WithdrawRaydium(privKeyFee, signersSwapPrivateKey, log.WalletAddress, log.DestContractAddress, amountOut)

	if err != nil {
		log.ErrCount += 1
		s.dao.Update(log)
		err = errors.Wrap(err, "c.RequestWithdraw()")
		go s.dao.TrackHistory(log, models.HistoryStatusFailure, fmt.Sprintf("cronWithdraw tradeId %v", log.ID), err.Error())
		return
	}

	fmt.Printf("uniswap trade RequestedWithdraw tx: %v\n", log.WithdrawTx)
	go s.dao.TrackHistory(log, models.HistoryStatusSuccess, fmt.Sprintf("cronWithdraw tradeId %v tx = %v", log.ID, txHash), "uniswap trade RequestedWithdraw!")

	balanceInChain := helper.ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(log.ExpectedOutputAmount, buyToken.Decimals, buyToken.PDecimals)

	log.ErrCount = 0
	log.WithdrawTx = txHash
	log.Status = models.RequestedWithdraw
	log.ErrCount = 0
	log.OutputAmount = balanceInChain
	s.dao.Update(log)

}

func (s *cronWithdraw) checkTx(log *models.Raydium) {

	ok, err := s.solClient.CheckTxStatus(log.SubmitProofTx)

	if err != nil {
		//error
		go s.dao.TrackHistory(log, models.HistoryStatusFailure, fmt.Sprintf("CheckTxStatus tx %v", log.WithdrawTx), err.Error())
		log.ErrCount += 1
		if log.ErrCount == models.MaxErr+10 {
			log.Status = models.WithdrawFailed
		}
		s.dao.Update(log)

		return
	}

	if ok {
		go s.dao.TrackHistory(log, models.HistoryStatusSuccess, fmt.Sprintf("cronSubmitProof tradeId %v", log.ID), "Tx submit proof succeed!")

		// pass:
		log.Status = models.WithdrawSuccess
		log.ErrCount = 0
		s.dao.Update(log)
		return
	} else {
		go s.dao.TrackHistory(log, models.HistoryStatusSuccess, fmt.Sprintf("cronSwapToken tradeId %v", log.ID), "Check tx submit false!")
	}

}
