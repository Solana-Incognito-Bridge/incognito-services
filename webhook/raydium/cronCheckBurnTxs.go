package raydium

import (
	"encoding/json"
	"fmt"
	"strconv"

	go_incognito "github.com/inc-backend/go-incognito"
	config "github.com/orgs/Solana-Incognito-Bridge/ognito-service/conf"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/dao"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/models"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/serializers"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service/sol"

	"go.uber.org/zap"
)

type cronCheckBurnTxs struct {
	bc        *go_incognito.PublicIncognito
	wallet    *go_incognito.Wallet
	trans     *go_incognito.Trans
	solClient *sol.Client
	dao       *dao.Raydium
	conf      *config.Config
	logger    *zap.Logger
}

func NewCronCheckBurnTxs(bc *go_incognito.PublicIncognito, solClient *sol.Client, dao *dao.Raydium, conf *config.Config, logger *zap.Logger) *cronCheckBurnTxs {
	blockInfo := go_incognito.NewBlockInfo(bc)
	wallet := go_incognito.NewWallet(bc, blockInfo)
	trans := go_incognito.NewTrans(bc)
	return &cronCheckBurnTxs{bc: bc, wallet: wallet, trans: trans, solClient: solClient, dao: dao, conf: conf, logger: logger}
}

func (s *cronCheckBurnTxs) Start() {

	logs, _, _ := s.dao.List(
		0,
		99999,
		map[string]string{
			"trade_type":         fmt.Sprintf("%v", int(models.Raydium)),
			"status":             fmt.Sprintf("%v", int(models.ReceivedBurnTx)),
			"err_count.lessthan": strconv.Itoa(s.conf.PencakeSwapProtocol.RetryTimes),
		},
	)

	for _, log := range logs {
		s.process(log)
	}

}

func (s *cronCheckBurnTxs) process(log *models.Raydium) {
	if errMsg, err := s.checkIncognitoTxBurn(log); err != nil {
		go s.dao.TrackHistory(log, models.HistoryStatusFailure, fmt.Sprintf("cronCheckBurnTxs tx %v", log.BurnTx), fmt.Sprintf("%v:%v", err.Error(), errMsg))
		s.logger.Error("process", zap.Error(err))

		if log.ErrCount == models.MaxErr+10 {
			log.Status = models.TxBurnInvalid
		}
		log.ErrCount += 1
		s.dao.Update(log)
		return
	}

	go s.dao.TrackHistory(log, models.HistoryStatusSuccess, fmt.Sprintf("cronCheckBurnTxs tx %v", log.BurnTx), "Tx trade succeed!")

	log.Status = models.SendFeeSuccess //todo rollback burn success.
	log.ErrCount = 0
	s.dao.Update(log)
}

func (s *cronCheckBurnTxs) checkIncognitoTxFees(log *models.Raydium) (string, error) {
	// get tx info from chain:
	amount, err := s.trans.CheckOtaKeyAndUserSelection(log.BurnTx, go_incognito.PRVToken, s.conf.Incognito.MasterFeeOtaKey, s.conf.Incognito.TxReadOnlyKey)
	if err != nil {
		return err.Error(), service.TxBurnInvalid
	}

	// reset err count:
	log.UserFeeAmount = strconv.FormatUint(amount, 10)
	return s.checkUserSelectionPRV(log, amount)
}

func (s *cronCheckBurnTxs) checkIncognitoTxBurn(log *models.Raydium) (string, error) {
	tx, err := s.trans.GetTransactionDetailByTxHash(log.BurnTx)
	if err != nil {
		s.logger.Error(fmt.Sprintf("[trade-checkIncognitoTx] Id = %v GetTransactionDetailByTxHash", log.ID), zap.Error(err))
		return err.Error(), service.TxBurnInvalid
	}

	if tx == nil || !tx.IsInBlock {
		return "tx invalid", service.TxBurnInvalid
	}

	return "", nil
}

func (s *cronCheckBurnTxs) checkUserSelectionPRV(log *models.Raydium, amount uint64) (string, error) {
	if log.UserFeeSelection == models.ByPRV {
		var fees *serializers.Fees
		var prvFee uint64

		err := json.Unmarshal([]byte(log.OutsideChainPrivacyFee), &fees)
		if err != nil {
			return "Cannot unmarshal OutsideChainPrivacyFee data", service.ErrInternalServerError
		}

		if log.UserFeeLevel == models.One {
			prvFee, _ = strconv.ParseUint(fees.Level1, 10, 64)
		} else if log.UserFeeLevel == models.Two {
			prvFee, _ = strconv.ParseUint(fees.Level2, 10, 64)
		}

		if amount != prvFee {
			logs := fmt.Sprintf("Raydium: amount != prvFee: %v != %v", strconv.FormatUint(amount, 10), prvFee)
			log.Status = models.FeeAmountInvalid
			log.Logs += "|feeAmount:" + strconv.FormatUint(amount, 10)
			return logs, service.FeeAmountInvalid
		}
	} else {
		log.Status = models.TxBurnInvalid
		return logs, service.TxInvalid
	}

	return "", nil
}
