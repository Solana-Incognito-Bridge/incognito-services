package api

import (
	"fmt"
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/incognito-services/serializers"
	"github.com/incognito-services/service"
)

const (
	userIDKey    = "id"
	userEmailKey = "email"
)

func (s *Server) WithAuthMw(authMw *jwt.GinJWTMiddleware) {
	s.authMw = authMw
}

func (s *Server) AuthMiddleware(key string) *jwt.GinJWTMiddleware {
	mw, _ := jwt.New(&jwt.GinJWTMiddleware{
		Key:         []byte(key),
		Timeout:     time.Hour * 24 * 7 * 52 * 10, // 10 year
		MaxRefresh:  time.Hour * 24 * 7 * 52 * 10, // 10 year,
		IdentityKey: userIDKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*serializers.UserResp); ok {
				fmt.Printf("v = %+v\n", v)
				return jwt.MapClaims{
					userIDKey:    v.ID,
					userEmailKey: v.Email,
				}
			}
			return jwt.MapClaims{}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			user, err := s.Authenticate(c)
			switch cErr := errors.Cause(err); cErr {
			case service.ErrEmailNotExists, service.ErrInactiveAccount, service.ErrInvalidPassword, service.ErrEmailIsNotVerified:
				return nil, cErr
			case nil:
				return user, nil
			default:
				return nil, err
			}
		},
		HTTPStatusMessageFunc: func(err error, c *gin.Context) string {
			c.Set("authorize_error", err)
			return err.Error()
		},
		Unauthorized: func(c *gin.Context, _ int, _ string) {
			err, _ := c.Get("authorize_error")
			c.JSON(http.StatusUnauthorized, serializers.Resp{
				Error: err,
			})
		},
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			c.JSON(http.StatusOK, serializers.Resp{
				Result: serializers.UserLoginResp{
					Token:   token,
					Expired: expire.Format(time.RFC3339),
				},
				Error: nil,
			})
		},
		RefreshResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			c.JSON(http.StatusOK, serializers.Resp{
				Result: serializers.UserLoginResp{
					Token:   token,
					Expired: expire.Format(time.RFC3339),
				},
				Error: nil,
			})
		},
	})
	return mw
}
