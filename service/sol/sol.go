package sol

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
)

const SYNC_NATIVE_TAG = 0x11
const NEW_TOKEN_ACC = 0x1
const ACCCOUNT_SIZE = 165
const VAULT_ACC = "G65gJS4feG1KXpfDXiySUGT7c6QosCJcGa4nUZsF55Du"

type Client struct {
	IncognitoProxy string
	ProgramID      string
	RpcClient      *rpc.Client
}

func NewClient(incognitoProxy, programID string, isTestnet bool) *Client {
	env := rpc.DevNet_RPC
	if !isTestnet {
		env = rpc.MainNetBeta_RPC
	}
	return &Client{
		IncognitoProxy: incognitoProxy,
		ProgramID:      programID,
		RpcClient:      rpc.New(env),
	}
}

func (e *Client) GenerateNativeAddress(feePlayerPrivkey string) (privKey string, pubKey string, address, txHash string, err error) {
	account := solana.NewWallet()

	privKey = account.PrivateKey.String()
	address = account.PublicKey().String()

	feePayer, err := solana.PrivateKeyFromBase58(feePlayerPrivkey) // account to create tx.
	shieldMaker, err := solana.PrivateKeyFromBase58(privKey)       // user fixed accout

	// find address:burnProofdAccounts
	shieldNativeTokenAcc, _, err := solana.FindAssociatedTokenAddress(
		shieldMaker.PublicKey(),
		solana.SolMint,
	)

	pubKey = shieldNativeTokenAcc.String()

	// check account exist
	needCreateAccount := false
	_, err = e.RpcClient.GetAccountInfo(context.TODO(), solana.MustPublicKeyFromBase58(shieldNativeTokenAcc.String()))
	if err != nil {
		if err.Error() == "not found" {
			fmt.Println("need init account")
			needCreateAccount = true
		} else {
			log.Println("GetAccountInfo err: ", err)
		}
	}

	if !needCreateAccount {
		err = nil
		return
	}

	// init account:
	recent, err := e.RpcClient.GetRecentBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			associatedtokenaccount.NewCreateInstruction(
				feePayer.PublicKey(),    // account fee to create tx.
				shieldMaker.PublicKey(), // owner of token.
				solana.SolMint,          // token id.
			).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(feePayer.PublicKey()),
	)
	signers := []solana.PrivateKey{
		feePayer,
	}
	sig, err := e.SignAndSendTx(tx, signers)
	if err != nil {
		return
	}

	txHash = sig.String()

	return
}

func (e *Client) ValidAddress(address string) bool {
	_, err := solana.PublicKeyFromBase58(address)
	return err == nil
}

func (e *Client) GenerateTokenAddress(feePlayerPrivkey, shieldMakerPrivateAddress, tokenID string) (pubKey, address, txHash string, err error) {

	feePayer, err := solana.PrivateKeyFromBase58(feePlayerPrivkey) // account to create tx.
	if err != nil {
		log.Println("parse feePayer err", feePayer)
		return
	}

	shieldMaker, err := solana.PrivateKeyFromBase58(shieldMakerPrivateAddress) // user fixed accout
	if err != nil {
		log.Println("parse shieldMaker err", err, shieldMaker.PublicKey())
		return
	}

	mintPubkey, err := solana.PublicKeyFromBase58(tokenID) // pubkey of token to shield. //PublicKeyFromBase58//MustPublicKeyFromBase58
	if err != nil {
		log.Println("parse mintPubkey err", tokenID, err)
		return
	}

	// find address:
	shieldNativeTokenAcc, _, err := solana.FindAssociatedTokenAddress(
		shieldMaker.PublicKey(),
		mintPubkey,
	)
	if err != nil {
		log.Println("Can not FindAssociatedTokenAddress", err)
		return
	}
	log.Println("shieldNativeTokenAcc", shieldNativeTokenAcc)
	address = shieldNativeTokenAcc.String()
	pubKey = address

	//////

	// check account exist
	needCreateAccount := false
	_, err = e.RpcClient.GetAccountInfo(context.TODO(), solana.MustPublicKeyFromBase58(shieldNativeTokenAcc.String()))
	if err != nil {
		if err.Error() == "not found" {
			fmt.Println("need init account")
			needCreateAccount = true
		} else {
			log.Println("GetAccountInfo err: ", err)
		}
	}

	if !needCreateAccount {
		err = nil
		return
	}

	// init account:
	recent, err := e.RpcClient.GetRecentBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			associatedtokenaccount.NewCreateInstruction(
				feePayer.PublicKey(),    // account fee to create tx.
				shieldMaker.PublicKey(), // owner of token.
				mintPubkey,              // token id.
			).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(feePayer.PublicKey()),
	)
	signers := []solana.PrivateKey{
		feePayer,
	}
	sig, err := e.SignAndSendTx(tx, signers)
	if err != nil {
		return
	}

	txHash = sig.String()

	return

}

// ref: https://github.com/gagliardetto/solana-go/blob/main/rpc/getSignatureStatuses.go
func (e *Client) CheckTxStatus(tx string) (bool, error) {

	txBase58, err := solana.SignatureFromBase58(tx)
	if err != nil {
		return false, err
	}

	out, err := e.RpcClient.GetSignatureStatuses(
		context.TODO(),
		true,
		txBase58,
	)
	if err != nil {
		log.Println("GetSignatureStatuses err: ", err)
		return false, err
	}

	return out.Value[0].ConfirmationStatus == rpc.ConfirmationStatusFinalized, nil

}

func (e *Client) SignAndSendTx(tx *solana.Transaction, signers []solana.PrivateKey) (solana.Signature, error) {
	_, err := tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		for _, candidate := range signers {
			if candidate.PublicKey().Equals(key) {
				return &candidate
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("unable to sign transaction: %v \n", err)
		return solana.Signature{}, err
	}
	// send tx
	signature, err := e.RpcClient.SendTransaction(context.Background(), tx)
	if err != nil {
		fmt.Printf("unable to send transaction: %v \n", err)
		return solana.Signature{}, err
	}
	return signature, nil
}

func (e *Client) GetTxTransferInfo(txID, accountAddress string, addressTokenIDs map[string]string) (balance uint64, tokenID, tokenAddress string, isNative bool, err error) {

	// var tokenID string
	// var balance uint64 = 0
	// var isNative = false

	// catch
	txhash, err := solana.SignatureFromBase58(txID) //"2yZtM9jcYG1jZzRLWqgLAnG6tqA7n5fQm8f5Ki1xmwHxaX5vnPK1tbeyarmpw2TYxRUYuJ8C5v1fo8ey5H4Byb9t")
	if err != nil {
		return
	}
	opts := rpc.GetTransactionOpts{
		Commitment: rpc.CommitmentConfirmed,
		Encoding:   solana.EncodingBase58,
	}
	txInfo, err := e.RpcClient.GetTransaction(context.Background(), txhash, &opts)
	if err != nil {
		return
	}

	log.Println("accout: ", accountAddress)
	log.Println("list addressTokenIDs: ", addressTokenIDs)

	fmt.Printf("account pre balances info %+v\n", txInfo.Meta.PostBalances)
	fmt.Printf("account balances info %+v\n", txInfo.Meta.PostBalances)
	fmt.Printf("token account balances info %+v\n", txInfo.Meta.PostTokenBalances)

	txDecode, err := solana.TransactionFromDecoder(bin.NewBinDecoder(txInfo.Transaction.GetBinary()))
	if err != nil {
		return
	}
	fmt.Printf("key list %+v \n", txDecode.Message.AccountKeys)

	// find account:
	for i := 0; i < len(txDecode.Message.AccountKeys); i++ {

		account := txDecode.Message.AccountKeys[i]

		// check with account address:
		if strings.EqualFold(account.String(), accountAddress) {
			balance = txInfo.Meta.PostBalances[i] - txInfo.Meta.PreBalances[i]
			log.Println("balance found: ", balance)
			isNative = true
			tokenAddress = accountAddress
			return

		} else {

			// list in list token:
			_, ok := addressTokenIDs[strings.ToLower(account.String())]
			if ok {
				isNative = false
				// found token here:
				// check with token address:
				for _, tokenPostInfo := range txInfo.Meta.PostTokenBalances {

					// found token:
					if tokenPostInfo.AccountIndex == uint16(i) {

						tokenID = tokenPostInfo.Mint.String()
						// find preTokenBalance
						for _, tokenPreInfo := range txInfo.Meta.PreTokenBalances {
							if tokenPreInfo.AccountIndex == tokenPostInfo.AccountIndex {
								preAmount, _ := strconv.ParseUint(tokenPreInfo.UiTokenAmount.Amount, 10, 64)
								postAmount, _ := strconv.ParseUint(tokenPostInfo.UiTokenAmount.Amount, 10, 64)
								balance = postAmount - preAmount
								tokenID = tokenPostInfo.Mint.String()
								tokenAddress = account.String()
								isNative = false
								return
							}
						}

					}
				}

			}

		}

	}

	err = errors.New("invalid!")

	return
}
func (e *Client) Transfer(fromPrivKey, toAddress string, amount uint64) (string, error) {

	fmt.Println("============ TRANSFER SOL =============")
	log.Println("toAddress, amount", toAddress, amount)

	toPub, err := solana.PublicKeyFromBase58(toAddress)
	if err != nil {
		return "", err
	}

	feePayer, err := solana.PrivateKeyFromBase58(fromPrivKey)
	if err != nil {
		return "", err
	}
	// account to create tx.
	recent, err := e.RpcClient.GetRecentBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return "", err
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(
				amount,
				feePayer.PublicKey(),
				toPub,
			).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(feePayer.PublicKey()),
	)
	signers := []solana.PrivateKey{
		feePayer,
	}
	sig, err := e.SignAndSendTx(tx, signers)
	if err != nil {
		return "", err
	}

	return sig.String(), nil
}

func (e *Client) ReadAllTx(feeAddress, accountAddress, solTokenAddress, lastestTx, receivedTx string) []string {

	log.Println("feeAddress: ", feeAddress, "accountAddress", accountAddress, "solTokenAddress: ", solTokenAddress)
	log.Println("lastestTx: ", lastestTx, "receivedTx", receivedTx)

	// https://api-devnet.solscan.io/account/transaction?address=4ZNvVPr7jki7HJMCaT7P6w88bPiopXv9AndsG59WDJBk&before=
	url := "https://api-devnet.solscan.io/account/transaction?address=%s&before=%s"

	var listTx []string

	method := "GET"

	payload := strings.NewReader(``)

	client := &http.Client{}

	before := ""

	for {

		req, err := http.NewRequest(method, fmt.Sprintf(url, solTokenAddress, before), payload)

		if err != nil {
			fmt.Println(err)
			return listTx
		}
		req.Header.Add("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			return listTx
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return listTx
		}
		// fmt.Println(string(body))

		var txInfo struct {
			Success bool `json:"succcess"`
			Data    []struct {
				TxHash string   `json:"txHash"`
				Signer []string `json:"signer"`
			} `json:"data"`
		}
		err = json.Unmarshal(body, &txInfo)
		if err != nil {
			fmt.Println(err)
			return listTx
		}

		for _, data := range txInfo.Data {
			log.Println("data.Signer[0]", data.Signer[0])
			if data.Signer[0] == accountAddress || data.Signer[0] == feeAddress || data.Signer[0] == solTokenAddress {
				log.Println("Same from/fee address, rejected!", accountAddress, solTokenAddress, feeAddress, data.Signer[0], data.TxHash)
				continue
			}
			if data.TxHash == lastestTx || data.TxHash == receivedTx {
				log.Println("= lastest tx: ", data.TxHash)
				return listTx
			}
			// append:
			listTx = append(listTx, data.TxHash)
		}

		if len(txInfo.Data) < 10 {
			log.Println("end list, len(data)=", len(txInfo.Data))
			break
		}
		before = txInfo.Data[len(txInfo.Data)-1].TxHash
	}

	return listTx
}

func (e *Client) GetAddress() {
	// tokenSell := "BEcGFQK1T1tSu3kvHC17cyCkQ5dvXqAJ7ExB2bb5Do7a"

	signer := solana.MustPrivateKeyFromBase58("3Y6mUgjKHdaNt5CDbRhnhzq6zwSU1CzfuLG6xaCawXQoVy4s9cpz2bhyw7ccineSCsWkFq1XPmNVSB73ZHjv85QQ")

	// program, err := solana.PublicKeyFromBase58(e.ProgramID)
	// if err != nil {
	// 	return
	// }

	// signerTokenAuthority, _, err := solana.FindProgramAddress(
	// 	[][]byte{signer.PublicKey().Bytes()},
	// 	program,
	// )

	// signerSellToken, _, err := solana.FindAssociatedTokenAddress(
	// 	signerTokenAuthority,
	// 	solana.MustPublicKeyFromBase58(tokenSell),
	// )
	log.Println("signerTokenAuthority", signer.PublicKey())
}
