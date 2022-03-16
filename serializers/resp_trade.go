package serializers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/inc-backend/go-incognito/common"

	config "github.com/orgs/Solana-Incognito-Bridge/ognito-service/conf"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/models"
)

type Fees struct {
	Level1 string `json:"Level1,omitempty"`
	Level2 string `json:"Level2,omitempty"`
	Level3 string `json:"Level3,omitempty"`
	Level4 string `json:"Level4,omitempty"`
}

const (
	RaydiumPending  = "Pending"
	RaydiumAccepted = "Accepted"
	RaydiumReject   = "Rejected"
)

type EstimateTradingFeesResp struct {
	ID          uint
	FeeAddress  string
	SignAddress string

	TokenFees   *Fees `json:"TokenFees,omitempty"`
	PrivacyFees *Fees `json:"PrivacyFees,omitempty"`

	// QuoteData QuoteDataResp `json:"QuoteData,omitempty"`
}

type QuoteDataResp1 struct {
	Message string `json:"message"`
	Data    struct {
		FeePath []struct {
			TokenAddress0 string `json:"token_address_0"`
			TokenAddress1 string `json:"token_address_1"`
			Fee           string `json:"fee"`
		} `json:"feePath"`

		Path               []string `json:"path"`
		ExactIn            string   `json:"exactIn"`
		GasAdjustedQuoteIn string   `json:"gasAdjustedQuoteIn"`
		GasUsedQuoteToken  string   `json:"gasUsedQuoteToken"`
		GasUsedUSD         string   `json:"gasUsedUSD"`
	} `json:"data"`
}

type QuoteDataResp struct {
	TokenIn      string
	TokenOut     string
	AmountIn     string
	AmountInRaw  string
	AmountOut    string
	AmountOutRaw string
	PriceImpact  float64
}

type RaydiumHistoryResp struct {
	Id     uint `json:"id,omitempty"`
	UserID uint `json:"userID,omitempty"`

	WalletAddress string `json:"walletAddress,omitempty"`
	// SignAddress   string

	SrcTokens  string `json:"sellTokenId,omitempty"`
	DestTokens string `json:"buyTokenId,omitempty"`

	SrcSymbol  string `json:"srcSymbol,omitempty"`
	DestSymbol string `json:"destSymbol,omitempty"`

	SrcContractAddress  string `json:"srcContractAddress,omitempty"`
	DestContractAddress string `json:"destContractAddress,omitempty"`

	SrcQties             string `json:"amount,omitempty"`
	ExpectedOutputAmount string `json:"mintAccept,omitempty"`
	OutputAmount         string `json:"amountOut,omitempty"`

	Path []string `json:"tradingPath,omitempty"`

	IsNative bool `json:"isNative,omitempty"`

	Status        models.RaydiumStatus `json:"statusCode,omitempty"`
	StatusMessage string               `json:"status,omitempty"`
	StatusDetail  string               `json:"statusDetail,omitempty"`

	//fees and charge fees
	//PrivacyFee uint64 `json:"fee,omitempty"`

	// OutsideChainPrivacyFee string
	FeeToken string `json:"feeToken,omitempty"`
	Fee      string `json:"fee,omitempty"`
	FeeLevel int    `json:"feeLevel,omitempty"`

	//tx
	BurnTx string `json:"requestTx,omitempty"`

	SubmitProofTx string `json:"submitProofTx,omitempty"`

	ExecuteSwapTx string `json:"executeSwapTx,omitempty"`

	WithdrawTx string `json:"withdrawTx,omitempty"`

	MintTx   string `json:"mintTx,omitempty"`
	RefurnTx string `json:"refurnTx,omitempty"`

	CreatedAt int64 `json:"requestime,omitempty"`

	RespondTxs []string `json:"respondTxs,omitempty"`
}

func NewRaydiumHistorysResp(data models.Raydium, conf *config.Config) *RaydiumHistoryResp {
	history := &RaydiumHistoryResp{
		Id:     data.ID,
		UserID: data.UserID,

		WalletAddress: data.WalletAddress,
		// SignAddress:   data.SignAddress,

		SrcTokens:  data.SrcTokens,
		DestTokens: data.DestTokens,

		SrcSymbol:  data.SrcSymbol,
		DestSymbol: data.DestSymbol,

		SrcContractAddress:  data.SrcContractAddress,
		DestContractAddress: data.DestContractAddress,

		SrcQties:             data.SrcQties,
		ExpectedOutputAmount: data.ExpectedOutputAmount,
		OutputAmount:         data.OutputAmount,
		IsNative:             data.IsNative,

		Status: data.Status,

		//fees and charge fees
		// OutsideChainPrivacyFee: data.OutsideChainPrivacyFee,
		FeeLevel: int(data.UserFeeLevel),

		//tx
		BurnTx: data.BurnTx,

		SubmitProofTx: data.SubmitProofTx,

		ExecuteSwapTx: data.ExecuteSwapTx,

		WithdrawTx: data.WithdrawTx,

		MintTx:   data.MintTx,
		RefurnTx: data.RefurnTx,

		CreatedAt: data.CreatedAt.Unix(),
	}

	if data.UserFeeSelection == models.ByPRV {
		history.FeeToken = common.PRVIDStr
	} else {
		history.FeeToken = data.SrcTokens
	}

	// path:
	history.Path = strings.Split(data.Path, ",")

	fee, _ := strconv.ParseInt(data.UserFeeAmount, 10, 64)
	if fee > 0 {
		history.Fee = data.UserFeeAmount
	} else {
		var prvFees *Fees
		err := json.Unmarshal([]byte(data.OutsideChainPrivacyFee), &prvFees)
		if err == nil {
			switch data.UserFeeLevel {
			case models.One:
				history.Fee = prvFees.Level1
			case models.Two:
				history.Fee = prvFees.Level2
			}
		}
	}

	// RefurnTx
	if len(data.MintTx) > 0 {
		history.RespondTxs = strings.Split(data.MintTx, ",")
	}

	if len(data.RefurnTx) > 0 {
		history.RespondTxs = strings.Split(data.RefurnTx, ",")
	}

	if data.EstFeeAt != nil {
		history.CreatedAt = (*data.EstFeeAt).Unix()
	}

	history.setStatusDetails()
	// history.setReceivedOutChainTx(conf)

	return history
}

func NewRaydiumHistoryListResp(trades []*models.Raydium, conf *config.Config) []RaydiumHistoryResp {
	var historyListResp []RaydiumHistoryResp
	for _, trade := range trades {
		historyResp := NewRaydiumHistorysResp(*trade, conf)
		historyListResp = append(historyListResp, *historyResp)
	}
	return historyListResp
}

func (result *RaydiumHistoryResp) setStatusDetails() {

	linkTosupport := "https://we.incognito.org/t/instructions-to-contact-support/5287"

	supportMessage := fmt.Sprintf("Please contact support for assistance, <a rel='noopener noreferrer' target='_blank' href='%s'>[using these instructions]</a>", linkTosupport)

	//supportMessageWait := fmt.Sprintf("This is taking longer than expected. Please wait while your transaction is retried.<br />If you require further support, <a rel='noopener noreferrer' target='_blank' href='%s'>[please follow these instructions]</a>.", linkTosupport)

	processStatus := fmt.Sprintf("Processing [%d]", result.Status)

	switch result.Status {
	case models.ReceivedBurnTx:
		// 1 received the burn tx from the app:
		result.StatusMessage = RaydiumPending
		result.StatusDetail = fmt.Sprintf("Exiting Incognito mode. %v", processStatus)

	case models.TxBurnInvalid:
		// 2: burn tx invalid
		result.StatusMessage = RaydiumReject
		result.StatusDetail = fmt.Sprintf("Exiting Incognito mode. %v", processStatus)

	case models.TxBurnSuccess:
		result.StatusMessage = RaydiumPending
		result.StatusDetail = fmt.Sprintf("Exiting Incognito mode. %v", processStatus)

	case models.FeeAmountInvalid:
		result.StatusMessage = RaydiumReject
		result.StatusDetail = fmt.Sprintf("Exiting Incognito mode. %v", processStatus)

	case models.SentFee:
		result.StatusMessage = RaydiumPending
		result.StatusDetail = fmt.Sprintf("Depositing funds to smart contract. %v", processStatus)

	case models.SendFeeSuccess:
		result.StatusMessage = RaydiumPending
		result.StatusDetail = fmt.Sprintf("Depositing funds to smart contract. %v", processStatus)

	case models.SendFeeFailed:
		result.StatusMessage = RaydiumReject
		result.StatusDetail = fmt.Sprintf("Depositing funds to smart contract. %v", processStatus)

	// submit proof
	case models.SubmitedProofToSC:
		// 8: submiting proof to sc with the burn tx hash.
		result.StatusMessage = RaydiumPending
		result.StatusDetail = fmt.Sprintf("Depositing funds to smart contract. %v", processStatus)

	case models.SubmitedFroofToSCFailed:
		// 10: submited proof failed
		result.StatusMessage = RaydiumReject
		result.StatusDetail = fmt.Sprintf("Depositing funds to smart contract. %v", processStatus)

	case models.SubmitedProofToSCSucceed:
		// 9: tx success/submited proof succeed
		result.StatusMessage = RaydiumPending
		result.StatusDetail = fmt.Sprintf("Depositing funds to smart contract. %v", processStatus)

	// swap:
	case models.SwapedCoin:
		// 11 call to swap token
		result.StatusMessage = RaydiumPending
		result.StatusDetail = fmt.Sprintf("Swapping. %v", processStatus)

	case models.SwapedCoinFailed:
		// 12 can not swap coin.
		result.StatusMessage = RaydiumReject
		result.StatusDetail = fmt.Sprintf("Swapping. %v", processStatus)

	case models.SwapedCoinSucceed:
		// 13 swap coin successful.
		result.StatusMessage = RaydiumPending
		result.StatusDetail = fmt.Sprintf("Swapping. %v", processStatus)

		//todo: withdraw refund
	case models.RequestedWithdraw:
		//  // 14 request withdraw (add fund to vault)
		result.StatusMessage = RaydiumPending
		result.StatusDetail = fmt.Sprintf("Turning your public coins into privacy coins. %v", processStatus)

	case models.WithdrawSuccess:
		result.StatusMessage = RaydiumPending
		result.StatusDetail = fmt.Sprintf("Turning your public coins into privacy coins. %v", processStatus)

	case models.WithdrawFailed:
		result.StatusMessage = RaydiumReject
		result.StatusDetail = fmt.Sprintf("Turning your public coins into privacy coins. %v", processStatus)

	//todo: mint
	case models.MintRequested:
		// 17: call mint token:
		result.StatusMessage = RaydiumPending
		result.StatusDetail = fmt.Sprintf("Turning your public coins into privacy coins. %v", processStatus)

	case models.MintTxRejected:
		//19 tx mint rejected! status birdge = 3
		result.StatusMessage = RaydiumReject
		result.StatusDetail = fmt.Sprintf("Turning your public coins into privacy coins. %v", processStatus)

	case models.MintFailed:
		//20 mint failed! status bridge = 0
		result.StatusMessage = RaydiumReject
		result.StatusDetail = fmt.Sprintf("Turning your public coins into privacy coins. %v", processStatus)

	case models.Minted:
		// 18: mint success:
		result.StatusMessage = RaydiumAccepted
		result.StatusDetail = "Swap successfully"

		//todo: withdraw refund
	case models.NeedToRefundRequest:
		//21: swap failed:
		result.StatusMessage = RaydiumPending
		result.StatusDetail = fmt.Sprintf("Refunding your public coins into privacy coins. %v", processStatus)

	case models.RefundRequestedWithdraw:
		//21: swap failed:
		result.StatusMessage = RaydiumPending
		result.StatusDetail = fmt.Sprintf("Refunding your public coins into privacy coins. %v", processStatus)

	case models.RefundWithdrawSuccess:
		//21: swap failed:
		result.StatusMessage = RaydiumPending
		result.StatusDetail = fmt.Sprintf("Refunding your public coins into privacy coins. %v", processStatus)

		//todo: mint
	case models.RefundWithdrawFailed:
		//21: swap failed:
		result.StatusMessage = RaydiumReject
		result.StatusDetail = fmt.Sprintf("Refunding your public coins into privacy coins. %v", processStatus)

	case models.RefurnRequested:
		// 22: requested mint to refund:
		result.StatusMessage = RaydiumPending
		result.StatusDetail = fmt.Sprintf("Refunding your public coins into privacy coins. %v", processStatus)

	case models.RefurnTxRejected:
		// tx mint for refund faield:
		result.StatusMessage = RaydiumReject
		result.StatusDetail = fmt.Sprintf("Refunding your public coins into privacy coins. %v", processStatus)

	case models.RefurnFailed:
		// tx mint for refund faield
		result.StatusMessage = RaydiumReject
		result.StatusDetail = fmt.Sprintf("Refunding your public coins into privacy coins. %v", processStatus)

	case models.Refurned:
		result.StatusMessage = RaydiumReject
		result.StatusDetail = "Refunded successfully"

	case models.Invalid:
		// unknow issue:
		result.StatusMessage = RaydiumReject
		result.StatusDetail = supportMessage
	}
}

func (history *RaydiumHistoryResp) setReceivedOutChainTx(conf *config.Config) {

	// bsc tx:
	if history.SubmitProofTx != "" {
		history.SubmitProofTx = conf.SmartChain.NetWorkEndPoint + "tx/" + history.SubmitProofTx
	}
	if history.ExecuteSwapTx != "" {
		history.ExecuteSwapTx = conf.SmartChain.NetWorkEndPoint + "tx/" + history.ExecuteSwapTx
	}
	if history.WithdrawTx != "" {
		history.WithdrawTx = conf.SmartChain.NetWorkEndPoint + "tx/" + history.WithdrawTx
	}

	// incognito tx:
	if history.BurnTx != "" {
		history.BurnTx = conf.Incognito.NetWorkEndPoint + "tx/" + history.BurnTx
	}
	if history.MintTx != "" {
		history.MintTx = conf.Incognito.NetWorkEndPoint + "tx/" + history.MintTx
	}
	if history.RefurnTx != "" {
		history.RefurnTx = conf.Incognito.NetWorkEndPoint + "tx/" + history.RefurnTx
	}

}
