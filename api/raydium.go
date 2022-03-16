package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	common "github.com/orgs/Solana-Incognito-Bridge/ognito-service/common"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/serializers"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service"

	"go.uber.org/zap"
)

func (s *Server) GetListTokensRaydium(c *gin.Context) {
	resp, err := s.raydiumSvc.GetListPTokenRaydium()
	if err != nil {
		s.logger.Error("s.GetListPToken", zap.Error(err))
		//return list empty
		c.JSON(http.StatusOK, serializers.Resp{Result: resp})
		return
	}

	c.JSON(http.StatusOK, serializers.Resp{Result: resp})
}

func (s *Server) RaydiumQuote(c *gin.Context) {

	query := c.Request.URL.Query()
	tokenIn := query.Get("tokenIn")
	tokenOut := query.Get("tokenOut")
	amount := query.Get("amount")

	resp, err := s.raydiumSvc.RaydiumQuote(tokenIn, tokenOut, amount)
	if err != nil {
		s.logger.Error("s.RaydiumQuote", zap.Error(err))
		c.JSON(http.StatusOK, serializers.Resp{Result: err})
		return
	}

	c.JSON(http.StatusOK, serializers.Resp{Result: resp})
}

func (s *Server) GenRaydiumAddress(c *gin.Context) {
	user, err := s.userFromContext(c)
	if err != nil {
		s.logger.Error("s.userFromContext", zap.Error(err))
		c.JSON(http.StatusUnauthorized, serializers.Resp{Error: service.ErrInternalServerError})
		return
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, serializers.Resp{Error: service.ErrInternalServerError})
	}

	var req serializers.EstimateTradingFeesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("c.ShouldBindJSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, serializers.Resp{Error: service.ErrInvalidArgument})
		return
	}

	resp, err := s.raydiumSvc.GenRaydiumAddress(user, &req)
	if err != nil {
		s.logger.Error("s.raydiumSvc.GenRaydiumAddress", zap.Error(err))
		c.JSON(http.StatusInternalServerError, serializers.Resp{Error: service.ErrInternalServerError})
		return
	}

	c.JSON(http.StatusOK, serializers.Resp{Result: resp})
}

func (s *Server) HistoryRaydium(c *gin.Context) {
	filter := c.QueryMap("filter")

	if !common.StringInMap("wallet_address", filter) {
		c.JSON(http.StatusBadRequest, serializers.Resp{Error: service.ErrInvalidArgument})
		return
	}

	page, limit := s.pagingFromContext(c)
	resp, total, err := s.raydiumSvc.HistoryRaydium(page, limit, filter)
	if err != nil {
		s.logger.Error("s.raydiumSvc.raydiumSvc", zap.Error(err))
		c.JSON(http.StatusBadRequest, serializers.Resp{Error: service.ErrInternalServerError})
		return
	}

	c.JSON(http.StatusOK, serializers.Resp{
		Result: map[string]interface{}{
			"History": serializers.NewRaydiumHistoryListResp(resp, s.raydiumSvc.GetConfig()),
			"Total":   total,
			"Page":    page,
			"Limit":   limit,
		},
	})
}

func (s *Server) GetListItemsRaydium(c *gin.Context) {
	user, err := s.userFromContext(c)
	if err != nil || user == nil {
		s.logger.Error("s.userFromContext", zap.Error(err))
		c.JSON(http.StatusUnauthorized, serializers.Resp{Error: service.ErrInternalServerError})
		return
	}

	filter := c.QueryMap("filter")

	page, limit := s.pagingFromContext(c)
	resp, total, err := s.raydiumSvc.HistoryRaydium(page, limit, filter)
	if err != nil {
		s.logger.Error("s.raydiumSvc.raydiumSvc", zap.Error(err))
		c.JSON(http.StatusBadRequest, serializers.Resp{Error: service.ErrInternalServerError})
		return
	}

	c.JSON(http.StatusOK, serializers.Resp{
		Result: map[string]interface{}{
			"Data":  resp,
			"Total": total,
			"Page":  page,
			"Limit": limit,
		},
	})
}

func (s *Server) GetItemHistoryByIDsRaydium(c *gin.Context) {
	user, err := s.userFromContext(c)
	if err != nil || user == nil {
		s.logger.Error("s.userFromContext", zap.Error(err))
		c.JSON(http.StatusUnauthorized, serializers.Resp{Error: service.ErrInternalServerError})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	ID := uint(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, serializers.Resp{Error: service.ErrInvalidArgument})
		return
	}

	trade, err := s.raydiumSvc.Detail(ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, serializers.Resp{Error: &service.Error{Code: -9000, Message: err.Error()}})
		return
	}

	//history
	history, err := s.raydiumSvc.GetJobHistoryRaydium(ID)

	if err != nil {
		c.JSON(http.StatusBadRequest, serializers.Resp{Error: &service.Error{Code: -9000, Message: err.Error()}})
		return
	}

	c.JSON(http.StatusOK, serializers.Resp{
		Result: map[string]interface{}{
			"Data":    trade,
			"History": history,
		},
	})
}
