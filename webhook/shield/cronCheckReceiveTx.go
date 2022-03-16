package solana

import (
	"fmt"
	"sync"
	"time"

	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service/workerpool"

	config "github.com/orgs/Solana-Incognito-Bridge/ognito-service/conf"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/dao"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/models"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service/sol"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type cronCheckReceiveTx struct {
	Base
	solClient       *sol.Client
	dao             *dao.Shield
	conf            *config.Config
	logger          *zap.Logger
	trackingHistory *workerpool.Pool
}

func NewCronCheckReceiveTx(solClient *sol.Client, dao *dao.Shield, conf *config.Config, logger *zap.Logger, trackingHistory *workerpool.Pool) *cronCheckReceiveTx {
	return &cronCheckReceiveTx{solClient: solClient, dao: dao, conf: conf, logger: logger, trackingHistory: trackingHistory}
}

func (s *cronCheckReceiveTx) Start() {

	unprocessedLogs, err := s.dao.ListShieldByStatus([]models.ShieldStatus{
		models.TxReceiveNew,
	})

	if err != nil {
		s.logger.Warn("s.dao.ListShiedDecentralizedAddresses", zap.Error(err))
		return
	}
	s.run(unprocessedLogs)
}

func (s *cronCheckReceiveTx) run(logs []*models.Shield) {
	for _, log := range logs {
		if err := s.process(log); err != nil {
			s.logger.Error("process", zap.Error(err))
			s.trackHistory(log, models.HistoryStatusFailure, fmt.Sprintf("checkTxReceive shieldId %v", log.ID), fmt.Sprintf("%v", err.Error()))

			s.updateErrCount(log, s.dao)
		}
		time.Sleep(time.Second * 1)
	}
}

func (s *cronCheckReceiveTx) process(log *models.Shield) error {

	fmt.Println("cronCheckReceiveTx log ID->", log.ID)

	// check tx is success: or not
	ok, err := s.checkStatusOfTx(s.solClient, log.TxReceive)
	if err != nil {
		return err
	}

	if ok {
		// set filed for update:
		log.Status = models.EstimatedFee
		if err = s.updateSuccess(log, s.dao); err != nil {
			return err
		}
		return nil
	}

	return errors.New("Tx not confirm yet!!!")
}

func (s *cronCheckReceiveTx) trackHistory(item *models.Shield, status models.HistoryStatus, requestMsg string, responseMsg string) {
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
