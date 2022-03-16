package raydium

import (
	"fmt"

	"github.com/inc-backend/3rd-libs/3rd/slack"
	config "github.com/orgs/Solana-Incognito-Bridge/ognito-service/conf"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/dao"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/models"
)

func NotifyJobError(conf *config.Config, logDAO *dao.Raydium) error {
	processList, err := logDAO.GetListTimeout()

	if err != nil {
		return err
	}

	if len(processList) <= 0 {
		return nil
	}

	for _, value := range processList {

		statusStr := models.ConvertRaydiumToString(value.Status)

		msgForSlack := fmt.Sprintf(":scream: Uniswap RequestId `%v` error at status `%v (%v)`", value.ID, statusStr, value.Status)
		s := slack.InitSlack(conf.Slack.OAuthToken, conf.Slack.PencakeSwapProtocol)
		go s.PostMsg(msgForSlack)
	}

	return nil
}
