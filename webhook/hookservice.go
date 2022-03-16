package webhook

import (
	"net/http"

	config "github.com/incognito-services/conf"
	"github.com/incognito-services/constants"
	"github.com/incognito-services/dao"
	"github.com/incognito-services/serializers"
	"github.com/incognito-services/webhook/stakingpool"
	"go.uber.org/zap"
)

type HookService struct {
	notificationDAO *dao.NotificationDAO
	bc              *incognito.Blockchain
	conf            *config.Config
	logger          *zap.Logger
}

func NewHookService(notificationDAO *dao.NotificationDAO, bc *incognito.Blockchain, conf *config.Config, logger *zap.Logger) *HookService {
	return &HookService{notificationDAO: notificationDAO, bc: bc, conf: conf, logger: logger}
}

func (o *HookService) HookStakingPoolEvent(stepCase string) *serializers.EventHookResult {
	switch stepCase {
	case constants.STEP_PROCESS_CHECK_STAKING:
		step := stakingpool.NewStakeCron(
			o.notificationDAO,
			o.bc,
			o.conf,
			o.logger.With(zap.String("HookStakingPoolEvent", constants.STEP_PROCESS_CHECK_STAKING)),
		)
		step.Start()

	case constants.STEP_PROCESS_REWARD:
		step := stakingpool.NewRewardCron(

			o.bc,
			o.conf,
			o.logger.With(zap.String("HookStakingPoolEvent", constants.STEP_PROCESS_REWARD)),
		)
		step.Start()

	case constants.STEP_PROCESS_CREATE_UNSTAKING_WITHDRAW:
		step := stakingpool.NewUnstakeWithdrawCron1(
			o.bc,
			o.conf,
			o.logger.With(zap.String("HookStakingPoolEvent", constants.STEP_PROCESS_CREATE_UNSTAKING_WITHDRAW)),
		)
		step.Start()

	case constants.STEP_PROCESS_CHECK_UNSTAKING_WITHDRAW:
		step := stakingpool.NewUnstakeWithdrawCron2(
			o.notificationDAO,
			o.bc,
			o.conf,
			o.logger.With(zap.String("HookStakingPoolEvent", constants.STEP_PROCESS_CHECK_UNSTAKING_WITHDRAW)),
		)
		step.Start()
	default:
		return &serializers.EventHookResult{
			StatusCode: http.StatusBadRequest,
			IsDone:     true,
			Error:      "The step is not defined",
		}
	}

	return &serializers.EventHookResult{
		StatusCode: http.StatusOK,
		IsDone:     true,
		Error:      "",
	}
}
