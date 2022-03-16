package serializers

type AuthReq struct {
	Address  string `json:"Address"`
	Password string `json:"Password"`
	IP       string
}

type UserReq struct {
	FirstName       string `json:"FirstName"`
	LastName        string `json:"LastName"`
	FullName        string `json:"FullName"`
	Email           string `json:"Email"`
	Password        string `json:"Password"`
	ConfirmPassword string `json:"ConfirmPassword"`
	UserName        string `json:"UserName"`
	Bio             string `json:"Bio"`
	Address         string `json:"Address"`
	IP              string
}

type UserResetPasswordReq struct {
	Token              string `json:"Token" binding:"required"`
	NewPassword        string `json:"NewPassword" binding:"required"`
	ConfirmNewPassword string `json:"ConfirmNewPassword" binding:"required"`
}

type UserVerificationEmail struct {
	Token string `json:"Token"`
}

type AirdropReq struct {
	WalletAddress     string `json:"WalletAddress" binding:"required"`
	PDexWalletAddress string `json:"PDexWalletAddress" binding:"required"`
	IP                string
}

type UserSignupReq struct {
	FullName string `json:"FullName"`
	Email    string `json:"Email"`
	IP       string
}
