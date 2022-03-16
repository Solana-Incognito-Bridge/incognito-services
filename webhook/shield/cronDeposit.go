package solana

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/inc-backend/crypto-libs/eth/bridge"
	sdk "github.com/inc-backend/sdk/encryption"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service/sol"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service/workerpool"

	config "github.com/orgs/Solana-Incognito-Bridge/ognito-service/conf"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/dao"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type cronDeposit struct {
	Base
	solClient       *sol.Client
	dao             *dao.Shield
	conf            *config.Config
	logger          *zap.Logger
	bridge          *bridge.Bridge
	trackingHistory *workerpool.Pool
	privateKey      string
}

func NewCronDeposit(solClient *sol.Client, dao *dao.Shield, conf *config.Config, logger *zap.Logger, bridge *bridge.Bridge, trackingHistory *workerpool.Pool) *cronDeposit {
	privateKey, _ := sdk.DecryptToString(conf.Solana.MasterShieldPrivateKey, conf.Ethereum.KeyDecrypt)
	if len(privateKey) == 0 {
		privateKey = conf.Solana.MasterShieldPrivateKey
	}
	return &cronDeposit{solClient: solClient, dao: dao, conf: conf, logger: logger, bridge: bridge, trackingHistory: trackingHistory, privateKey: privateKey}
}

func (s *cronDeposit) Start() {

	unprocessedLogs, err := s.dao.ListShieldByStatus([]models.ShieldStatus{
		models.ReceivedFeeSol,
		models.ApprovedSol,
		models.RequestedApproveSol,
		models.RequestedDepositSol,
	})

	if err != nil {
		s.logger.Warn("s.dao.ListShieldByStatus", zap.Error(err))
		return
	}

	s.run(unprocessedLogs)
}

func (s *cronDeposit) run(logs []*models.Shield) {
	for _, log := range logs {
		if err := s.process(log); err != nil {
			s.logger.Error("process", zap.Error(err))
			s.trackHistory(log, models.HistoryStatusFailure, fmt.Sprintf("deposit shieldId %v", log.ID), fmt.Sprintf("%v", err.Error()))
			s.updateErrCount(log, s.dao)
		}
		time.Sleep(time.Second * 1)
	}
}

func (s *cronDeposit) process(log *models.Shield) error {
	fmt.Println("cronDeposit log ID->", log.ID)

	// get token info:
	ptoken, _ := s.dao.GetPTokenByID(log.IncognitoToken)
	if ptoken == nil {
		return service.MissTokenAddress
	}
	// check vault token first:
	vaultAddressModel, _ := s.dao.GetVaultTokenByTokenID(log.SplToken)
	if vaultAddressModel == nil {
		// create vault:
		vaultAddress, txHash, err := s.solClient.CreateVaultAddress(s.privateKey, log.SplToken)
		if err != nil {
			fmt.Println("can not create vault address! token: ", log.SplToken)
			return service.MissTokenAddress
		}
		// insert vault:
		vaultAddressModel = &models.VaultSolToken{
			SplToken:     log.SplToken,
			VaultAddress: vaultAddress,
			Tx:           txHash,
		}
		if err = s.dao.CreateVaultAddress(vaultAddressModel); err != nil {
			return err
		}
		return nil
	}

	privKey := ""

	shieldWallet, _ := s.dao.GetShieldWalletByAddress(log.Address)
	if shieldWallet == nil {
		return service.MissTokenAddress
	}

	if log.Address == "0x377B27a1f354dD0141fCF73B6dd307178B03f346" {
		privKey = shieldWallet.PrivKey
	} else {
		privKey, _ = sdk.DecryptToString(shieldWallet.PrivKey, s.conf.Ethereum.KeyDecrypt)
		if privKey == "" {
			privKey = shieldWallet.PrivKey
		}
	}
	tokenAccountAddress := shieldWallet.Address

	switch log.Status {

	case models.ReceivedFeeSol:
		// will call approve/ deposit:
		return s.deposit(log, ptoken, privKey, tokenAccountAddress, vaultAddressModel.VaultAddress)
	case models.RequestedDeposit:
		// check status deposit
		return s.checkDeposit(log)
	}

	return nil
}

//deposit:
func (s *cronDeposit) deposit(log *models.Shield, ptoken *models.PToken, privKey, shieldTokenAddress, vaultAddress string) error {

	// check pending approve/deposit:
	pendingApproveDeposit, _ := s.dao.ListShieldByStatusAddress([]models.ShieldStatus{
		models.RequestedDeposit,
	}, log.Address)

	if len(pendingApproveDeposit) > 0 {
		return nil
	}

	shieldAmount, err := strconv.ParseUint(log.ShieldAmount, 10, 64)
	if err != nil {
		return errors.New("Cant not parser ShieldAmount")
	}
	tx := ""
	if log.CurrencyType == models.SOL {
		tx, err = s.solClient.ShieldNative(privKey, log.IncAddress, vaultAddress, shieldAmount)
	} else {
		tx, err = s.solClient.ShieldToken(privKey, log.IncAddress, shieldTokenAddress, vaultAddress, shieldAmount)
	}

	if err != nil {
		s.notify(s.conf.Slack.OAuthToken, s.conf.Slack.ShieldDecentralizeSolIssues,
			fmt.Sprintf(":scream: ShieldId `%v` address = `%v`\n ```depositETHError: %v```",
				log.ID,
				log.Address,
				ptoken.Symbol,
				err.Error(),
			))
		s.trackHistory(log, models.HistoryStatusFailure, fmt.Sprintf("deposit => receiveAmount = %v, shieldAmount = %v, fee = %v", log.ReceivedAmount, log.ShieldAmount, log.Fee), fmt.Sprintf("%v", err.Error()))
		return err
	}

	fmt.Println("Deposit sol successfully with tx = ", tx)
	log.TxDeposit = tx

	s.trackHistory(log, models.HistoryStatusSuccess, fmt.Sprintf("deposit amount %v successfully", shieldAmount), tx)

	log.Status = models.RequestedDepositSol
	if err = s.updateSuccess(log, s.dao); err != nil {
		return err
	}

	return nil
}

func (s *cronDeposit) checkDeposit(log *models.Shield) error {
	// check tx is success: or not
	ok, err := s.checkStatusOfTx(s.solClient, log.TxDeposit)
	if err != nil {
		return err
	}

	if ok {
		// set filed for update:
		log.Status = models.DepositedSol
		if err = s.updateSuccess(log, s.dao); err != nil {
			return err
		}
	}

	return nil
}

func (s *cronDeposit) trackHistory(item *models.Shield, status models.HistoryStatus, requestMsg string, responseMsg string) {
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
