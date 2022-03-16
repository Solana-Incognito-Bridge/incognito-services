package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type UserGender int

const (
	InvalidGender = iota
	Male
	Female
)

type UserType int

const (
	InvalidUserType UserType = iota
	Borrower
	Lender
)

func (u UserType) String() string {
	return [...]string{"invalid", "borrower", "lender"}[u]
}

type UserRole int

const (
	UserRoleInvalid UserRole = iota
	UserRoleNormal
	UserRoleAdmin
	UserRoleRoot
)

type User struct {
	gorm.Model

	FirstName  string
	MiddleName string
	LastName   string
	FullName   string
	UserName   string
	Email      string
	Password   string
	Bio        string `gorm:"type:text"`

	Address string

	IsActive        bool
	IsVerifiedEmail bool
	Role            UserRole
	DeviceToken     string
	IP              string
}

func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

func (u *User) DisplayName() string {
	return u.FirstName + " " + u.LastName
}

type UserVerificationType int

const (
	UserVerificationTypeEmail UserVerificationType = iota
	UserVerificationTypeForgotPassword
)

type UserVerification struct {
	gorm.Model

	User   *User
	UserID int

	Type      UserVerificationType
	Token     string
	IsValid   bool
	ExpiredAt time.Time
}

type BlackListIP struct {
	gorm.Model
	IP   string
	Note string
}

type IncognitoConfig struct {
	gorm.Model
	MinerPrice           float64
	MinerShipInfo        string
	DisableDecentralized bool
	DisableCentralized   bool
	// StakingPoolIsAutoStake bool   `gorm:"default:false"`
	StakingPoolRewardRate   int    `gorm:"default:45"`
	StakingPoolCompoundRate int    `gorm:"default:57"`
	StakingPoolMinToStake   uint64 `gorm:"default:10000000000"`

	StakingPoolConfig string `gorm:"type:text"`
	/*
		[
			{
				"ID": 0,
				"Min": 0.01,
				"Max": 100000,
				"APR": 32,
				"APY": 57
			},
			{
				"ID": 1,
				"Min": 0.01,
				"Max": 100000,
				"APR": 32,
				"APY": -1,
			},
		]

	*/

	BuyNodePTokensSupport string
	BuyNodePTokensPartner string
	DappPartner           string

	RunJob string `sql:"type:LONGTEXT"`
}
