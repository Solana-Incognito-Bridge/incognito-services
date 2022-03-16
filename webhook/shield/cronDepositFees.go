package solana

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	go_incognito "github.com/inc-backend/go-incognito"
	sdk "github.com/inc-backend/sdk/encryption"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service/sol"

	shieldCommon "github.com/orgs/Solana-Incognito-Bridge/ognito-service/common"
	config "github.com/orgs/Solana-Incognito-Bridge/ognito-service/conf"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/dao"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/models"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service/workerpool"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type CronDepositFees struct {
	Base
	solClient       *sol.Client
	logger          *zap.Logger
	conf            *config.Config
	dao             *dao.Shield
	trackingHistory *workerpool.Pool
	privKey         string
	bc              *go_incognito.PublicIncognito
	wallet          *go_incognito.Wallet
}

func NewCronDepositFees(solClient *sol.Client, logger *zap.Logger, conf *config.Config, dao *dao.Shield, trackingHistory *workerpool.Pool, bc *go_incognito.PublicIncognito) *CronDepositFees {
	blockInfo := go_incognito.NewBlockInfo(bc)
	wallet := go_incognito.NewWallet(bc, blockInfo)

	masterPrivateKey, _ := sdk.DecryptToString(conf.Solana.MasterShieldPrivateKey, conf.Ethereum.KeyDecrypt)
	if len(masterPrivateKey) == 0 || conf.Env != "production" {
		masterPrivateKey = conf.Solana.MasterShieldPrivateKey
	}

	return &CronDepositFees{solClient: solClient, logger: logger, conf: conf, dao: dao, trackingHistory: trackingHistory, privKey: masterPrivateKey, bc: bc, wallet: wallet}
}

func (s *CronDepositFees) Start() {

	listStatus := []models.ShieldStatus{
		models.TxReceiveSuccess,
		models.EstimatedFee,
		models.SendingFee,
	}

	listShieldItem, err := s.dao.ListShieldByStatus(listStatus)
	if err != nil {
		s.logger.Error("s.dao.ListShieldByStatus", zap.Error(err))
		return
	}

	for _, item := range listShieldItem {
		s.logger.Info(fmt.Sprintf("Start shield deposit #%v", item.ID))

		shieldWallet, err := s.dao.GetShieldWalletByAddress(item.Address)
		if err != nil {
			s.logger.Error("s.dao.GetShieldWalletById", zap.Error(err))

			s.updateErrCount(item, s.dao)
			s.trackHistory(item, models.HistoryStatusFailure, fmt.Sprintf("get wallet %v", item.Address), fmt.Sprintf("%v", err.Error()))

			continue
		}

		var token *models.PToken

		if item.SplToken == shieldCommon.SolAddrStr {
			token, _ = s.dao.GetSOLToken()
		} else {
			token, _ = s.dao.GetPTokenByContractID(item.SplToken)
		}

		if token == nil {
			item.Status = models.InvalidInfoSol
			s.updateSuccess(item, s.dao)
			s.updateErrCount(item, s.dao)

			s.trackHistory(item, models.HistoryStatusFailure, fmt.Sprintf("get token %v is null", item.SplToken), "")

			continue
		}

		var errs error
		switch item.Status {
		case models.TxReceiveSuccess:
			//merge amount pre tx by status = invalid fees
			//shield amount
			//merge pre tx
			//fees amount
			// errs = s.estimatedFee(item, shieldWallet, token)
		case models.EstimatedFeeBsc:
			//send fees
			s.sendFeesTmpWallet(item, shieldWallet, token)
		case models.SendingFee:
			errs = s.checkTxStatus(item)
		}

		if errs != nil {
			s.trackHistory(item, models.HistoryStatusFailure, fmt.Sprintf("depositFee shieldId %v", item.ID), fmt.Sprintf("%v", errs.Error()))
			s.logger.Error(fmt.Sprintf("shieldId %v depositFeeToken %v", item.ID, item.SplToken), zap.Error(errs))

			s.updateErrCount(item, s.dao)
		}

		time.Sleep(time.Second * 1)
	}
}

func (s *CronDepositFees) sendFeesTmpWallet(item *models.Shield, shieldWallet *models.ShieldSolWallet, shieldToken *models.PToken) (bool, error) {

	shieldAmount, err := strconv.ParseUint(item.ShieldAmount, 10, 64)

	if err != nil {
		return false, errors.New("Can not parser ShieldAmount")
	}

	if shieldAmount == 0 {
		s.trackHistory(item, models.HistoryStatusFailure, "Shield amount invalid", "0")

		item.Status = models.InvalidFeeSol
		if err := s.updateSuccess(item, s.dao); err != nil {
			return false, err
		}

		return false, nil
	}

	totalTxFees, err := strconv.ParseUint(item.Fee, 10, 64)
	if err != nil {
		return false, errors.New("Can not parser Fee")
	}

	totalChargeFee, err := strconv.ParseUint(item.ChargeFee, 10, 64)

	if err != nil {
		return false, errors.New("Can not parser ChargeFee")
	}

	//sol
	if shieldToken.CurrencyType == models.SOL {
		totalTxFees = totalTxFees - totalChargeFee
	}

	//send tx
	// hard code fee now:
	totalChargeFee = 50000 * 4

	txId, err := s.sendFees(shieldWallet, totalChargeFee)
	if err != nil {
		//notify important
		s.notify(s.conf.Slack.OAuthToken, s.conf.Slack.ShieldDecentralizeSolIssues,
			fmt.Sprintf(":scream: ShieldId `%v` address = `%v`\n gasFee = `%.8f Sol`\n tranferAmount = `%f SOL`  ```depostiFeeError: %v```",
				item.ID, item.Address,
				totalTxFees/1e9,
				totalChargeFee/1e9,
				err.Error(),
			))

		return false, errors.Wrap(err, "s.sendFees")
	}

	if len(txId) <= 0 {
		return false, errors.New("Txid is empty")
	}

	s.trackHistory(item, models.HistoryStatusSuccess, fmt.Sprintf("depositFee amount = %v to address = %v ", totalTxFees, shieldWallet.Address), txId)

	item.Status = models.SendingFeeSol
	item.TxFee = txId

	if err := s.updateSuccess(item, s.dao); err != nil {
		return false, err
	}

	return true, nil
}

func (s *CronDepositFees) checkTxStatus(item *models.Shield) error {
	status, err := s.checkStatusOfTx(s.solClient, item.TxFee)
	if err != nil {
		return errors.Wrap(err, "s.checkStatusOfTx")
	}

	if !status {
		return errors.New(fmt.Sprintf("txId %v is not successfully", item.TxFee))
	}

	item.Status = models.ReceivedFeeSol
	if err := s.updateSuccess(item, s.dao); err != nil {
		return err
	}
	return nil
}

func (s *CronDepositFees) calculateFees(shieldWalletAddress string, tokenShieldId string) (uint64, error) {

	if tokenShieldId == shieldCommon.SolAddrStr {

	}
	return s.getGasPrice() * 2, nil

}

func (s *CronDepositFees) sendFees(wallet *models.ShieldSolWallet, feesAmount uint64) (string, error) {

	tx, err := s.solClient.Transfer(s.privKey, wallet.SolAddress, feesAmount)
	if err != nil {
		return "", errors.Wrap(err, "solClient.Transfer")
	}

	return tx, nil
}

func (s *CronDepositFees) trackHistory(item *models.Shield, status models.HistoryStatus, requestMsg string, responseMsg string) {
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
