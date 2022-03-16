package api

import (
	"strconv"

	"github.com/incognito-services/webhook"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/incognito-services/service"
	"go.uber.org/zap"
)

type Server struct {
	g       *gin.Engine
	authMw  *jwt.GinJWTMiddleware
	up      *websocket.Upgrader
	userSvc *service.User
	logger  *zap.Logger
	staSvc  *service.StakingPoolService
	hookSvc *webhook.HookService
}

func (s *Server) pagingFromContext(c *gin.Context) (int, int) {
	var (
		pageS  = c.DefaultQuery("page", "1")
		limitS = c.DefaultQuery("limit", "10")
		page   int
		limit  int
		err    error
	)

	page, err = strconv.Atoi(pageS)
	if err != nil {
		page = 1
	}

	limit, err = strconv.Atoi(limitS)
	if err != nil {
		limit = 10
	}

	return page, limit
}

func NewServer(g *gin.Engine, up *websocket.Upgrader,
	userSvc *service.User,
	staSvc *service.StakingPoolService,
	hookSvc *webhook.HookService,
	logger *zap.Logger) *Server {
	return &Server{
		g:       g,
		up:      up,
		userSvc: userSvc,
		logger:  logger,
		staSvc:  staSvc,
		hookSvc: hookSvc,
	}
}
