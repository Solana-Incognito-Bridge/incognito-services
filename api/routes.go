package api

import "github.com/gin-gonic/gin"

func (s *Server) Routes() {

	s.g.GET("/healthz", func(c *gin.Context) {
		c.String(200, "OK")
	})

	// auth API group
	auth := s.g.Group("/auth")
	auth.POST("/", s.Auth)

	auth.Use(s.authMw.MiddlewareFunc())
	{
		auth.GET("/me", s.Profile)
		auth.GET("/profile", s.Profile)
		auth.PUT("/update", s.UpdateUser)
	}

	// shield/unshield Solana network:
	sol := s.g.Group("/sol")
	sol.Use(s.authMw.MiddlewareFunc())
	{
		sol.POST("/generate", s.GenerateAddressSol)
		sol.POST("/estimate-fees", s.GenSolanaAddress)
		sol.POST("/add-tx-withdraw", s.AddNewTXWithdrawSol)
	}

	// swap via raydium:
	raydium := s.g.Group("/raydium")
	raydium.GET("/tokens", s.GetListTokensRaydium)
	raydium.GET("/quote", s.RaydiumQuote)
	raydium.Use(s.authMw.MiddlewareFunc())
	{

		raydium.POST("/estimate-fees", s.GenRaydiumAddress)
		raydium.POST("/submit-trading-tx", s.AddTxRaydium)
		raydium.GET("/history", s.HistoryRaydium)
		raydium.GET("/history/:id", s.Detail)
	}

}
