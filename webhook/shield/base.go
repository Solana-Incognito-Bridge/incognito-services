package solana

import (
	"fmt"

	"github.com/inc-backend/3rd-libs/3rd/slack"
	config "github.com/orgs/Solana-Incognito-Bridge/ognito-service/conf"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/dao"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/models"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service/sol"
	"go.uber.org/zap"
)

const (
	MaxErr = 122
)

type Base struct {
}

func (b Base) notifyShieldDecentalized(block uint64, message string, conf *config.Config) {
	b.notify(conf.Slack.OAuthToken, conf.Slack.ShieldDecentralizeSolIssues, fmt.Sprintf("Error from get tx solana: BlockHeight `%v`, Error: `%v`", block, message))
}

func (b Base) notify(token string, channel string, msg string) {
	slk := slack.InitSlack(token, channel)
	go slk.PostMsg(msg)
}

func (b Base) getStatusName(status int) string {
	if v, ok := models.ShieldStatusName[status]; ok {
		return v
	}

	return ""
}
func (b Base) checkStatusOfTx(solClient *sol.Client, tx string) (bool, error) {
	return solClient.CheckTxStatus(tx)
}

func (b Base) updateErrCount(item *models.Shield, dao *dao.Shield) {
	if err := dao.IncreaseShieldErrCount(item); err != nil {
		fmt.Println("IncreaseShieldErrCount", zap.Error(err))
	}
}

func (b Base) updateSuccess(item *models.Shield, dao *dao.Shield) error {
	item.ErrCount = 0
	if err := dao.UpdateAddress(item); err != nil {
		fmt.Println("UpdateAddress", zap.Error(err))
		return err
	}

	return nil
}

func (b Base) getGasPrice() uint64 {
	return 5000
}
