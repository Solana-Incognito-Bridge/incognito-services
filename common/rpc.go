package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Payload struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type Response struct {
	Error  interface{} `json:"error"`
	Result interface{} `json:"result"`
}

func RequestHTTP(url string, method string, params []interface{}) (interface{}, error) {
	// Initialize a JSON body for requesting
	data := Payload{
		"2.0",
		method,
		params,
		1,
	}
	reqBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(reqBytes)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode and return the result
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request return status %v", resp.StatusCode)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	res := new(Response)
	err = json.Unmarshal(bodyBytes, res)
	if err != nil {
		return nil, err
	}
	if res.Error != nil {
		errorInterface := res.Error.(map[string]interface{})
		return nil, fmt.Errorf("HTTP request error code %v", errorInterface["Code"])
	}
	return res.Result, nil
}
