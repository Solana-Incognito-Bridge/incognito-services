package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type SolFeeHistory struct {
	gorm.Model

	TransactionHash string
	BlockNumber     int64
	BlockHash       string
	TransactionFee  string
	IncTokenID      string
}

type Shield struct {
	gorm.Model

	Address             string `gorm:"index:address"`
	FromAddress         string
	IncAddress          string
	SignPublicKeyEncode string

	TxReceive      string `gorm:"index:tx_receive;unique;not null;"`
	ReceivedAmount string

	ShieldAmount string
	ChargeFee    string

	IncognitoAmount string // convert receive shield amount to.

	Status ShieldStatus

	SplToken       string `gorm:"index:spl_token"`
	IncognitoToken string

	Fee               string
	CurrentSOLBalance string
	SPLBalance        string

	TxFee       string
	TxChargeFee string

	TxApprove       string
	TxDeposit       string
	TxMintBurnToken string

	Logs     string
	ErrCount int

	WalletId    uint
	ReferenceId uint
	NonceIndex  int64

	AddressType  AddressType
	CurrencyType CurrencyType
	Symbol       string

	Note string

	// for unshield:
	UserID                          uint
	UserPaymentAddress              string
	TxWithdraw                      string
	Memo                            string
	RequestedAmount                 string
	TokenFee                        string `sql:"DEFAULT:'0'"`
	PrivacyFee                      string `sql:"DEFAULT:'0'"`
	OutsideChainTokenFee            string `gorm:"type:text"` // fee to send out (by token)
	OutsideChainPrivacyFee          string `gorm:"type:text"` // fee to send out (by PRV)
	UserFeeSelection                UserFeeSelection
	UserFeeLevel                    FeeSpeedOption
	UserFeeAmount                   string
	IncognitoTxToPayOutsideChainFee string

	EstFeeAt  *time.Time
	ExpiredAt time.Time
}

type ShieldStatus int

const (

	// for shield:
	TxReceiveNewSol     = iota // 0 peding
	TxReceiveSuccessSol        //1

	EstimatedFeeSol //2
	SendingFeeSol   //3
	ReceivedFeeSol  //4

	ApproveFeeInvalidSol //5 //need to notify to slack.
	RequestedApproveSol  //6
	ApprovedSol          //7

	DepositFeeInvalidSol //8 //need to notify to slack.
	RequestedDepositSol  //9
	DepositedSol         //10

	RequestedMintSol //11
	MintedSol        //12

	TxRejectedSol    //13 tx mint rejected!
	InvalidFeeSol    //14 wait for the next request.
	ReplacedByFeeSol //15

	InvalidInfoSol // 16 wrong address, token ...

	CollectingFeeSol    // 17
	CollectedFeeSol     // 18
	CollectFeeFailedSol // 19

	// for unshield:
	EstimatedWithdrawFeeSol   // 20: estimated fee amount
	ReceivedWithdrawTxSol     // 21: burning token on incognito
	FailedGettingBurnProofSol // 22: fail getting proof from incognito
	BurnProofInvalidSol       // 23: invalid burn proof, might need to wait for swap
	ReleasingTokenSol         // 24: withdrawing token
	ReleaseTokenSucceedSol    // 25: withdrawing succeed
	ReleaseTokenFailedSol     // 26: withdrawing failed
	ExpiredSol                // 27: expired
	TxInvalidSol              // 28: invalid tx (ex: fee address doesn't match...)
	FeeAmountInvalidSol       // 29: invalid fee amount

)

type ShieldHistory struct {
	gorm.Model
	JobId uint `gorm:"index:job_id"`

	JobStatus     int
	JobStatusName string

	Status HistoryStatus

	RequestMsg  string `sql:"type:LONGTEXT"`
	ResponseMsg string `sql:"type:LONGTEXT"`
}

var ShieldStatusName = map[int]string{
	TxReceiveNewSol:     "TxReceiveNew",
	TxReceiveSuccessSol: "TxReceiveSuccess",

	EstimatedFeeSol: "EstimatedFee",
	SendingFeeSol:   "SendingFee",
	ReceivedFeeSol:  "ReceivedFee",

	ApproveFeeInvalidSol: "ApproveFeeInvalid",
	RequestedApproveSol:  "RequestedApprove",
	ApprovedSol:          "Approved",

	DepositFeeInvalidSol: "DepositFeeInvalid",
	RequestedDepositSol:  "RequestedDeposit",
	DepositedSol:         "Deposited",

	RequestedMintSol: "RequestedMint",
	MintedSol:        "Minted",

	TxRejectedSol:    "TxRejected",
	InvalidFeeSol:    "InvalidFee",
	ReplacedByFeeSol: "ReplacedByFee",

	InvalidInfoSol: "InvalidInfo",

	CollectingFeeSol: "CollectingFee",
	CollectedFeeSol:  "CollectedFee",

	CollectFeeFailedSol: "CollectFeeFailedSol",

	// for unshield:
	EstimatedWithdrawFeeSol:   "EstimatedWithdrawFee",
	ReceivedWithdrawTxSol:     "ReceivedWithdrawTx",
	FailedGettingBurnProofSol: "FailedGettingBurnProof",
	BurnProofInvalidSol:       "BurnProofInvalid",
	ReleasingTokenSol:         "ReleasingToken",
	ReleaseTokenSucceedSol:    "ReleaseTokenSucceed",
	ReleaseTokenFailedSol:     "ReleaseTokenFailed",
	ExpiredSol:                "Expired",
	TxInvalidSol:              "TxInvalid",
	FeeAmountInvalidSol:       "FeeAmountInvalid",
}
