package service

import (
	"math/rand"
	"time"

	"github.com/inc-backend/crypto-libs/incognito/entity"

	"github.com/incognito-services/dao"
	"github.com/incognito-services/models"
	"github.com/incognito-services/serializers"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/inc-backend/crypto-libs/bnb"
	"github.com/inc-backend/crypto-libs/btc"
	"github.com/inc-backend/crypto-libs/incognito"
	config "github.com/incognito-services/conf"
	emailHelper "github.com/incognito-services/service/email"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	tokenLength                      = 10
	letters                          = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	verificationTokenExpiredDuration = 24 * time.Hour
)

type User struct {
	conf               *config.Config
	r                  *dao.User
	stakeDao           *dao.StakingPool
	bc                 *incognito.Blockchain
	btcClient          *btc.BlockcypherService
	bnbClient          *bnb.BinanceService
	mailer             *emailHelper.Email
	logger             *zap.Logger
	listTokenChain     []entity.PCustomToken
	timeListChainToken time.Time
}

func NewUserService(conf *config.Config, r *dao.User, stakeDao *dao.StakingPool, bc *incognito.Blockchain, mailer *emailHelper.Email, logger *zap.Logger) *User {

	return &User{
		conf:     conf,
		r:        r,
		stakeDao: stakeDao,
		bc:       bc,
		mailer:   mailer,
		logger:   logger,
	}
}

func (u *User) validate(firstName, lastName, email, password, confirmPassword string) error {
	if email == "" {
		return ErrInvalidEmail
	}
	if password == "" || confirmPassword == "" {
		return ErrInvalidPassword
	}
	if password != confirmPassword {
		return ErrPasswordMismatch
	}
	return nil
}

func (u *User) RegisterUserByAddress(req *serializers.AuthReq) (*models.User, error) {
	emailFromAddress := req.Address + "@incognito.org"

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, "bcrypt.GenerateFromPassword")
	}

	user, err := u.r.FindByEmail(emailFromAddress)

	if user != nil {

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			return nil, ErrInvalidPassword
		}
		return user, nil
	}
	// create user from address:
	user = &models.User{Email: emailFromAddress, Address: req.Address, Password: string(hashed), IsActive: true}

	if err := u.r.Create(user); err != nil {
		return nil, errors.Wrap(err, "u.Dao.Create")
	}
	return user, nil
}

func (u *User) FindByID(id uint) (*models.User, error) {
	user, err := u.r.FindByID(id)
	if err != nil {
		return nil, errors.Wrap(err, "u.portalDao.FindByID")
	}
	return user, nil
}

func (u *User) Authenticate(email, password string) (*serializers.UserResp, error) {
	user, err := u.r.FindByEmail(email)
	if err != nil {
		return nil, errors.Wrap(err, "u.portalDao.FindByEmail")
	}
	if user == nil {
		return nil, ErrEmailNotExists
	}

	if user.Password != "" {

		// if !user.IsActive {
		// 	return nil, ErrInactiveAccount
		// }
		// if !user.IsVerifiedEmail {
		// 	return nil, ErrEmailIsNotVerified
		// }

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			return nil, ErrInvalidPassword
		}
	}
	return assembleUser(user), nil
}

func (u *User) generateVerificationToken() string {
	b := make([]byte, tokenLength)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (u *User) Update(user *models.User, req *serializers.UserReq) error {
	if req.UserName != "" {
		user.UserName = req.UserName
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if err := u.r.Update(user); err != nil {
		return errors.Wrap(err, "u.r.Update")
	}
	return nil
}

func assembleUser(u *models.User) *serializers.UserResp {
	return &serializers.UserResp{
		ID:        u.ID,
		UserName:  u.UserName,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Address:   u.Address,
	}
}
