package serializers

type UserReq struct {
	FirstName       string `json:"FirstName"`
	LastName        string `json:"LastName"`
	FullName        string `json:"FullName"`
	Email           string `json:"Email"`
	Password        string `json:"Password"`
	ConfirmPassword string `json:"ConfirmPassword"`
	UserName        string `json:"UserName"`
	Bio             string `json:"bio"`
	PaymentAddress  string `json:"PaymentAddress"`
	IP              string
}