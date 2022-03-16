package serializers

import "github.com/orgs/Solana-Incognito-Bridge/ognito-service/models"

type EstimateTradingFeesReq struct {
	WalletAddress string `json:"WalletAddress" binding:"required"`
	SrcTokens     string `json:"SrcTokens" binding:"required"`
	DestTokens    string `json:"DestTokens" binding:"required"`

	SrcQties string `json:"SrcQties" binding:"required"`

	RaydiumType int `json:"RaydiumType"`
}

type SubmitTradingTXReq struct {
	ID     uint   `json:"ID" binding:"required"`
	BurnTx string `json:"BurnTx" binding:"required"`

	WalletAddress string `json:"WalletAddress" binding:"required"`
	SrcTokens     string `json:"SrcTokens" binding:"required"`
	DestTokens    string `json:"DestTokens" binding:"required"`

	SrcQties string `json:"SrcQties" binding:"required"`

	UserSelection models.UserFeeSelection `json:"UserFeeSelection" binding:"required"`
	UserFeeLevel  models.FeeSpeedOption   `json:"UserFeeLevel" binding:"required"`

	ExpectedOutputAmount string `json:"ExpectedOutputAmount"`
	Path                 string `json:"Path" binding:"required"`
}

type SubmitUniSwapTXReq struct {
	// SubmitTradingTXReq
	ID     uint   `json:"ID" binding:"required"`
	BurnTx string `json:"BurnTx" binding:"required"`

	WalletAddress string `json:"WalletAddress" binding:"required"`
	SrcTokens     string `json:"SrcTokens" binding:"required"`
	DestTokens    string `json:"DestTokens" binding:"required"`

	SrcQties string `json:"SrcQties" binding:"required"`

	UserSelection models.UserFeeSelection `json:"UserFeeSelection" binding:"required"`
	UserFeeLevel  models.FeeSpeedOption   `json:"UserFeeLevel" binding:"required"`

	ExpectedOutputAmount string `json:"ExpectedOutputAmount"`
	Path                 string `json:"Path" binding:"required"`

	Fee      string `json:"Fee" binding:"required"`
	Percents string `json:"Percents" binding:"required"`
	// IsMulti  bool   `json:"IsMulti, omitempty" binding:"required"`
	IsMulti *bool `json:"IsMulti" binding:"required"`
}

type SubmitCurveTXReq struct {
	// SubmitTradingTXReq
	ID     uint   `json:"ID" binding:"required"`
	BurnTx string `json:"BurnTx" binding:"required"`

	WalletAddress string `json:"WalletAddress" binding:"required"`
	SrcTokens     string `json:"SrcTokens" binding:"required"`
	DestTokens    string `json:"DestTokens" binding:"required"`

	SrcQties string `json:"SrcQties" binding:"required"`

	UserSelection models.UserFeeSelection `json:"UserFeeSelection" binding:"required"`
	UserFeeLevel  models.FeeSpeedOption   `json:"UserFeeLevel" binding:"required"`

	ExpectedOutputAmount string `json:"ExpectedOutputAmount"`
}
