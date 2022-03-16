package api

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/incognito-services/helpers"
	"github.com/incognito-services/models"
	"github.com/incognito-services/serializers"
	"github.com/incognito-services/service"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (s *Server) Authenticate(c *gin.Context) (interface{}, error) {
	var req serializers.UserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	return s.userSvc.Authenticate(req.Email, req.Password)
}

func (s *Server) Auth(c *gin.Context) {
	var req serializers.AuthReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, serializers.Resp{Error: service.ErrInvalidArgument})
		return
	}
	user, err := s.userSvc.RegisterUserByAddress(&req)

	req.IP = helpers.GetIPAdress(c.Request)

	switch cErr := errors.Cause(err); cErr {
	case service.ErrInvalidPassword:
		c.JSON(http.StatusBadRequest, serializers.Resp{Error: cErr.(*service.Error)})
	case nil:
		var NewBufferString = fmt.Sprintf("{\"Email\":\"%s\", \"Password\": \"%s\"}", user.Email, req.Password)
		c.Request, _ = http.NewRequest(c.Request.Method, c.Request.URL.Scheme, bytes.NewBufferString(NewBufferString))

		s.authMw.LoginHandler(c)
	default:
		s.logger.Error("u.svc.Auth", zap.Error(err))
		c.JSON(http.StatusInternalServerError, serializers.Resp{Error: service.ErrInternalServerError})
	}
}

func (s *Server) userFromContext(c *gin.Context) (*models.User, error) {
	userIDVal, ok := c.Get(userIDKey)
	if !ok {
		return nil, errors.New("failed to get userIDKey from context")
	}

	userID := userIDVal.(float64)
	user, err := s.userSvc.FindByID(uint(userID))
	if err != nil {
		return nil, errors.Wrap(err, "s.userSvc.FindByID")
	}
	return user, nil
}
