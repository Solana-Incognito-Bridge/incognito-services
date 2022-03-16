package service

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/inc-backend/crypto-libs/helper"
	go_incognito "github.com/inc-backend/go-incognito"

	"go.uber.org/zap"

	sdk "github.com/inc-backend/sdk/encryption"
	config "github.com/orgs/Solana-Incognito-Bridge/ognito-service/conf"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/dao"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/models"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/serializers"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/service/sol"
	"github.com/pkg/errors"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Shield struct {
	logger    *zap.Logger
	solClient *sol.Client
	config    *config.Config
	bc        *go_incognito.PublicIncognito
	wallet    *go_incognito.Wallet
	solDao    *dao.Shield
}

func NewShield(logger *zap.Logger, solClient *sol.Client, config *config.Config, bc *go_incognito.PublicIncognito, solDao *dao.Shield) *Shield {
	blockInfo := go_incognito.NewBlockInfo(bc)
	wallets := go_incognito.NewWallet(bc, blockInfo)
	return &Shield{logger: logger, solClient: solClient, config: config, bc: bc, wallet: wallets, solDao: solDao}
}

func (o *Shield) GenerateAddress(user *models.User, req *serializers.ShieldNewReq) (*serializers.GenerateAddressResp, error) {

	ptoken, err := o.solDao.GetPTokenByID(req.PrivacyTokenAddress)

	if err != nil || ptoken == nil {
		return nil, MissTokenAddress
	}

	if ptoken.CurrencyType != models.SOL_SPL && ptoken.CurrencyType != models.SOL {
		return nil, MissTokenAddress
	}

	return o.flowGenerateAddressVer2(req, ptoken)
}

func (o *Shield) flowGenerateAddressVer2(req *serializers.ShieldNewReq, shieldToken *models.PToken) (*serializers.GenerateAddressResp, error) {

	var privKey, pubKey, address, txSign string
	var err error

	// create wallet if not exist:
	masterFeePrivateKey, err := sdk.DecryptToString(o.config.Solana.MasterShieldPrivateKey, o.config.Ethereum.KeyDecrypt)
	if err != nil {
		
			return nil, errors.Wrap(err, "sdk.EncryptToString")
		
	}

	// get main account first:
	solWallet, _ := o.solDao.GetItemByIncWallet(req.WalletAddress, constant.Soltoken)
	if solWallet == nil {
		// create wallet if not exist:
		privKey, pubKey, address, txSign, err = o.solClient.GenerateNativeAddress(masterFeePrivateKey)

		if err != nil {
			return nil, err
		}

		privateKeyEndCrypt, err := sdk.EncryptToString(privKey, o.config.Ethereum.KeyDecrypt)		
		if err != nil {
				return nil, errors.Wrap(err, "sdk.EncryptToString")
			
		}
		solWallet = &models.ShieldSolWallet{
			IncAddress:          req.WalletAddress,
			Address:             address,
			SPLTokenID:          constant.Soltoken,
			SolAddress:          address,
			PubKey:              pubKey,
			PrivKey:             privateKeyEndCrypt,
			KeyVersion:          o.config.KeyWalletVersion,
			SignPublicKeyEncode: req.SignPublicKeyEncode,
			TxSign:              txSign,
		}

		if err = o.solDao.CreateShieldSolWallet(solWallet); err != nil {
			return nil, errors.Wrap(err, "Can not create ")
		}
	}
	if shieldToken.ContractID == constant.Soltoken {
		return &serializers.GenerateAddressResp{
			ID:            solWallet.ID,
			FeeAddress:    o.config.Incognito.MasterFeeAddress,
			Address:       solWallet.Address,
			ExpiredAt:     time.Time{},
			Decentralized: 6,
		}, nil
	}

	// find sol wallet token address:
	walletToken, _ := o.solDao.GetItemByIncWallet(req.WalletAddress, shieldToken.ContractID)
	if walletToken == nil {

		masterMakerPrivateKey, err := sdk.DecryptToString(solWallet.PrivKey, o.config.Ethereum.KeyDecrypt)
		if err != nil {			
				return nil, errors.Wrap(err, "sdk.EncryptToString")		
		}

		if len(masterFeePrivateKey) == 0 || len(masterMakerPrivateKey) == 0 {			
			return nil, MissTokenAddress
		}

		pubKey, address, txSign, err = o.solClient.GenerateTokenAddress(masterFeePrivateKey, masterMakerPrivateKey, shieldToken.ContractID)

		if err != nil {
			return nil, errors.Wrap(err, "o.solClient.GenerateTokenAddress()")
		}

		// create wallet token:
		walletToken = &models.ShieldSolWallet{
			IncAddress:          req.WalletAddress,
			Address:             address,
			SPLTokenID:          shieldToken.ContractID,
			SolAddress:          solWallet.Address,
			PubKey:              pubKey,
			PrivKey:             solWallet.PrivKey,
			KeyVersion:          o.config.KeyWalletVersion,
			SignPublicKeyEncode: req.SignPublicKeyEncode,
			TxSign:              txSign,
		}

		if err = o.solDao.CreateShieldSolWallet(walletToken); err != nil {
			return nil, errors.Wrap(err, "Can not create ")
		}
	}
	return &serializers.GenerateAddressResp{
		ID:            walletToken.ID,
		FeeAddress:    o.config.Incognito.MasterFeeAddress,
		Address:       walletToken.Address,
		ExpiredAt:     time.Time{},
		Decentralized: 6,
	}, nil

//gen address to withdraw
func (e *Shield) EstimateFees(user *models.User, req *serializers.UnshieldNewReq) (
	*serializers.EstimateFeesResp, error) {

	ptoken, err := e.solDao.GetPTokenByID(req.PrivacyTokenAddress)

	if err != nil || ptoken == nil {
		return nil, MissTokenAddress
	}

	if ptoken.CurrencyType != models.SOL_SPL && ptoken.CurrencyType != models.SOL {
		return nil, MissTokenAddress
	}

	solToken, _ := e.solDao.GetSOLToken()
	if solToken == nil {
		fmt.Println("Token Polygon Mactic not found!!!")
		return nil, ErrInternalServerError
	}

	if len(strings.TrimSpace(req.RequestedAmount)) == 0 {
		return nil, errors.New("Some data invalid!")
	}

	if len(strings.TrimSpace(req.WalletAddress)) == 0 {
		return nil, errors.New("Some data invalid!!")
	}

	if len(strings.TrimSpace(req.PaymentAddress)) == 0 {
		return nil, errors.New("Some data invalid!!!")
	}

	_, err = e.wallet.GetPaymentAddressV1(req.WalletAddress)
	if err != nil {
		return nil, ErrAddressWrong
	}

	// check exit:
	sol, _ := e.solDao.GetAddressEstFeeExists(req.PaymentAddress, req.WalletAddress, req.PrivacyTokenAddress, req.AddressType)

	if sol != nil {
		fmt.Println("exit record sol: ", sol.ID)
		t1 := time.Now()
		t2 := sol.UpdatedAt
		duration := t1.Sub(t2).Seconds()
		now := time.Now()
		fmt.Println("duration < 30", duration)
		if duration < 30 {
			//return this record
			sol.RequestedAmount = req.RequestedAmount
			sol.ExpiredAt = time.Now()
			sol.EstFeeAt = &now
			sol.Memo = req.Memo

			if err := e.solDao.UpdateAddress(sol); err != nil {
				return nil, errors.Wrap(err, "e.solDao.UpdateAddress")
			}

			var tokenfees *serializers.Fees

			err := json.Unmarshal([]byte(sol.OutsideChainTokenFee), &tokenfees)
			if err == nil {
				var prvfees *serializers.Fees
				err := json.Unmarshal([]byte(sol.OutsideChainPrivacyFee), &prvfees)
				if err == nil {
					return &serializers.EstimateFeesResp{
						ID:          sol.ID,
						FeeAddress:  e.config.Incognito.MasterFeeAddress,
						TokenFees:   tokenfees,
						PrivacyFees: prvfees,
					}, nil
				}
			}
		}
	}

	fee, err := e.returnSolFee(ptoken.TokenID)
	if err != nil {
		return nil, errors.Wrap(err, "s.EstimateFees")
	}

	fmt.Println("Est SOL fee from returnSolFee:", fee)

	var privacyFee, tokenFee uint64

	var jsonTokenFees []byte

	var tokenFees *serializers.Fees

	// cal price:
	tokenFee = fee

	privacyFee = uint64(float64(tokenFee) * solToken.PricePrv)

	if tokenFee > 0 {
		tokenFees = &serializers.Fees{
			Level1: strconv.FormatUint(tokenFee, 10),
			// Level2: new(big.Int).Mul(big.NewInt(2), tokenFee).String(),
		}
		jsonTokenFees, _ = json.Marshal(tokenFees)
	}

	prvFees := &serializers.Fees{
		Level1: strconv.FormatUint(privacyFee, 10),
		// Level2: new(big.Int).Mul(big.NewInt(2), privacyFee).String(),
	}
	jsonPRVFees, _ := json.Marshal(prvFees)

	now := time.Now()

	if sol != nil {
		sol.RequestedAmount = req.RequestedAmount
		sol.Memo = req.Memo
		sol.ExpiredAt = now.Add(time.Hour * 2)
		sol.OutsideChainTokenFee = string(jsonTokenFees)
		sol.OutsideChainPrivacyFee = string(jsonPRVFees)
		sol.EstFeeAt = &now

		if err := e.solDao.UpdateAddress(sol); err != nil {
			return nil, errors.Wrap(err, "e.solDao.UpdateAddress")
		}
		return &serializers.EstimateFeesResp{
			ID:          sol.ID,
			FeeAddress:  e.config.Incognito.MasterFeeAddress,
			TokenFees:   tokenFees,
			PrivacyFees: prvFees,
		}, nil
	}
	// create new record:

	solAddress := &models.Shield{
		UserID:       user.ID,
		ExpiredAt:    now.Add(time.Hour),
		Status:       models.EstimatedWithdrawFeeSol,
		AddressType:  models.Withdraw,
		CurrencyType: ptoken.CurrencyType,
		IncAddress:   req.WalletAddress,
		Symbol:       ptoken.Symbol,

		UserPaymentAddress: req.PaymentAddress,
		Memo:               req.Memo,
		RequestedAmount:    req.RequestedAmount,
		SplToken:           ptoken.ContractID,

		IncognitoToken: req.PrivacyTokenAddress,

		TxMintBurnToken:        req.IncognitoTx,
		OutsideChainTokenFee:   string(jsonTokenFees),
		OutsideChainPrivacyFee: string(jsonPRVFees),
		SignPublicKeyEncode:    req.SignPublicKeyEncode,
		EstFeeAt:               &now,
		TxReceive:              uuid.New().String(),
	}
	if err := e.solDao.Create(solAddress); err != nil {
		return nil, errors.Wrap(err, "o.dao.Create")
	}
	return &serializers.EstimateFeesResp{
		ID:          solAddress.ID,
		FeeAddress:  e.config.Incognito.MasterFeeAddress,
		TokenFees:   tokenFees,
		PrivacyFees: prvFees,
	}, nil

}

func (e *Shield) returnSolFee(tokenID string) (uint64, error) {
	return 5000, nil
}

func (o *Shield) AddNewTXTransferSol(user *models.User, req *serializers.AddNewTXTransferSolReq) (bool, error) {

	// get payment address:
	listTokenAddress, _ := o.solDao.GetShieldSolWalletByIncAddress(req.WalletAddress)

	if listTokenAddress == nil {
		return false, WalletAddressInvalid
	}

	if len(listTokenAddress) == 0 {
		return false, WalletAddressInvalid
	}
	accountAddress := listTokenAddress[0].SolAddress
	// create list wallet address:
	listWalletMap := map[string]string{}
	for _, item := range listTokenAddress {
		listWalletMap[strings.ToLower(item.Address)] = ""
	}

	balance, splTokenID, tokenAddress, isNative, err := o.solClient.GetTxTransferInfo(req.TxID, accountAddress, listWalletMap)

	log.Println("balance, tokenID, tokenAddress, isNative, err", balance, splTokenID, tokenAddress, isNative, err)

	if err != nil {
		return false, err
	}

	if balance == 0 {
		return false, WalletAddressInvalid
	}

	var decimal int64 = 0
	var pDecimal int64 = 0
	var currencyType models.CurrencyType
	var incTokenID = ""
	var splToken = ""

	if isNative {
		solToken, _ := o.solDao.GetSOLToken()
		if solToken == nil {
			return false, TokenInvalid
		}
		pDecimal = solToken.PDecimals
		decimal = solToken.Decimals
		currencyType = solToken.CurrencyType
		incTokenID = solToken.TokenID
		splToken = solToken.ContractID
	} else {
		// get token info:
		shieldToken, _ := o.solDao.GetPTokenByContractID(splTokenID)
		if shieldToken == nil {
			return false, TokenInvalid
		}
		pDecimal = shieldToken.PDecimals
		decimal = shieldToken.Decimals
		currencyType = shieldToken.CurrencyType
		incTokenID = shieldToken.TokenID
		splToken = shieldToken.ContractID
	}

	// create model
	incognitoAmount := helper.ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", balance), decimal, pDecimal)

	shieldUnshieldSol := &models.Shield{
		Address:        tokenAddress,
		IncAddress:     req.WalletAddress,
		FromAddress:    "",
		ReceivedAmount: fmt.Sprintf("%v", balance),
		TxReceive:      req.TxID,
		AddressType:    models.Deposit,
		CurrencyType:   currencyType,

		ExpiredAt: time.Now(),

		Status: models.EstimatedFeeBsc, //models.TxReceiveNew,

		SplToken:       splToken,
		IncognitoToken: incTokenID,

		ShieldAmount:    fmt.Sprintf("%v", balance),
		ChargeFee:       "0",
		IncognitoAmount: incognitoAmount,
		Fee:             "1000000", // 0.001 SOL
	}
	if err = o.solDao.Create(shieldUnshieldSol); err != nil {
		return false, err
	}

	return true, nil
}

func (e *Shield) AddNewTXWithdraw(user *models.User, req *serializers.AddNewTXBscWithdrawReq) (bool, error) {

	ptoken, err := e.solDao.GetPTokenByID(req.PrivacyTokenAddress)
	if err != nil {
		return false, InvalidTokenAddress
	}

	if ptoken.Status == 0 {
		fmt.Println("Invalid token ptoken.Status = 0!!!")
		return false, InvalidTokenAddress
	}

	var incognitoTx = req.IncognitoTx
	var incognitoAmount = req.IncognitoAmount
	if incognitoTx == "" {
		return false, errors.Errorf("incognitoTx is required")
	}
	if incognitoAmount == "" {
		return false, errors.Errorf("incognitoAmount is required")
	}

	sol, err := e.solDao.GetShieldById(req.ID)
	if err != nil || sol == nil {
		return false, WithdrawInvalid
	}

	// check status:
	if sol.Status != models.EstimatedWithdrawFeeSol {
		return false, WithdrawInvalid
	}

	if sol.IncognitoToken != req.PrivacyTokenAddress {
		return false, InvalidTokenAddress
	}

	if sol.IncAddress != req.WalletAddress {
		return false, WalletAddressInvalid
	}

	sol.TxMintBurnToken = req.IncognitoTx
	sol.IncognitoTxToPayOutsideChainFee = req.IncognitoTx
	sol.IncognitoAmount = req.IncognitoAmount
	sol.RequestedAmount = req.RequestedAmount
	if req.UserSelection == models.ByToken && sol.OutsideChainTokenFee == "" {
		return false, ErrInvalidArgument
	}
	sol.UserFeeSelection = req.UserSelection
	sol.UserFeeLevel = req.FeeLevel
	sol.Status = models.ReceivedWithdrawTxSol
	if err := e.solDao.UpdateAddress(sol); err != nil {
		return false, errors.Wrap(err, "e.solDao.UpdateAddress")
	}
	return true, nil
}

func (o *Shield) ListShield(page int, limit int, fields map[string]string) ([]*models.Shield, uint, error) {
	return o.solDao.ListShield(page, limit, fields)
}

func (o *Shield) GetShield(id uint) (*models.Shield, error) {
	return o.solDao.GetShieldById(id)
}

func (o *Shield) GetShieldHistory(id uint) ([]*models.ShieldHistory, error) {
	return o.solDao.GetShieldHistoryById(id)
}
