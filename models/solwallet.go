package models

import (
	"github.com/jinzhu/gorm"
)

type ShieldSolWallet struct {
	gorm.Model

	IncAddress          string `gorm:"unique;not null;index:incAddress"`
	SolAddress          string // ref to Address
	Address             string `gorm:"unique;not null;index:incAddress"`
	SPLTokenID          string
	PubKey              string
	PrivKey             string
	TxSign              string
	KeyVersion          string `sql:"DEFAULT:'1'"`
	SignPublicKeyEncode string
}

type VaultSolToken struct {
	gorm.Model
	SplToken     string `gorm:"unique;not null;index:vaultAddress"`
	VaultAddress string `gorm:"unique;not null;index:vaultAddress"`
	Tx           string
}
