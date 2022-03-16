package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type PTokenType int

const (
	Coin PTokenType = iota
	TokenERC20
)

type PToken struct {
	gorm.Model

	TokenID            string // for incognito token
	Symbol             string
	OriginalSymbol     string
	Name               string
	ContractID         string
	Decimals           int64
	PDecimals          int64
	Status             int `sql:"DEFAULT:1"`
	Type               PTokenType
	CurrencyType       CurrencyType
	PSymbol            string
	Default            bool `sql:"DEFAULT:0"`
	UserID             uint
	PriceUsd           float64
	Verified           bool    `sql:"DEFAULT:0"`
	LiquidityReward    float64 `sql:"DEFAULT:0"`
	PercentChange1H    string  `json:"PercentChange1h"`
	PercentChangePrv1H string  `json:"PercentChangePrv1h"`
	CurrentPRVPool     float64 `json:"CurrentPrvPool"`
	PricePrv           float64 `json:"PricePrv"`
	Volume24           float64 `json:"volume24"`
	ParentID uint `sql:"DEFAULT:0"`	
	ListChildToken    []PToken `gorm:"foreignkey:ParentID"`
}

type PCustomToken struct {
	gorm.Model

	TokenID          string
	Image            string
	IsPrivacy        int
	Name             string
	Symbol           string
	OwnerAddress     string
	OwnerName        string
	OwnerEmail       string
	OwnerWebsite     string
	UserID           uint
	ShowOwnerAddress int
	Description      string
	Verified         bool `sql:"DEFAULT:0"`
	Amount           uint64
}

type UserFollowToken struct {
	// gorm.Model
	UserID    uint   `gorm:"primary_key;auto_increment:false"`
	TokenID   string `gorm:"primary_key;auto_increment:false"`
	PublicKey string `gorm:"primary_key;auto_increment:false"`
	CreatedAt time.Time
}
