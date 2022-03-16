package service

import (
	go_incognito "github.com/inc-backend/go-incognito"
	config "github.com/orgs/Solana-Incognito-Bridge/ognito-service/conf"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/dao"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service/sol"

	"go.uber.org/zap"
)

type Raydium struct {
	solClient *sol.Client
	bc        *go_incognito.PublicIncognito
	dao       *dao.Raydium
	conf      *config.Config
	logger    *zap.Logger
}

func NewRaydiumService(solClient *sol.Client, bc *go_incognito.PublicIncognito, dao *dao.Raydium, conf *config.Config, logger *zap.Logger) *Raydium {
	return &Raydium{solClient: solClient, bc: bc, dao: dao, conf: conf, logger: logger}
}

func (s *Raydium) GetConfig() *config.Config {
	return s.conf
}
