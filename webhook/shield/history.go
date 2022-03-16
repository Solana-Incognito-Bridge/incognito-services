package solana

import (
	"fmt"

	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/dao"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/models"

	"github.com/pkg/errors"
)

type HistoryTask struct {
	model *models.ShieldHistory
	dao   *dao.Shield
}

func NewHistoryTask(model *models.ShieldHistory, dao *dao.Shield) *HistoryTask {
	return &HistoryTask{model: model, dao: dao}
}

func (s *HistoryTask) Task() {
	if err := s.dao.CreateNewShieldHistory(s.model); err != nil {
		err = errors.WithStack(err)
		fmt.Printf("Error updating log: id %v err %v", s.model.ID, err)
	}

}

func (s *HistoryTask) TrackHistory(status int, message string, responseData string) {
}
