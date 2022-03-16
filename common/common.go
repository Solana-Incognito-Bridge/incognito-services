package common

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/incognito-core-libs/btcutil"
)

func RandInt() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Int()
}

func IndexStringAtArray(str string, strs []string) int {
	for k, v := range strs {
		if str == v {
			return k
		}
	}
	return -1
}

func RemoveStringAtArray(str string, strs []string) []string {
	rets := make([]string, 0)
	for idxT, strT := range strs {
		if str == strT {
			rets = append(rets, strs[0:idxT]...)
			rets = append(rets, strs[idxT+1:]...)
			return rets
		}
	}
	rets = append(rets, strs...)
	return strs
}

func RemoveIndexAtArray(idx int, strs []string) []string {
	rets := make([]string, 0)
	for idxT, _ := range strs {
		if idx == idxT {
			rets = append(rets, strs[0:idxT]...)
			rets = append(rets, strs[idxT+1:]...)
			return rets
		}
	}
	rets = append(rets, strs...)
	return strs
}

func ConvertEtherFromWei(val *big.Int) float64 {
	return float64(new(big.Int).Div(val, big.NewInt(1e15)).Int64()) * 1e-3
}

func ConvertEtherFromAda(val uint64) float64 {
	return float64(val) * 1e-3
}

func ConvertWeiFromEther(val float64) *big.Int {
	return new(big.Int).Mul(big.NewInt(int64(val*1e3)), big.NewInt(1e15))
}

func ConvertSatoshiFromBTC(val float64) *big.Float {
	amount, err := btcutil.NewAmount(val)
	if err == nil {
		return big.NewFloat(amount.ToUnit(btcutil.AmountSatoshi))
	}
	return big.NewFloat(0)
}

func ConvertBTCFromSatoshi(val float64) float64 {
	amount := btcutil.Amount(val)
	return amount.ToBTC()
}

// InterfaceSlice receives a slice which is a interface
// and converts it into slice of interface
func InterfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		log.Println("InterfaceSlice() given a non-slice type")
		return nil
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}

// ex: PNetworkName: DAI, PTokenName: pDAI
func GenPrivacyTokenID(PNetworkName string, PTokenName string) string {

	if PNetworkName == "" {
		log.Println("Wrong param")
		return ""
	}
	if PTokenName == "" {
		log.Println("Wrong param")
		return ""
	}
	tokenID := Hash{}

	hashPNetWork := HashH([]byte(PNetworkName))
	log.Printf("hashPNetWork: %+v\n", hashPNetWork.String())
	copy(tokenID[:16], hashPNetWork[:16])
	log.Printf("tokenID: %+v\n", tokenID.String())

	hashPToken := HashH([]byte(PTokenName))
	log.Printf("hashPToken: %+v\n", hashPToken.String())
	copy(tokenID[16:], hashPToken[:16])

	log.Printf("Result tokenID: %+v\n", tokenID.String())

	return tokenID.String()
}

func ConvertToDecimal(amount *big.Float, decimal int64) *big.Float {
	coin := big.NewFloat(math.Pow10(int(decimal)))
	return new(big.Float).Quo(amount, coin)
}

func ConvertCoinToIncNanoTokenString(amountStr string, decimal, pDecimals int64) string {

	if decimal == pDecimals {
		return amountStr
	}

	amount := new(big.Float)
	amount, err := amount.SetString(amountStr)

	if !err {
		fmt.Println("err", err)
		return "0"
	}

	value1 := ConvertToDecimal(amount, decimal)
	fmt.Println("nano coin: ", value1.String())

	value2 := ConvertToNanoIncognitoToken(value1, pDecimals)
	fmt.Println(value2.String())
	return value2.String()
}

//

func ConvertCoinToIncNanoTokenUint64(amountStr string, decimal, pDecimals int64) uint64 {
	amount := new(big.Float)
	amount, err := amount.SetString(amountStr)

	if !err {
		fmt.Println("err", err)
		return 0
	}

	value1 := ConvertToDecimal(amount, decimal)
	fmt.Println("nano coin: ", value1.String())

	return ConvertToNanoIncognitoTokenUint(value1, pDecimals)
}

func ConvertToNanoIncognitoTokenUint(amountDecimal *big.Float, pDecimals int64) uint64 {

	g := amountDecimal.Mul(amountDecimal, big.NewFloat(math.Pow10(int(pDecimals))))

	ing, _ := g.Uint64()

	return ing
}

//
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

func InArray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}

func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func JoinAddress(address1 string, address2 string) string {
	if address2 != "" {
		return fmt.Sprintf("%s %s", address1, address2)
	}
	return address1
}

func GetBody(req *http.Request) ([]byte, error) {
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		log.Println(err)
		return make([]byte, 0), err
	}
	if resp.StatusCode != http.StatusOK {
		bytes, _ := ioutil.ReadAll(resp.Body)
		return bytes, errors.New("HttpResponse Status Code = " + strconv.Itoa(resp.StatusCode))
	}
	return ioutil.ReadAll(resp.Body)
}

func Round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return newVal
}

func MinBigint(x, y *big.Int) *big.Int {
	if r := x.Cmp(y); r < 1 {
		return x
	}
	return y
}

func MinInt64(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

// erc20:
// convert coin nano amout to coin amount: ex: 2000000000000000(nano ETH)/1e18=0.002 ETH
func ConvertNanoCoinToCoin(nanoAmount *big.Float, decimal int64) *big.Float {
	value := big.NewFloat(math.Pow10(int(decimal)))
	return new(big.Float).Quo(nanoAmount, value)
}

// convert coin amount to incognito nano token amount: ex: 002(ETH)*1e9=2000000 nano pETH
func ConvertToNanoIncognitoToken(coinAmount *big.Float, pdecimal int64) *big.Float {
	value := big.NewFloat(math.Pow10(int(pdecimal)))
	return new(big.Float).Mul(coinAmount, value)
}

// convert nano coin to nano token: ex: 2000000000000000 (nano eth) => 2000000 (nano pETH)
func ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(amountStr string, decimal, pDecimals int64) string {

	if decimal == pDecimals {
		return amountStr
	}
	amount := new(big.Float)
	amount, err := amount.SetString(amountStr)

	if !err {
		fmt.Println("err", err)
		return "0"
	}

	value1 := ConvertNanoCoinToCoin(amount, decimal)
	fmt.Println("nano coin: ", value1.String())
	value2 := ConvertToNanoIncognitoToken(value1, pDecimals)
	fmt.Println("nano ptoken: ", value2)

	value2Int64, _ := value2.Uint64()
	fmt.Println("value2Int: ", value2Int64)

	amountBigInt := new(big.Int)
	amountBigInt.SetUint64(value2Int64)

	return amountBigInt.String()
}
