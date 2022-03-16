package models

import "github.com/jinzhu/gorm"

type FaucetTokenHistory struct {
	gorm.Model

	Amount         uint64
	Asset          string
	PaymentAddress string
	TxID           string
}

type AddressType int

const (
	_        AddressType = iota //0
	Deposit                     //1
	Withdraw                    //2

)

// func (r AddressType) String() string {
// 	return [...]string{"", "deposit", "withdraw", "Swap"}[r]
// }

type CurrencyType int

const (
	_        CurrencyType = iota
	ETH                   //1
	BTC                   //2
	ERC20                 //3
	BNB                   //4
	BNB_BEP2              //5
	USD                   //6

	BNB_BSC   //7
	BNB_BEP20 //8

	TOMO //9
	ZIL  //10
	XMR  //11
	NEO  //12
	DASH //13
	LTC  //14
	DOGE //15
	ZEC  //16
	DOT  //17
	PDEX //18 0000000000000000000000000000000000000000000000000000000000000006

	// Polygon:
	MATIC     //19
	PLG_ERC20 //20

	FTM       //21
	FTM_ERC20 //22

	SOL     //23
	SOL_SPL //24

)
