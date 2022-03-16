package dao

import (
	"github.com/incognito-services/models"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type User struct {
	db *gorm.DB
}

func NewUser(db *gorm.DB) *User {
	return &User{db}
}

func (u *User) Create(user *models.User) error {
	return errors.Wrap(u.db.Create(user).Error, "u.db.Create")
}

func (u *User) GetIOSUsers() ([]models.User, error) {
	var users []models.User
	if err := u.db.Where("LENGTH(email) = 50").Find(&users).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "u.db.Where.First")
	}
	return users, nil
}

func (u *User) GetAndroidUsers() ([]models.User, error) {
	var users []models.User
	if err := u.db.Where("LENGTH(email) <> 50").Find(&users).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "u.db.Where.First")
	}
	return users, nil
}

func (u *User) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := u.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "u.db.Where.First")
	}
	return &user, nil
}

func (u *User) FindByAddress(address string) (*models.User, error) {
	var user models.User
	if err := u.db.Where("address = ?", address).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "u.db.Where.First")
	}
	return &user, nil
}

func (u *User) FindByID(id uint) (*models.User, error) {
	var user models.User
	if err := u.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "u.db.Where.First")
	}
	return &user, nil
}

func (u *User) Update(user *models.User) error {
	return errors.Wrap(u.db.Save(user).Error, "u.db.save")
}

func (u *User) ResetPassword(r *models.UserVerification) (err error) {
	tx := u.db.Begin()
	if tErr := tx.Error; tErr != nil {
		err = errors.Wrap(tErr, "tx.Error")
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = errors.Wrap(tx.Commit().Error, "tx.Commit")
	}()

	if tErr := tx.Save(r.User).Error; tErr != nil {
		err = errors.Wrap(tErr, "tx.Save")
		return
	}
	if tErr := tx.Save(r).Error; tErr != nil {
		err = errors.Wrap(tErr, "tx.Save")
		return
	}
	return
}

func (u *User) VerifyEmail(r *models.UserVerification) (err error) {
	tx := u.db.Begin()
	if tErr := tx.Error; tErr != nil {
		err = errors.Wrap(tErr, "tx.Error")
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = errors.Wrap(tx.Commit().Error, "tx.Commit")
	}()

	if tErr := tx.Save(r.User).Error; tErr != nil {
		err = errors.Wrap(tErr, "tx.Save")
		return
	}
	if tErr := tx.Save(r).Error; tErr != nil {
		err = errors.Wrap(tErr, "tx.Save")
		return
	}
	return
}

func (u *User) GetPaymentAddressByPublicKey(pubkey string) (string, error) {
	var ret struct {
		PaymentAddress string
	}
	if err := u.db.Table("users").Where("pubkey = ?", pubkey).Select("payment_address").Scan(&ret).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", errors.Wrap(err, "u.db.Table")
	}
	return ret.PaymentAddress, nil
}

// BlackListIP
// create airdrop:
func (u *User) InsertToBlackListIP(blackListIP *models.BlackListIP) error {

	var ip models.BlackListIP
	if err := u.db.Where("ip=?", blackListIP.IP).First(&ip).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.Wrap(u.db.Create(blackListIP).Error, "u.db.Create")
		}
		return errors.Wrap(err, "u.db.Where.First")
	}
	return nil
}

func (u *User) GetBlackListIP() ([]*models.BlackListIP, error) {
	var blackListIP []*models.BlackListIP
	if err := u.db.Find(&blackListIP).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return blackListIP, nil
		}
		return blackListIP, errors.Wrap(err, "u.db.Where.First")
	}
	return blackListIP, nil
}
func (u *User) GetBlackListIPByIP(ipAddress string) (*models.BlackListIP, error) {
	var ip models.BlackListIP
	if err := u.db.Where("ip=?", ipAddress).First(&ip).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "u.db.Where.First")
	}
	return &ip, nil
}

func (u *User) GetIncognitoConfigs() (*models.IncognitoConfig, error) {
	var configs models.IncognitoConfig
	if err := u.db.First(&configs).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "u.db.Where.First")
	}
	return &configs, nil
}
