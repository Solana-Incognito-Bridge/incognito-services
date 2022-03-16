package api

import (
	"github.com/gin-gonic/gin"
	"github.com/incognito-services/serializers"
)

func (s *Server) HookStakingPoolEvent(c *gin.Context) {
	step := c.Param("step")
	resp := s.hookSvc.HookStakingPoolEvent(step)
	c.JSON(resp.StatusCode, serializers.Resp{Result: resp})
}
