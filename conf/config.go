package config

import (
	"encoding/json"
	"log"
	"os"
)

var config *Config

func init() {
	file, err := os.Open("conf/conf.json")
	if err != nil {
		log.Println("error:", err)
		panic(err)
	}
	decoder := json.NewDecoder(file)
	v := Config{}
	err = decoder.Decode(&v)
	if err != nil {
		log.Println("error:", err)
		panic(err)
	}
	config = &v
}

func GetConfig() *Config {
	return config
}

// FirebaseConfig : struct
type FirebaseConfig struct {
	APIKey         string `json:"api_key"`
	Database       string `json:"database"`
	TopicSendUsers string `json:"topic_send_users"`
}

//  : struct
type SlackConfig struct {
	OAuthToken       string `json:"api_token"`
	Channel          string `json:"channel"`
	PoolStakeChannel string `json:"pool_stake_channel"`

	PoolRequestStakeChannel    string `json:"pool_request_stake_channel"`
	PoolRequestWithdrawChannel string `json:"pool_request_withdraw_channel"`
	PoolRequestUnStakeChannel  string `json:"pool_request_unstake_channel"`

	PoolIssue string `json:"pool_issue"`
	PoolAlert string `json:"pool_alert"`
}

type IncognitoConfig struct {
	ChainEndpoint   string `json:"chain_endpoint"`
	NetWorkEndPoint string `json:"network_endpoint"`

	Password string `json:"password"`
	Username string `json:"username"`

	Fee int `json:"fee"`

	ProxyAddress   string `json:"proxy_address"`
	VaultAddress   string `json:"vault_address"`
	BurningAddress string `json:"burning_address"`
	ErrorCount     int    `json:"error_count"`
}

type SendgridConfig struct {
	APIKey string `json:"api_key"`
}

type Config struct {
	Port    int    `json:"port"`
	Env     string `json:"env"`
	BaseURL string `json:"base_url"`
	Db      string `json:"db"`

	// 0. Incognito chain config:
	Incognito IncognitoConfig `json:"incognito"`

	TokenSecretKey string `json:"token_secret_key"`

	// Slack:
	Slack SlackConfig `json:"slack"`

	// Firebase:
	Firebase FirebaseConfig `json:"firebase"`

	StakeCommission float64 `json:"stake_commission"`

	SentryDSN string `json:"sentry_dsn"`

	Sendgrid SendgridConfig `json:"sendgrid"`
}
