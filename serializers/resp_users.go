package serializers

import (
	"github.com/incognito-services/models"
)

type Resp struct {
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
}

type UserResp struct {
	ID        uint            `json:"ID"`
	UserName  string          `json:"UserName"`
	FirstName string          `json:"FirstName"`
	LastName  string          `json:"LastName"`
	Email     string          `json:"Email"`
	Address   string          `json:"Address"`
	Bio       string          `json:"Bio"`
	Role      models.UserRole `json:"RoleID"`
}

func NewUserResp(data models.User) *UserResp {
	result := UserResp{
		Address:   data.Address,
		ID:        data.ID,
		Email:     data.Email,
		FirstName: data.FirstName,
		LastName:  data.LastName,
		UserName:  data.UserName,
		Bio:       data.Bio,
		Role:      data.Role,
	}
	return &result
}

type UserRegisterResp struct {
	Message         string `json:"Message,omitempty"`
	EncryptedString string `json:"EncryptedString,omitempty"`
}

type UserLoginResp struct {
	Token   string `json:"Token"`
	Expired string `json:"Expired"`
}

type MessageResp struct {
	Message string `json:"Message"`
}
