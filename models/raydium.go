package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

var MaxErr int = 10

type UserFeeSelection int

const (
	_ UserFeeSelection = iota
	ByToken
	ByPRV
)

type FeeSpeedOption int

const (
	_ FeeSpeedOption = iota
	One
	Two // x2 faster
)

type RaydiumType int

const (
	_RaydiumType RaydiumType = iota
	Pancake
	Uniswap
	Curve
	Raydium
)

type RaydiumStatus int

const (

	// est fee only
	EstimateFee RaydiumStatus = iota // 0 est fee...

	// save the burn tx:
	ReceivedBurnTx   // 1 received the burning tx from the app...
	TxBurnSuccess    // 2 tx success
	TxBurnInvalid    // 3 tx invalid
	FeeAmountInvalid // 4 invalid amount

	SentFee        //5
	SendFeeSuccess //6
	SendFeeFailed  //7

	// submit proof
	SubmitedProofToSC        // 8: submiting proof to sc with the tx hash.
	SubmitedProofToSCSucceed // 9: tx success/submited proof succeed
	SubmitedFroofToSCFailed  // 10: submited proof failed

	// swap:
	SwapedCoin        // 11 call to swap token
	SwapedCoinFailed  // 12 can not swap coin.
	SwapedCoinSucceed // 13 swap coin successful.

	//withdraw
	RequestedWithdraw // 14
	WithdrawSuccess   // 15
	WithdrawFailed    // 16

	// mint incognito
	MintRequested  //17
	Minted         //18: minted coin
	MintTxRejected //19 tx mint rejected!
	MintFailed     //20 mint failed!

	//refund withdraw
	NeedToRefundRequest     //21
	RefundRequestedWithdraw //22
	RefundWithdrawSuccess   // 23
	RefundWithdrawFailed    // 24

	//mint refund
	RefurnRequested  //25
	Refurned         //26: minted to refurn coin
	RefurnTxRejected //27 tx mint rejected!
	RefurnFailed     //28 mint failed!

	Invalid // 29 Stop!
)

type Raydium struct {
	gorm.Model

	UserID   uint `json:"-"`
	ErrCount int  `sql:"DEFAULT:0"`

	RaydiumType RaydiumType `sql:"DEFAULT:1"`

	WalletAddress string
	SignAddress   string
	PrivKey       string
	KeyVersion    string

	SrcTokens  string
	DestTokens string

	SrcSymbol  string
	DestSymbol string

	SrcContractAddress  string
	DestContractAddress string

	SrcQties             string `gorm:"type:decimal(36,0);default:0"` //in-chain
	ExpectedOutputAmount string `gorm:"type:decimal(36,0);default:0"` //in-chain
	OutputAmount         string `gorm:"type:decimal(36,0);default:0"` //in-chain

	Path string
	Fee  string

	IsNative bool
	IsMulti  bool
	Percents string

	Status RaydiumStatus `gorm:"index:status"`

	//fees and charge fees
	TokenFee                        string `sql:"DEFAULT:'0'"`
	PrivacyFee                      string `sql:"DEFAULT:'0'"`
	OutsideChainTokenFee            string `gorm:"type:text"` // fee to send out (by token)
	OutsideChainPrivacyFee          string `gorm:"type:text"` // fee to send out (by PRV)
	UserFeeSelection                UserFeeSelection
	UserFeeLevel                    FeeSpeedOption
	UserFeeAmount                   string
	IncognitoTxToPayOutsideChainFee string

	FeeSendToTempWallet string

	//tx
	BurnTx string

	SendFeeTx  string
	TxFeeNonce uint `gorm:"default:'0'"`

	SubmitProofTx string

	ExecuteSwapTx string

	WithdrawTx string `gorm:"index:withdraw_tx"`

	MintTx   string
	RefurnTx string

	Metadata string `sql:"type:LONGTEXT"`

	ExpiredAt time.Time

	Logs string `sql:"type:LONGTEXT"`

	EstFeeAt *time.Time

	GasPrice string
}

type HistoryStatus int

const (
	HistoryStatusInit    = iota
	HistoryStatusSuccess = 1
	HistoryStatusFailure = 2
)

type RaydiumHistory struct {
	gorm.Model
	JobId uint `gorm:"index:job_id"`

	JobStatus     int
	JobStatusName string

	Status HistoryStatus

	RequestMsg  string `sql:"type:LONGTEXT"`
	ResponseMsg string `sql:"type:LONGTEXT"`
}

type RaydiumReward struct {
	ID        uint `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time `json:"-"`

	DeletedAt *time.Time `gorm:"index" json:"-"`

	RaydiumType RaydiumType `sql:"DEFAULT:1"`

	WalletAddress  string `json:"-"`
	SumTotalVolume float64
	TotalVolume    float64 // usd
	RewardAmount   uint64  // prv

	FromTime *time.Time
	ToTime   *time.Time

	Status      int  // 0 pending, 1 processing, 2: success, 3 failed.
	RequestSend bool `json:"_"`
	Tx          string

	RaydiumIDs   string `json:"-"`
	MaxRaydiumID uint   `json:"-"`

	ErrCount int `json:"_"`
}

// for pCurve:
type CurvePool struct {
	gorm.Model
	PoolAddress string
}
type CurvePoolIndex struct {
	gorm.Model
	CurveTokenIndex  int `sql:"DEFAULT:-1"`
	CurvePoolAddress string
	DappTokenAddress string
	DappTokenSymbol  string
}

//////

/*func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[EstimateFee-0]
	_ = x[ReceivedBurnTx-1]
	_ = x[TxBurnInvalid-2]
	_ = x[FeeAmountInvalid-3]
	_ = x[BurnTxSuccess-4]
	_ = x[SentFee-5]
	_ = x[SendFeeSuccess-6]
	_ = x[SendFeeFailed-7]
	_ = x[FailedGettingBurnProof-8]
	_ = x[BurnProofInvalid-9]
	_ = x[BurnExpiredBsc-10]
	_ = x[SubmitedProofToSC-11]
	_ = x[SubmitedProofToSCSucceed-12]
	_ = x[SubmitedFroofToSCFailed-13]
	_ = x[SwapedCoin-14]
	_ = x[SwapedCoinFailed-15]
	_ = x[SwapedCoinSucceed-16]
	_ = x[RequestedWithdraw-17]
	_ = x[WithdrawSuccess-18]
	_ = x[WithdrawFailed-19]
	_ = x[MintRequested-20]
	_ = x[Minted-21]
	_ = x[MintTxRejected-22]
	_ = x[MintFailed-23]
	_ = x[NeedToRefund-24]
	_ = x[RefurnRequested-25]
	_ = x[Refurned-26]
	_ = x[RefurnTxRejected-27]
	_ = x[RefurnFailed-28]
	_ = x[Invalid-29]
}

const _RaydiumStatus_name = "EstimateFeeReceivedBurnTxTxBurnInvalidFeeAmountInvalidBurnTxSuccessSentFeeSendFeeSuccessSendFeeFailedFailedGettingBurnProofBurnProofInvalidBurnExpiredBscSubmitedProofToSCSubmitedProofToSCSucceedSubmitedFroofToSCFailedSwapedCoinSwapedCoinFailedSwapedCoinSucceedRequestedWithdrawWithdrawSuccessWithdrawFailedMintRequestedMintedMintTxRejectedMintFailedNeedToRefundRefurnRequestedRefurnedRefurnTxRejectedRefurnFailedInvalid"

var _RaydiumStatus_index = [...]uint16{0, 11, 28, 44, 63, 76, 83, 97, 110, 135, 154, 171, 191, 218, 244, 257, 276, 296, 316, 333, 350, 363, 372, 389, 402, 414, 432, 443, 462, 477, 484}

func (i RaydiumStatus) String() string {
	if i < 0 || i >= RaydiumStatus(len(_RaydiumStatus_index)-1) {
		return "RaydiumStatus(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _RaydiumStatus_name[_RaydiumStatus_index[i]:_RaydiumStatus_index[i+1]]
}*/

func ConvertRaydiumToString(status RaydiumStatus) string {
	switch status {
	case EstimateFee:
		return "EstimateFee"
		// save the burn tx:
	case ReceivedBurnTx:
		return "ReceivedBurnTx"
	case TxBurnSuccess: // 4 tx success
		return "TxBurnSuccess"
	case TxBurnInvalid: // 2 tx invalid
		return "TxBurnInvalid"
	case FeeAmountInvalid: // 3 invalid amount
		return "FeeAmountInvalid"
	case SentFee: //5
		return "SentFee"
	case SendFeeSuccess: //6
		return "SendFeeSuccess"
	case SendFeeFailed: //7
		return "SendFeeFailed"

		// submit proof
	case SubmitedProofToSC: // 11: submiting proof to sc with the tx hash.
		return "SubmitedProofToSC"
	case SubmitedProofToSCSucceed: // 12: tx success/submited proof succeed
		return "SubmitedProofToSCSucceed"
	case SubmitedFroofToSCFailed: // 13: submited proof failed
		return "SubmitedFroofToSCFailed"
		// swap:
	case SwapedCoin: // 14 call to swap token
		return "SwapedCoin"
	case SwapedCoinFailed: // 15 can not swap coin.
		return "SwapedCoinFailed"
	case SwapedCoinSucceed: // 16 swap coin successful.
		return "SwapedCoinSucceed"
		//withdraw
	case RequestedWithdraw: // 17
		return "RequestedWithdraw"
	case WithdrawSuccess: // 18
		return "WithdrawSuccess"
	case WithdrawFailed: // 19
		return "WithdrawFailed"
		// mint incognito
	case MintRequested: //20
		return "MintRequested"
	case Minted: //21: minted coin
		return "Minted"
	case MintTxRejected: //22 tx mint rejected!
		return "MintTxRejected"
	case MintFailed: //23 mint failed!
		return "MintFailed"
		//refund withdraw
	case NeedToRefundRequest: //24
		return "NeedToRefundRequest"
	case RefundRequestedWithdraw:
		return "RefundRequestedWithdraw"
	case RefundWithdrawSuccess: // 18
		return "RefundWithdrawSuccess"
	case RefundWithdrawFailed: // 19
		return "RefundWithdrawFailed"
		//mint refund
	case RefurnRequested:
		return "RefurnRequested"
	case Refurned: // 15: minted to refurn coin
		return "Refurned"
	case RefurnTxRejected: //26 tx mint rejected!
		return "RefurnTxRejected"
	case RefurnFailed: //27 mint failed!
		return "RefurnFailed"
	case Invalid: // 28 Stop!
		return "Invalid"
	}

	return ""
}
