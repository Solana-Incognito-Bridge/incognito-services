package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/incognito-services/constants"
	ginprometheus "github.com/zsais/go-gin-prometheus"

	"github.com/incognito-services/webhook"

	"github.com/incognito-services/common"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/inc-backend/crypto-libs/incognito"
	"github.com/incognito-services/api"
	config "github.com/incognito-services/conf"
	"github.com/incognito-services/dao"
	"github.com/incognito-services/database"
	"github.com/incognito-services/service"
	"github.com/incognito-services/service/3rd/sendgrid"
	"github.com/incognito-services/service/email"
)

func main() {
	conf := config.GetConfig()

	logger := service.NewLogger(conf)
	// defer logger.Sync()

	upgrader := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	db, err := database.Init(conf)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("database.Init")
	}
	// auto migrate test test:
	if err := dao.AutoMigrate(db); err != nil {
		logger.With(zap.Error(err)).Fatal("failed to auto migrate")
	}

	if err != nil {
		logger.Fatal("stdgcloudstorage.NewClient", zap.Error(err))
	}

	var (
		client = &http.Client{}
		bc     = incognito.NewBlockchain(client, conf.Incognito.ChainEndpoint, conf.Incognito.Username, conf.Incognito.Password, conf.Incognito.NetWorkEndPoint, common.ConstantID)

		mailClient  = sendgrid.Init(conf)
		emailHelper = email.New(mailClient)
	)

	stDAO := dao.NewStakingPool(db)
	staSvc := service.NewStakingPoolService(stDAO, bc, conf, logger)
	notificationDao := dao.NewNotificationDAO(db)

	var (
		userDAO = dao.NewUser(db)
		userSvc = service.NewUserService(conf, userDAO, stDAO, bc, emailHelper, logger.With(zap.String("service", "user")))

		hookSvc = webhook.NewHookService(
			stDAO,
			notificationDao,
			bc,
			conf,
			logger,
		)
	)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "HEAD", "OPTIONS", "DELETE"},
		AllowHeaders:    []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		MaxAge:          12 * time.Hour,
	}))

	p := ginprometheus.NewPrometheus("api_service", constants.ResponseTimeHistogram)
	p.SetMetricsPath(r)
	r.Use(api.MeasureResponseDuration(p))

	svr := api.NewServer(r, upgrader, userSvc, staSvc, hookSvc, logger.With(zap.String("module", "api")))

	authMw := svr.AuthMiddleware(conf.TokenSecretKey)
	svr.WithAuthMw(authMw)
	svr.Routes()

	port := conf.Port
	if conf.Env != "development" {
		port = 8888
	}

	address := fmt.Sprintf(":%d", port)
	srv := &http.Server{
		Addr:    address,
		Handler: r,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		fmt.Printf("Listening and serving HTTP on %s\n", address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("router.Run: %s\n", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Error("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
