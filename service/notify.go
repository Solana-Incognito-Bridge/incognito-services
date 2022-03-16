package service

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/incognito-services/service/3rd/slack"
)

func BigIntFloat(balance *big.Int, decimals int64) *big.Float {
	return bigIntFloat(balance, decimals)
}
func BigIntString(balance *big.Int, decimals int64) string {
	return bigIntString(balance, decimals)
}

func bigIntString(balance *big.Int, decimals int64) string {
	amount := bigIntFloat(balance, decimals)
	deci := fmt.Sprintf("%%0.%vf", decimals)
	return clean(fmt.Sprintf(deci, amount))
}

func bigIntFloat(balance *big.Int, decimals int64) *big.Float {
	if balance.Sign() == 0 {
		return big.NewFloat(0)
	}
	bal := big.NewFloat(0)
	bal.SetInt(balance)
	pow := bigPow(10, decimals)
	p := big.NewFloat(0)
	p.SetInt(pow)
	bal.Quo(bal, p)
	return bal
}

func bigPow(a, b int64) *big.Int {
	r := big.NewInt(a)
	return r.Exp(r, big.NewInt(b), nil)
}

func clean(newNum string) string {
	stringBytes := bytes.TrimRight([]byte(newNum), "0")
	newNum = string(stringBytes)
	if stringBytes[len(stringBytes)-1] == 46 {
		newNum += "0"
	}
	if stringBytes[0] == 46 {
		newNum = "0" + newNum
	}
	return newNum
}

func notifySlack(msg, slackChanel, oAuthToken string) {
	fmt.Println(msg)
	s := slack.InitSlack(oAuthToken, slackChanel)
	go s.PostMsg(msg)
}
