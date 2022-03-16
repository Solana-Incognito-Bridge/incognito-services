package sol

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"

	"github.com/Solana-Incognito-Bridge/bridge-programs/services-go/shield"
	unshield "github.com/Solana-Incognito-Bridge/bridge-programs/services-go/unshield"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
)

func (e *Client) ShieldNative(privKey, incAddress, vaultAddress string, amount uint64) (txHash string, err error) {

	log.Println("privKey", privKey, "incAddress", incAddress, "vaultAddress", vaultAddress, "amount", amount)

	shieldMaker, err := solana.PrivateKeyFromBase58(privKey) // user fixed accout
	if err != nil {
		return
	}
	log.Println("shieldMaker address", shieldMaker.PublicKey())

	incognitoProxy, err := solana.PublicKeyFromBase58(e.IncognitoProxy)
	if err != nil {
		return
	}

	program := solana.MustPublicKeyFromBase58(e.ProgramID)

	vaultTokenAuthority, _, err := solana.FindProgramAddress(
		[][]byte{incognitoProxy.Bytes()},
		program,
	)
	if err != nil {
		return
	}
	vaultNativeTokenAcc, _, err := solana.FindAssociatedTokenAddress(
		vaultTokenAuthority,
		solana.SolMint,
	)

	if err != nil {
		return
	}

	fmt.Println("vaultNativeTokenAcc:", vaultNativeTokenAcc.String()) //vaultNativeTokenAcc: HCekuqhnuFUq1RqJ4szQg9ikXhmxXvDGn5ubzyxNTgjY

	shieldNativeTokenAcc, _, err := solana.FindAssociatedTokenAddress(
		shieldMaker.PublicKey(),
		solana.SolMint,
	)

	if err != nil {
		return
	}

	log.Println("shieldNativeTokenAcc", shieldNativeTokenAcc)

	shieldAccounts := []*solana.AccountMeta{
		solana.NewAccountMeta(shieldNativeTokenAcc, true, false), // token maker account
		solana.NewAccountMeta(vaultNativeTokenAcc, true, false),  // vault token
		solana.NewAccountMeta(incognitoProxy, false, false),
		solana.NewAccountMeta(shieldMaker.PublicKey(), false, true),
		solana.NewAccountMeta(solana.TokenProgramID, false, false),
	}
	signers := []solana.PrivateKey{
		shieldMaker,
	}

	shieldInstruction := shield.NewShield(
		incAddress,
		amount,
		program,
		shieldAccounts,
		byte(0),
	)

	// account to create tx.
	recent, err := e.RpcClient.GetRecentBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return "", err
	}

	// build sync native token program
	syncNativeeInst := solana.NewInstruction(
		solana.TokenProgramID,
		[]*solana.AccountMeta{
			solana.NewAccountMeta(shieldNativeTokenAcc, true, false),
		},
		[]byte{SYNC_NATIVE_TAG},
	)
	tx, err := solana.NewTransaction(

		[]solana.Instruction{

			system.NewTransferInstruction(
				amount,
				shieldMaker.PublicKey(),
				shieldNativeTokenAcc,
			).Build(),
			syncNativeeInst, // remove khi token
			shieldInstruction.Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(shieldMaker.PublicKey()),
	)

	if err != nil {
		return
	}
	sig, err := e.SignAndSendTx(tx, signers)
	if err != nil {
		return
	}
	txHash = sig.String()

	return
}

func (e *Client) ShieldToken(privKey, incAddress, shieldTokenAddress, vaultAddress string, amount uint64) (txHash string, err error) {

	log.Println("incAddress", incAddress, "shieldTokenAddress", shieldTokenAddress, "vaultAddress", vaultAddress, "amount", amount)

	shieldMaker, err := solana.PrivateKeyFromBase58(privKey) // user fixed accout
	if err != nil {
		return
	}

	log.Println("shieldMaker:", shieldMaker.PublicKey())

	incognitoProxy, err := solana.PublicKeyFromBase58(e.IncognitoProxy)
	if err != nil {
		return
	}

	program := solana.MustPublicKeyFromBase58(e.ProgramID)

	vaultTokenAcc, err := solana.PublicKeyFromBase58(vaultAddress) //
	if err != nil {
		return
	}
	shieldTokenAcc, err := solana.PublicKeyFromBase58(shieldTokenAddress)
	if err != nil {
		return
	}

	shieldAccounts := []*solana.AccountMeta{
		solana.NewAccountMeta(shieldTokenAcc, true, false), // token maker account
		solana.NewAccountMeta(vaultTokenAcc, true, false),  // vault token
		solana.NewAccountMeta(incognitoProxy, false, false),
		solana.NewAccountMeta(shieldMaker.PublicKey(), false, true),
		solana.NewAccountMeta(solana.TokenProgramID, false, false),
	}
	signers := []solana.PrivateKey{
		shieldMaker,
	}

	shieldInstruction := shield.NewShield(
		incAddress,
		amount,
		program,
		shieldAccounts,
		byte(0),
	)

	var tx *solana.Transaction

	// account to create tx.
	recent, err := e.RpcClient.GetRecentBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return "", err
	}

	tx, err = solana.NewTransaction(
		[]solana.Instruction{ // remove cho token
			shieldInstruction.Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(shieldMaker.PublicKey()),
	)

	if err != nil {
		return
	}
	sig, err := e.SignAndSendTx(tx, signers)
	if err != nil {
		return
	}
	txHash = sig.String()

	return
}

func (e *Client) UnShieldToken(feePlayerPrivkey, txBurn, splToken, userPaymentAddress string) (txHash string, err error) {

	feePayer, err := solana.PrivateKeyFromBase58(feePlayerPrivkey) // account to create tx.
	if err != nil {
		return
	}

	vaultAcc := solana.MustPublicKeyFromBase58(VAULT_ACC)

	mintPubkey, err := solana.PublicKeyFromBase58(splToken)
	if err != nil {
		return
	}
	incognitoProxy, err := solana.PublicKeyFromBase58(e.IncognitoProxy)
	if err != nil {
		return
	}

	program, err := solana.PublicKeyFromBase58(e.ProgramID)
	if err != nil {
		return
	}

	vaultTokenAuthority, _, err := solana.FindProgramAddress(
		[][]byte{incognitoProxy.Bytes()},
		program,
	)
	if err != nil {
		return
	}

	vaultTokenAcc, _, err := solana.FindAssociatedTokenAddress(
		vaultTokenAuthority,
		mintPubkey,
	)

	if err != nil {
		return
	}

	recent, err := e.RpcClient.GetRecentBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return
	}

	signers := []solana.PrivateKey{
		feePayer,
	}

	shieldMakerTokenAccount, err := solana.PublicKeyFromBase58(userPaymentAddress)

	if err != nil {
		return
	}

	unshieldAccounts := []*solana.AccountMeta{
		solana.NewAccountMeta(vaultTokenAcc, true, false),
		solana.NewAccountMeta(shieldMakerTokenAccount, true, false),
		solana.NewAccountMeta(vaultTokenAuthority, false, false),
		solana.NewAccountMeta(vaultAcc, true, false),
		solana.NewAccountMeta(incognitoProxy, false, false),
		solana.NewAccountMeta(solana.TokenProgramID, false, false),
	}

	unshield := unshield.NewUnshield(
		txBurn,
		"getsolburnproof",
		"https://fullnode.solana-bridge-demo.incognito.org",
		program,
		unshieldAccounts,
	)
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			unshield.Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(feePayer.PublicKey()),
	)
	sig, err := e.SignAndSendTx(tx, signers)
	if err != nil {
		return
	}
	txHash = sig.String()

	return
}

func (e *Client) UnShieldNative(feePlayerPrivkey, txBurn, userPaymentAddress string) (txHash string, err error) {

	feePayer, err := solana.PrivateKeyFromBase58(feePlayerPrivkey) // account to create tx.
	if err != nil {
		return
	}

	incognitoProxy, err := solana.PublicKeyFromBase58(e.IncognitoProxy)
	if err != nil {
		return
	}

	program, err := solana.PublicKeyFromBase58(e.ProgramID)
	if err != nil {
		return
	}

	vaultTokenAuthority, _, err := solana.FindProgramAddress(
		[][]byte{incognitoProxy.Bytes()},
		program,
	)
	if err != nil {
		return
	}

	vaultNativeTokenAcc, _, err := solana.FindAssociatedTokenAddress(
		vaultTokenAuthority,
		solana.SolMint,
	)

	if err != nil {
		panic(err)
	}
	fmt.Println(vaultNativeTokenAcc.String())

	vaultAcc := solana.MustPublicKeyFromBase58("G65gJS4feG1KXpfDXiySUGT7c6QosCJcGa4nUZsF55Du")

	nativeAccountToken, err := solana.WalletFromPrivateKeyBase58("YpRLgTL3DPc83MTjdsVE6ALv5RqUQt3jZ35aVQDxeAmJLqsDhXZFtPnFXganq6DfQ7Q91guGQjKc13YMVjyX8vP")
	if err != nil {
		return
	}

	shieldMakerTokenAccount, err := solana.PublicKeyFromBase58(userPaymentAddress)

	if err != nil {
		return
	}

	unshieldAccounts := []*solana.AccountMeta{
		solana.NewAccountMeta(vaultNativeTokenAcc, true, false), // vault sol
		solana.NewAccountMeta(nativeAccountToken.PublicKey(), true, false),
		solana.NewAccountMeta(vaultTokenAuthority, false, false),
		solana.NewAccountMeta(vaultAcc, true, false), // hardcode
		solana.NewAccountMeta(incognitoProxy, false, false),
		solana.NewAccountMeta(solana.TokenProgramID, false, false),
		solana.NewAccountMeta(shieldMakerTokenAccount, true, false),
	}

	signers := []solana.PrivateKey{
		feePayer,
		nativeAccountToken.PrivateKey,
	}

	unshields := unshield.NewUnshield(
		txBurn,
		"getsolburnproof",
		"https://fullnode.solana-bridge-demo.incognito.org",
		program,
		unshieldAccounts,
	)
	exemptLamport, err := e.RpcClient.GetMinimumBalanceForRentExemption(context.Background(), ACCCOUNT_SIZE, rpc.CommitmentConfirmed)
	if err != nil {
		return
	}

	// build create new token acc
	newAccToken := solana.NewInstruction(
		solana.TokenProgramID,
		[]*solana.AccountMeta{
			solana.NewAccountMeta(nativeAccountToken.PublicKey(), true, false),
			solana.NewAccountMeta(solana.SolMint, false, false),
			solana.NewAccountMeta(vaultTokenAuthority, false, false),
			solana.NewAccountMeta(solana.SysVarRentPubkey, false, false),
		},
		[]byte{NEW_TOKEN_ACC},
	)
	recent, err := e.RpcClient.GetRecentBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewCreateAccountInstruction(
				exemptLamport,
				ACCCOUNT_SIZE,
				solana.TokenProgramID,
				feePayer.PublicKey(),
				nativeAccountToken.PublicKey(),
			).Build(),
			newAccToken,
			unshields.Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(feePayer.PublicKey()),
	)
	sig, err := e.SignAndSendTx(tx, signers)
	if err != nil {
		return
	}

	txHash = sig.String()

	return
}
func (e *Client) GetVaultNativeWrapForAccount(feePlayerPrivkey, addressMaker string) (vaultAddress, txHash string, err error) {
	addressMakerPub, err := solana.PublicKeyFromBase58(addressMaker)

	// find address:
	shieldNativeTokenAcc, _, err := solana.FindAssociatedTokenAddress(
		addressMakerPub,
		solana.SolMint,
	)

	vaultAddress = shieldNativeTokenAcc.String()

	log.Println("native vault: ", vaultAddress)

	// check account exist
	needCreateAccount := false
	acc, err := e.RpcClient.GetAccountInfo(context.TODO(), solana.MustPublicKeyFromBase58(shieldNativeTokenAcc.String()))
	if err != nil {
		if err.Error() == "not found" {
			fmt.Println("need init account")
			needCreateAccount = true
		} else {
			log.Println("GetAccountInfo err: ", err)
		}
	}
	log.Println("acc", acc)

	if !needCreateAccount {
		err = nil
		return
	}

	// init account:
	recent, err := e.RpcClient.GetRecentBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return
	}

	feePayer, err := solana.PrivateKeyFromBase58(feePlayerPrivkey) // account to create tx.
	if err != nil {
		return
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			associatedtokenaccount.NewCreateInstruction(
				feePayer.PublicKey(), // account fee to create tx.
				addressMakerPub,      // owner of token.
				solana.SolMint,       // token id.
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

func (e *Client) CreateVaultAddress(feePlayerPrivkey, tokenID string) (vaultAddress, txHash string, err error) {

	mintPubkey, err := solana.PublicKeyFromBase58(tokenID)
	if err != nil {
		return
	}
	incognitoProxy, err := solana.PublicKeyFromBase58(e.IncognitoProxy)
	if err != nil {
		return
	}

	program, err := solana.PublicKeyFromBase58(e.ProgramID)
	if err != nil {
		return
	}

	vaultTokenAuthority, _, err := solana.FindProgramAddress(
		[][]byte{incognitoProxy.Bytes()},
		program,
	)
	if err != nil {
		return
	}

	vaultAddressObj, _, err := solana.FindAssociatedTokenAddress(
		vaultTokenAuthority,
		mintPubkey,
	)

	if err != nil {
		return
	}
	vaultAddress = vaultAddressObj.String()

	log.Println("vaultAddress: ", vaultAddress)

	// check before create:
	needCreateAccount := false
	_, err = e.RpcClient.GetAccountInfo(context.TODO(), solana.MustPublicKeyFromBase58(vaultAddress))
	if err != nil {
		if err.Error() == "not found" {
			fmt.Println("need init account")
			needCreateAccount = true
		} else {
			log.Println("GetAccountInfo err: ", err)
		}
	}

	if !needCreateAccount {
		log.Println("exits account vault address!")
		err = nil
		return
	}

	// init account:
	recent, err := e.RpcClient.GetRecentBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return
	}

	feePayer, err := solana.PrivateKeyFromBase58(feePlayerPrivkey) // account to create tx.
	if err != nil {
		return
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			associatedtokenaccount.NewCreateInstruction(
				feePayer.PublicKey(), // account fee to create tx.
				vaultTokenAuthority,  // owner of token.
				mintPubkey,           // token id.
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

// SWAP ON RAYDIUM
func (e *Client) Swap(feePlayerPrivkey, privKey, tokenIn, tokenOut string, amountIn, expectAmountOut uint64) (txHash string, err error) {

	feePayer, err := solana.PrivateKeyFromBase58(feePlayerPrivkey) // account to create tx.
	if err != nil {
		return
	}

	tokenSell := solana.MustPublicKeyFromBase58(tokenIn) //7a62ccb67d889804a436ca7c687fdfe0e4a9a6debafff3748e131c71ee82d6d0
	tokenBuy := solana.MustPublicKeyFromBase58(tokenOut) //

	signer, err := solana.PrivateKeyFromBase58(privKey)

	if err != nil {
		return
	}

	signersSwap := []solana.PrivateKey{
		signer,
		feePayer,
	}

	program := solana.MustPublicKeyFromBase58(e.ProgramID)

	signerTokenAuthority, _, err := solana.FindProgramAddress(
		[][]byte{signer.PublicKey().Bytes()},
		program,
	)

	signerSellToken, _, err := solana.FindAssociatedTokenAddress(
		signerTokenAuthority,
		tokenSell,
	)

	signerBuyToken, _, err := solana.FindAssociatedTokenAddress(
		signerTokenAuthority,
		tokenBuy,
	)

	recent, err := e.RpcClient.GetRecentBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return "", err
	}

	// create and mint token
	// get token amount out
	// assume amount out 0
	// swap token
	ammProgramId := solana.MustPublicKeyFromBase58("9rpQHSyFVM1dkkHFQ2TtTzPEW7DVmEyPmN8wVniqJtuC")
	swapAccounts := []*solana.AccountMeta{
		solana.NewAccountMeta(signer.PublicKey(), false, true),
		solana.NewAccountMeta(solana.TokenProgramID, false, false),
		solana.NewAccountMeta(solana.MustPublicKeyFromBase58("HeD1cekRWUNR25dcvW8c9bAHeKbr1r7qKEhv7pEegr4f"), true, false),  // amm account
		solana.NewAccountMeta(solana.MustPublicKeyFromBase58("DhVpojXMTbZMuTaCgiiaFU7U8GvEEhnYo4G9BUdiEYGh"), false, false), // amm authority
		solana.NewAccountMeta(solana.MustPublicKeyFromBase58("HboQAt9BXyejnh6SzdDNTx4WELMtRRPCr7pRSLpAW7Eq"), true, false),  // amm open orders
		solana.NewAccountMeta(solana.MustPublicKeyFromBase58("6TzAjFPVZVMjbET8vUSk35J9U2dEWFCrnbHogsejRE5h"), true, false),  // amm target order
		solana.NewAccountMeta(solana.MustPublicKeyFromBase58("3qbeXHwh9Sz4zabJxbxvYGJc57DZHrFgYMCWnaeNJENT"), true, false),  // pool_token_coin Amm
		solana.NewAccountMeta(solana.MustPublicKeyFromBase58("FrGPG5D4JZVF5ger7xSChFVFL8M9kACJckzyCz8tVowz"), true, false),  // pool_token_pc Amm
		solana.NewAccountMeta(solana.MustPublicKeyFromBase58("DESVgJVGajEgKGXhb6XmqDHGz3VjdgP7rEVESBgxmroY"), false, false), // serum dex
		solana.NewAccountMeta(solana.MustPublicKeyFromBase58("3tsrPhKrWHWMB8RiPaqNxJ8GnBhZnDqL4wcu5EAMFeBe"), true, false),  // serum market accounts
		solana.NewAccountMeta(solana.MustPublicKeyFromBase58("ANHHchetdZVZBuwKWgz8RSfVgCDsRpW9i2BNWrmG9Jh9"), true, false),  // bid account
		solana.NewAccountMeta(solana.MustPublicKeyFromBase58("ESSri17GNbVttqrp7hrjuXtxuTcCqytnrMkEqr29gMGr"), true, false),  // ask account
		solana.NewAccountMeta(solana.MustPublicKeyFromBase58("FGAW7QqNJGFyhakh5jPzGowSb8UqcSJ95ZmySeBgmVwt"), true, false),  // event q accounts
		solana.NewAccountMeta(solana.MustPublicKeyFromBase58("E1E5kQqWXkXbaqVzpY5P2EQUSi8PNAHdCnqsj3mPWSjG"), true, false),  // coin vault
		solana.NewAccountMeta(solana.MustPublicKeyFromBase58("3sj6Dsw8fr8MseXpCnvuCSczR8mQjCWNyWDC5cAfEuTq"), true, false),  // pc vault
		solana.NewAccountMeta(solana.MustPublicKeyFromBase58("C2fDkZJqHH5PXyQ7UWBNZsmu6vDXxrEbb9Ex9KF7XsAE"), false, false), // vault signer account
		solana.NewAccountMeta(signerSellToken, true, false),                                                                 // source token acc
		solana.NewAccountMeta(signerBuyToken, true, false),                                                                  // dest token acc
		solana.NewAccountMeta(signerTokenAuthority, false, false),                                                           // user owner
		solana.NewAccountMeta(ammProgramId, false, false),                                                                   // user owner
	}

	// swapbasein
	tag := byte(9)

	amountInBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(amountInBytes, amountIn)

	amountOutBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(amountOutBytes, expectAmountOut)
	swapData := append([]byte{tag}, amountInBytes...)
	swapData = append(swapData, amountOutBytes...)
	data := append([]byte{0x3}, []byte{byte(len(swapData))}...)
	data = append(data, swapData...)
	data = append(data, []byte{byte(len(swapAccounts) - 2)}...)
	data = append(data, []byte{byte(17)}...)
	fmt.Printf("data %v\n", data)
	txSwap, err := solana.NewTransaction(
		[]solana.Instruction{
			solana.NewInstruction(
				program,
				swapAccounts,
				data,
			),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(feePayer.PublicKey()),
	)
	if err != nil {
		return
	}

	txSig, err := e.SignAndSendTx(txSwap, signersSwap)
	if err != nil {
		return
	}

	txHash = txSig.String()

	return
}
func (e *Client) GetExpectedAmount(tokenIn, tokenOut string, amountIn uint64) (amountOut uint64, priceImpact float64, err error) {

	log.Println("amountIn=>", amountIn)

	poolTokenAmm := solana.MustPublicKeyFromBase58("3qbeXHwh9Sz4zabJxbxvYGJc57DZHrFgYMCWnaeNJENT")
	pcTokenAmm := solana.MustPublicKeyFromBase58("FrGPG5D4JZVF5ger7xSChFVFL8M9kACJckzyCz8tVowz")
	swapAmount := uint64(amountIn)
	poolTokenBal, err := e.RpcClient.GetTokenAccountBalance(context.Background(), poolTokenAmm, rpc.CommitmentConfirmed)
	if err != nil {
		return
	}
	pcTokenBal, err := e.RpcClient.GetTokenAccountBalance(context.Background(), pcTokenAmm, rpc.CommitmentConfirmed)
	if err != nil {
		return
	}

	fmt.Printf("pool token bal: %v \n", poolTokenBal.Value.Amount)
	fmt.Printf("pc token bal: %v \n", pcTokenBal.Value.Amount)

	ammId := solana.MustPublicKeyFromBase58("HeD1cekRWUNR25dcvW8c9bAHeKbr1r7qKEhv7pEegr4f")
	resp, err := e.RpcClient.GetAccountInfo(
		context.TODO(),
		ammId,
	)
	if err != nil {
		return
	}
	ammData := resp.Value.Data.GetBinary()
	swap_fee_numerator := ammData[22*8 : 23*8]
	swap_fee_denominator := ammData[23*8 : 24*8]
	swapFeeNum := binary.LittleEndian.Uint64(swap_fee_numerator)
	swapFeeDe := binary.LittleEndian.Uint64(swap_fee_denominator)
	// hardcode fee 0.3%
	fromAmountWithFee := new(big.Int).SetUint64(swapAmount * (swapFeeDe - swapFeeNum) / swapFeeDe)
	poolAmountBig, _ := new(big.Int).SetString(poolTokenBal.Value.Amount, 10)
	pcAmountBig, _ := new(big.Int).SetString(pcTokenBal.Value.Amount, 10)
	denominator := big.NewInt(0).Add(fromAmountWithFee, poolAmountBig)
	temp := big.NewInt(0).Mul(fromAmountWithFee, pcAmountBig)
	amountOut = big.NewInt(0).Div(temp, denominator).Uint64()
	fmt.Printf("Amount out: %v \n", amountOut)
	priceBefore := float64(pcAmountBig.Uint64()) / float64(poolAmountBig.Uint64())
	priceAfter := float64(pcAmountBig.Uint64()-amountOut) / float64(denominator.Uint64())
	priceImpact = math.Abs(priceBefore-priceAfter) * 100 / priceBefore
	fmt.Printf("price Impact: %v%%\n", priceImpact)

	return
}

// Submit proof for trade:
func (e *Client) SubmitProofTx(feePlayerPrivkey, privKey, txBurn, tokenIn string) (txHash string, err error) {

	log.Println("txBurn: ", txBurn)
	log.Println("tokenIn: ", tokenIn)

	feePayer, err := solana.PrivateKeyFromBase58(feePlayerPrivkey) // account to create tx.
	if err != nil {
		return
	}

	tokenSell := solana.MustPublicKeyFromBase58(tokenIn) // token unshield

	signer, err := solana.PrivateKeyFromBase58(privKey)

	if err != nil {
		return
	}

	signers := []solana.PrivateKey{
		signer,
		feePayer,
	}

	incognitoProxy, err := solana.PublicKeyFromBase58(e.IncognitoProxy)
	if err != nil {
		return
	}

	program := solana.MustPublicKeyFromBase58(e.ProgramID)

	log.Println("e.ProgramID", e.ProgramID)

	signerTokenAuthority, _, err := solana.FindProgramAddress(
		[][]byte{signer.PublicKey().Bytes()},
		program,
	)

	log.Println("signerTokenAuthority: ", signerTokenAuthority)

	signerSellToken, _, err := solana.FindAssociatedTokenAddress(
		signerTokenAuthority,
		tokenSell,
	)

	log.Println("signerSellToken", signerSellToken) // need to create vault account

	vaultAcc := solana.MustPublicKeyFromBase58(VAULT_ACC)

	vaultTokenAuthority, _, err := solana.FindProgramAddress(
		[][]byte{incognitoProxy.Bytes()},
		program,
	)
	if err != nil {
		return
	}

	//
	vaultAssTokenAcc, _, err := solana.FindAssociatedTokenAddress(vaultTokenAuthority, tokenSell)
	if err != nil {
		panic(err)
	}
	fmt.Println("vaultAssTokenAcc:", vaultAssTokenAcc.String())

	recent, err := e.RpcClient.GetRecentBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return "", err
	}

	burnProofdAccounts := []*solana.AccountMeta{
		solana.NewAccountMeta(vaultAssTokenAcc, true, false),
		solana.NewAccountMeta(signer.PublicKey(), false, false),
		solana.NewAccountMeta(vaultTokenAuthority, false, false),
		solana.NewAccountMeta(vaultAcc, true, false),
		solana.NewAccountMeta(incognitoProxy, false, false),
		solana.NewAccountMeta(solana.TokenProgramID, false, false),
		solana.NewAccountMeta(signerSellToken, true, false),
	}

	unshield := unshield.NewUnshield(
		txBurn,
		"getburnsolprooffordeposittosc", // update
		"https://fullnode.solana-bridge-demo.incognito.org",
		program,
		burnProofdAccounts,
	)

	unshieldInst := []solana.Instruction{}

	unshieldInst = append(unshieldInst, unshield.Build())

	tx, err := solana.NewTransaction(
		unshieldInst,
		recent.Value.Blockhash,
		solana.TransactionPayer(feePayer.PublicKey()),
	)
	sig, err := e.SignAndSendTx(tx, signers)
	if err != nil {
		return
	}
	txHash = sig.String()

	return
}

func (e *Client) WithdrawRaydium(feePlayerPrivkey, privKey, incAddress, tokenOut string, amount uint64) (txHash string, err error) {

	log.Println("incAddress", incAddress, "shieldTokenAddress", "amount", amount)

	feePayer, err := solana.PrivateKeyFromBase58(feePlayerPrivkey) // account to create tx.
	if err != nil {
		return
	}

	tokenBuy := solana.MustPublicKeyFromBase58(tokenOut) //

	signer, err := solana.PrivateKeyFromBase58(privKey)

	if err != nil {
		return
	}

	incognitoProxy, err := solana.PublicKeyFromBase58(e.IncognitoProxy)
	if err != nil {
		return
	}

	program := solana.MustPublicKeyFromBase58(e.ProgramID)

	signerTokenAuthority, _, err := solana.FindProgramAddress(
		[][]byte{signer.PublicKey().Bytes()},
		program,
	)

	signerBuyToken, _, err := solana.FindAssociatedTokenAddress(
		signerTokenAuthority,
		tokenBuy,
	)
	vaultTokenAuthority, _, err := solana.FindProgramAddress(
		[][]byte{incognitoProxy.Bytes()},
		program,
	)

	vaultBuyToken, _, err := solana.FindAssociatedTokenAddress(
		vaultTokenAuthority,
		tokenBuy,
	)

	log.Println("vaultBuyToken", vaultBuyToken)

	shieldAccounts := []*solana.AccountMeta{
		solana.NewAccountMeta(signerBuyToken, true, false),
		solana.NewAccountMeta(vaultBuyToken, true, false),
		solana.NewAccountMeta(incognitoProxy, false, false),
		solana.NewAccountMeta(signer.PublicKey(), false, true),
		solana.NewAccountMeta(signerTokenAuthority, false, false),
		solana.NewAccountMeta(solana.TokenProgramID, false, false),
	}

	signers := []solana.PrivateKey{
		feePayer,
		signer,
	}

	log.Println("tokenBuy", tokenBuy)

	depositAmount, err := e.RpcClient.GetTokenAccountBalance(context.Background(), signerBuyToken, rpc.CommitmentConfirmed)
	if err != nil {
		return
	}
	fmt.Printf("depositAmount token bal: %v \n", depositAmount.Value.Amount)

	amountUint, _ := strconv.ParseUint(depositAmount.Value.Amount, 10, 64)

	shieldInstruction := shield.NewShield(
		incAddress,
		amountUint,
		program,
		shieldAccounts,
		byte(4),
	)

	// account to create tx.
	recent, err := e.RpcClient.GetRecentBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return "", err
	}

	shieldInsGenesis := shieldInstruction.Build()
	if shieldInsGenesis == nil {
		return
	}
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			shieldInsGenesis,
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(feePayer.PublicKey()),
	)
	if err != nil {
		return
	}
	sig, err := e.SignAndSendTx(tx, signers)
	if err != nil {
		return
	}
	txHash = sig.String()

	return
}
