package api

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/inc-backend/api-ota-eta/serializers"
	"github.com/inc-backend/api-ota-eta/service"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (s *Server) GenerateAddressSol(c *gin.Context) {
	user, err := s.userFromContext(c)
	if err != nil {
		s.logger.Error("s.userFromContext", zap.Error(err))
		c.JSON(http.StatusUnauthorized, serializers.Resp{Error: service.ErrInternalServerError})
		return
	}
	//// show log post:
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
		log.Println("user submit GenerateAddressSol log: user_id:", user.ID, "post data:", string(bodyBytes))
	}
	// Restore the io.ReadCloser to its original state
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	//// end show log post.

	var req serializers.ShieldNewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("c.ShouldBindJSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, serializers.Resp{Error: service.ErrInvalidArgument})
		return
	}
	resp, err := s.solSvc.GenerateAddress(user, &req)

	switch cErr := errors.Cause(err); cErr {
	case service.AddressNotFound, service.MissTokenAddress:
		c.JSON(http.StatusBadRequest, serializers.Resp{Error: cErr.(*service.Error)})
	case nil:
		c.JSON(http.StatusOK, serializers.Resp{Result: resp})
	default:
		s.logger.Error("u.solSvc.GenerateAddress", zap.Error(err))
		c.JSON(http.StatusBadRequest, serializers.Resp{Error: err.Error()})
	}
}

func (s *Server) GenSolanaAddress(c *gin.Context) {
	user, err := s.userFromContext(c)
	if err != nil {
		s.logger.Error("s.userFromContext", zap.Error(err))
		c.JSON(http.StatusUnauthorized, serializers.Resp{Error: service.ErrInternalServerError})
		return
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, serializers.Resp{Error: service.ErrInternalServerError})
	}

	var req serializers.UnshieldNewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("c.ShouldBindJSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, serializers.Resp{Error: service.ErrInvalidArgument})
		return
	}
	resp, err := s.solSvc.EstimateFees(user, &req)
	if err != nil {
		s.logger.Error("s.solSvc.EstimateFees", zap.Error(err))
		c.JSON(http.StatusInternalServerError, serializers.Resp{Error: service.ErrInternalServerError})
		return
	}
	c.JSON(http.StatusOK, serializers.Resp{Result: resp})
}

func (s *Server) AddNewTXWithdrawSol(c *gin.Context) {
	user, err := s.userFromContext(c)
	if err != nil {
		s.logger.Error("s.userFromContext", zap.Error(err))
		c.JSON(http.StatusUnauthorized, serializers.Resp{Error: service.ErrInternalServerError})
		return
	}

	var req serializers.AddNewTXBscWithdrawReq
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("c.ShouldBindJSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, serializers.Resp{Error: service.ErrInvalidArgument})
		return
	}
	resp, err := s.solSvc.AddNewTXWithdraw(user, &req)
	if err == service.ErrInvalidArgument {
		s.logger.Error("c.ShouldBindJSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, serializers.Resp{Error: service.ErrInvalidArgument})
		return
	} else if err != nil {
		s.logger.Error("s.etaSvc.AddNewTXWithdraw", zap.Error(err))
		c.JSON(http.StatusInternalServerError, serializers.Resp{Error: err})
		return
	}
	c.JSON(http.StatusOK, serializers.Resp{Result: resp})
}
