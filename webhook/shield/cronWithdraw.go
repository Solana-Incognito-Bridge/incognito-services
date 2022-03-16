package solana

import (
	"fmt"
	"sync"
	"time"

	go_incognito "github.com/inc-backend/go-incognito"
	"github.com/pkg/errors"

	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/models"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service/sol"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service/workerpool"

	sdk "github.com/inc-backend/sdk/encryption"

	config "github.com/orgs/Solana-Incognito-Bridge/ognito-service/conf"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/dao"
	"go.uber.org/zap"
)

var (
	GetAndSubmitProofDuration   = time.Minute * 90
	GetWithdrawTxStatusDuration = time.Minute * 30
)

type TxsStatus int

const (
	Pending TxsStatus = iota
	NeedRetry
	Passed
	Failed
)

type cronWithdraw struct {
	Base
	solClient *sol.Client
	// bridge          *bridge.Bridge
	dao             *dao.Shield
	notificationDao *dao.NotificationDAO
	bc              *go_incognito.PublicIncognito
	wallet          *go_incognito.Wallet
	trans           *go_incognito.Trans
	conf            *config.Config
	logger          *zap.Logger
	trackingHistory *workerpool.Pool
	privKey         string
}

func NewCronWithdraw(solClient *sol.Client,
	// bridge *bridge.Bridge,
	dao *dao.Shield, notificationDao *dao.NotificationDAO,
	bc *go_incognito.PublicIncognito, conf *config.Config, logger *zap.Logger, trackingHistory *workerpool.Pool) *cronWithdraw {
	blockInfo := go_incognito.NewBlockInfo(bc)
	wallet := go_incognito.NewWallet(bc, blockInfo)
	trans := go_incognito.NewTrans(bc)

	privKey, _ := sdk.DecryptToString(conf.Solana.MasterShieldPrivateKey, conf.Ethereum.KeyDecrypt)

	if len(privKey) == 0 || conf.Env != "production" {
		privKey = conf.Solana.MasterShieldPrivateKey
	}

	return &cronWithdraw{solClient: solClient,
		// bridge: bridge,
		dao: dao, bc: bc, wallet: wallet, trans: trans, notificationDao: notificationDao, conf: conf,
		logger: logger, trackingHistory: trackingHistory, privKey: privKey}
}

func (s *cronWithdraw) Start() {

	listStatus := []models.ShieldStatus{
		models.ReleasingTokenSol,
		models.ReceivedWithdrawTxSol,
	}

	listUnShieldItem, err := s.dao.ListShieldByStatus(listStatus)
	if err != nil {
		s.logger.Error("s.dao.ListShieldByStatus", zap.Error(err))
		return
	}

	s.run(listUnShieldItem) // Check tx
}

func (s *cronWithdraw) run(logs []*models.Shield) {
	for _, log := range logs {
		if err := s.process(log); err != nil {
			s.logger.Error("process", zap.Error(err))
			s.trackHistory(log, models.HistoryStatusFailure, fmt.Sprintf("unshield shieldId %v", log.ID), fmt.Sprintf("%v", err.Error()))
			s.updateErrCount(log, s.dao)
		}
		time.Sleep(time.Second * 1)
	}
}

func (s *cronWithdraw) process(log *models.Shield) error {
	fmt.Println("cronDeposit log ID->", log.ID)

	if log.Status == models.ReleasingTokenSol {
		return s.checkTxStatus(log)
	}
	//ReceivedWithdrawTxSol
	return s.unshield(log)
}

func (s *cronWithdraw) unshield(log *models.Shield) error {
	// get token info:
	ptoken, _ := s.dao.GetPTokenByID(log.IncognitoToken)
	if ptoken == nil {
		return service.MissTokenAddress
	}

	var tx string
	var err error

	if ptoken.CurrencyType == models.SOL {
		tx, err = s.solClient.UnShieldNative(s.privKey, log.TxMintBurnToken, log.UserPaymentAddress)
	} else {
		tx, err = s.solClient.UnShieldToken(s.privKey, log.TxMintBurnToken, log.SplToken, log.UserPaymentAddress)
	}
	if err == nil {
		log.Status = models.ReleasingTokenSol
		log.TxWithdraw = tx
		if err := s.updateSuccess(log, s.dao); err != nil {
			return err
		}
	}

	return err
}

func (s *cronWithdraw) checkTxStatus(item *models.Shield) error {
	status, err := s.checkStatusOfTx(s.solClient, item.TxWithdraw)
	if err != nil {
		return errors.Wrap(err, "s.checkStatusOfTx")
	}

	if !status {
		return errors.New(fmt.Sprintf("txId %v is not successfully", item.TxFee))
	}

	item.Status = models.ReleaseTokenSucceedSol
	if err := s.updateSuccess(item, s.dao); err != nil {
		return err
	}
	go service.NotifySolana(item, s.conf, s.dao, s.notificationDao)

	return nil
}

func (s *cronWithdraw) trackHistory(item *models.Shield, status models.HistoryStatus, requestMsg string, responseMsg string) {
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
		s.notify(s.conf.Slack.OAuthToken, s.conf.Slack.ShieldDecentralizeSolIssues, fmt.Sprintf(":scream: UnShieldId `%v` address = `%v` statusName = `%v` errorCount `%v`, please check manual", item.ID, item.Address, s.getStatusName(int(item.Status)), item.ErrCount))
	}
}
