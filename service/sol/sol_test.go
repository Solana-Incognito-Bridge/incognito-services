package sol

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const INCOGNITO_PROXY = "5Tq3wvYAD6hRonCiUx62k37gELxxEABSYCkaqrSP3ztv"
const PROGRAM_ID = "BKGhwbiTHdUxcuWzZtDWyioRBieDEXTtgEk8u1zskZnk"
const PRIKEY_FEE = "588FU4PktJWfGfxtzpAAXywSNt74AvtroVzGfKkVN1LwRuvHwKGr851uH8czM5qm4iqLbs1kKoMKtMJG4ATR7Ld2"

type SolTestSuite struct {
	suite.Suite
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *SolTestSuite) SetupTest() {

}

func TestSolTestSuite(t *testing.T) {
	suite.Run(t, new(SolTestSuite))
}

// //Get balance, create wallet, transfer.
// func (suite *SolTestSuite) TestCreateWallet() {

// 	client := NewClient(INCOGNITO_PROXY, PROGRAM_ID)
// 	privKey, pubKey, address, err := client.GenerateNativeAddress()

// 	log.Println("privKey, pubKey, address, err: ", privKey, pubKey, address, err)

// 	assert.NotEmpty(suite.T(), "")
// 	// assert.Equal(suite.T(), true, balance.Uint64() > 0)
// }

// func (suite *SolTestSuite) TestCreateTokenWallet() {

// 	client := NewClient(INCOGNITO_PROXY, PROGRAM_ID, true)
// 	privKey, pubKey, address, err := client.GenerateNativeAddress()

// 	log.Println("privKey, pubKey, address, err: ", privKey, pubKey, address, err)

// 	feePlayerPrivkey := PRIKEY_FEE
// 	shieldMakerPrivateAddress := privKey
// 	tokenID := "7gSh4k2jhNJtCHfSbtaScYW4NFnuKkB9x84f9KCECXmf"

// 	privKey, pubKey, address, tx, err := client.GenerateTokenAddress(feePlayerPrivkey, shieldMakerPrivateAddress, tokenID)

// 	log.Println("privKey, pubKey, address, tx, err: ", privKey, pubKey, address, tx, err)

// 	assert.NotEmpty(suite.T(), "")
// 	// assert.Equal(suite.T(), true, balance.Uint64() > 0)
// }

// func (suite *SolTestSuite) TestGetStatusTx() {
// 	tx := "KxbWvpeNgJrjiPFkRRFp5NVsfn1kAdF4fQbFKipcM7VDjiteGx4b3W79GVnX7ScTMAie7viFJ5ULL6Usyh78Ett"
// 	client := NewClient(INCOGNITO_PROXY, PROGRAM_ID)
// 	status, err := client.CheckTxStatus(tx)
// 	log.Println("status, err", status, err)
// 	assert.NotEmpty(suite.T(), "")
// }

func (suite *SolTestSuite) TestGetVault() {
	client := NewClient(INCOGNITO_PROXY, PROGRAM_ID, true)
	vaultAddress, txHash, err := client.CreateVaultAddress(PRIKEY_FEE, "7gSh4k2jhNJtCHfSbtaScYW4NFnuKkB9x84f9KCECXmf")
	log.Println("vaultAddress, txHash, err", vaultAddress, txHash, err)
	assert.NotEmpty(suite.T(), "")
}

func (suite *SolTestSuite) TestGetVaultNativeWrapForAccount() {
	client := NewClient(INCOGNITO_PROXY, PROGRAM_ID, true)
	vaultAddress, txHash, err := client.GetVaultNativeWrapForAccount(PRIKEY_FEE, "")
	log.Println("vaultAddress, txHash, err", vaultAddress, txHash, err)

	if err != nil {
		log.Println("err==>", err)
	}
	assert.NotEmpty(suite.T(), "")
}

func (suite *SolTestSuite) TestShieldNative() {
	client := NewClient(INCOGNITO_PROXY, PROGRAM_ID, true)
	txHash, err := client.ShieldNative("3qyM4EdgNhoEbbxvudVZQzK6PHgnyxDMXQ9f44WwvsTvQWQgCRCy3t2uWVeAJdgiCd4aJTc4aHAQhZAKSqZbY3j1",
		"12shR6fDe7ZcprYn6rjLwiLcL7oJRiek66ozzYu3B3rBxYXkqJeZYj6ZWeYy4qR4UHgaztdGYQ9TgHEueRXN7VExNRGB5t4auo3jTgXVBiLJmnTL5LzqmTXezhwmQvyrRjCbED5xW7yMMeeWarKa",
		"HCekuqhnuFUq1RqJ4szQg9ikXhmxXvDGn5ubzyxNTgjY",
		1000000000)
	log.Println(" txHash, err", txHash, err)
	assert.NotEmpty(suite.T(), "")
}

func (suite *SolTestSuite) TestShieldToken() {
	client := NewClient(INCOGNITO_PROXY, PROGRAM_ID, true)
	txHash, err := client.ShieldToken(
		"4UvoaU3PodSRhbrQCZykcZxc7rzkP3iiDeRvSRaU7MzNTYXG8g6ACg7jPAngatZtkce93BSTD7hegjywjucjHKuN", // privKey account
		"12smSBNjponk5wQo7XrnJiALZvePp6SXMQtPbF9kFppJw2ztfVsrhAyPYFKc6ydiMs4kAk3mdwCpnNW79KkwWevEFDLMH9Djiy5xhz2SFx5VJPsRwrETs4NkKjD1edLiChAEpLGoLo3M8EiFzfAG",
		"JD4tdLcVySwoibHT7qXSevthgse13XYVQh6xzKE2uPGx", // tokenAcc
		"37bf7L1u1sTHXETcHb8fz9r6dA1xWwvLGuTh5ocJbdRP",
		50000000)
	log.Println(" txHash, err", txHash, err)
	assert.NotEmpty(suite.T(), "")
}

func (suite *SolTestSuite) TestGetAddress() {
	client := NewClient(INCOGNITO_PROXY, PROGRAM_ID, true)
	client.GetAddress()
}
