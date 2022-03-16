package service

import (
	"log"

	"github.com/TheZeroSlave/zapsentry"
	config "github.com/incognito-services/conf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(conf *config.Config) *zap.Logger {
	cfg := zapsentry.Configuration{
		Level: zapcore.ErrorLevel, //when to send message to sentry
		Tags: map[string]string{
			"Service": "Api-Staking-Pool",
		},
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()

	core, err := zapsentry.NewCore(cfg, zapsentry.NewSentryClientFromDSN(conf.SentryDSN))
	//in case of err it will return noop core. so we can safely attach it
	if err != nil {
		logger.Warn("failed to init zap", zap.Error(err))
	}

	return zapsentry.AttachCoreToLogger(core, logger)
}
