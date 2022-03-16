package dao

import (
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/database"
	"github.com/orgs/Solana-Incognito-Bridge/ognito-service/models"
	"github.com/pkg/errors"
)

type Shield struct {
	db *gorm.DB
}

func NewShield(db *gorm.DB) *Shield {
	return &Shield{db}
}

func (o *Shield) RunInTransaction(fn func(*gorm.DB) error) (err error) {
	tx := o.db.Begin()
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

	return fn(tx)
}

func (u *Shield) CreateSolFeeHistory(feeHistory *models.SolFeeHistory) error {
	return errors.Wrap(u.db.Create(feeHistory).Error, "o.db.CreateSolFeeHistory")
}

func (o *Shield) GetLatestShieldByTokenID(status models.ShieldStatus, incTokenID string) (*models.Shield, error) {
	var result models.Shield
	if err := o.db.Where("status = ? and privacy_token_address=?", status, incTokenID).Order("id DESC ").First(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "o.db.Where.First")
	}
	return &result, nil
}
func (o *Shield) GetLatestShieldByStatus(status models.ShieldStatus) (*models.Shield, error) {
	var result models.Shield
	if err := o.db.Where("status = ?", status).Order("id DESC ").First(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "o.db.Where.First")
	}
	return &result, nil
}

func (u *Shield) GetUserByTokenDevice(userID uint, token_ids []string) (*models.User, error) {
	var user models.User
	if err := u.db.Where("device_token in (?) and id = ?", token_ids, userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "u.db.Where.First")
	}
	return &user, nil
}

func (o *Shield) GetByIncognitoTx(tx string) (*models.Shield, error) {
	var result models.Shield

	queryString := "select * from shield_unshield_ftm where LOWER(incognito_tx)=? and address_type=1 limit 1"

	if err := o.db.Raw(queryString, strings.ToLower(tx)).Scan(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "o.db.CheckTxRaydiumOrShield")
	}
	return &result, nil
}

func (o *Shield) Create(data *models.Shield) error {
	return errors.Wrap(o.db.Create(data).Error, "u.db.Create")
}

func (o *Shield) UpdateAddress(model *models.Shield) error {
	if err := o.db.Save(model).Error; err != nil {
		return errors.Wrap(err, "tx.UpdateAddress")
	}
	return nil
}

func (o *Shield) UpdateShieldByReplaceByFeesStatus(ids []uint, referenceId uint) error {
	var shield models.Shield
	if err := o.db.Model(&shield).Where("id in (?)", ids).Update(map[string]interface{}{
		"status":       models.ReplacedByFee,
		"reference_id": referenceId,
	}).Error; err != nil {
		return errors.Wrap(err, "tx.Model")
	}
	return nil

}

func (o *Shield) ListShieldByStatusAddress(status []models.ShieldStatus, address string) ([]*models.Shield, error) {
	var results []*models.Shield
	if err := o.db.Where("status in (?) and address=?", status, address).Find(&results).Error; err != nil {
		return nil, errors.Wrap(err, "c.db.Where")
	}
	return results, nil
}

func (o *Shield) ListShieldByStatus(status []models.ShieldStatus) ([]*models.Shield, error) {
	var results []*models.Shield
	if err := o.db.Where("status in (?) and err_count <= 122", status).Find(&results).Limit(20).Error; err != nil {
		return nil, errors.Wrap(err, "c.db.Where")
	}
	return results, nil
}

func (o *Shield) ListShieldMintToken(status []models.ShieldStatus) ([]*models.Shield, error) {
	var results []*models.Shield
	if err := o.db.Where("status in (?) and err_count <= 122", status).Find(&results).Limit(20).Error; err != nil {
		return nil, errors.Wrap(err, "c.db.Where")
	}
	return results, nil
}

func (o *Shield) ListShieldTimeout() ([]*models.Shield, error) {
	var results []*models.Shield
	if err := o.db.Where("err_count > 122").Find(&results).Error; err != nil {
		return nil, errors.Wrap(err, "c.db.Where")
	}

	return results, nil
}

func (o *Shield) ListShieldByStatusAndContractId(status []models.ShieldStatus, address, contractId string) ([]*models.Shield, error) {
	var results []*models.Shield
	if err := o.db.Where("status in (?) and address = ?  and erc20_token  = ?", status, address, contractId).Find(&results).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "c.db.Where")
	}
	return results, nil
}

func (o *Shield) GetAddressEstFeeExists(paymentAddress, walletAddress, privacyTokenAddress string, addressType models.AddressType) (*models.Shield, error) {
	var result models.Shield
	if err := o.db.Where("user_payment_address = ? and inc_address=? and status = ? and address_type = ? and updated_at >= DATE_SUB(NOW(), INTERVAL 1 HOUR) and incognito_token = ?", paymentAddress, walletAddress, models.EstimatedFeeSol, int(addressType), privacyTokenAddress).Order("updated_at DESC").First(&result).Error; err != nil {
		return nil, errors.Wrap(err, "u.db.Where.Find")
	}
	return &result, nil
}

// list wallet:
func (o *Shield) GetAllShieldWallet() ([]*models.ShieldSolWallet, error) {
	var result []*models.ShieldSolWallet
	if err := o.db.Find(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "u.db.Where.Find")
	}
	return result, nil
}

func (o *Shield) GetShieldWalletById(ID uint) (*models.ShieldSolWallet, error) {
	var results models.ShieldSolWallet
	if err := o.db.Where("id = ?", ID).First(&results).Error; err != nil {
		return nil, errors.Wrap(err, "u.db.Where.First")
	}
	return &results, nil
}

func (o *Shield) GetShieldWalletByAddress(paymentAddress string) (*models.ShieldSolWallet, error) {
	var results models.ShieldSolWallet
	if err := o.db.Where("address = ?", paymentAddress).First(&results).Error; err != nil {
		return nil, errors.Wrap(err, "u.db.Where.First")
	}
	return &results, nil
}

func (o *Shield) CreateNewShieldHistory(data *models.ShieldHistory) error {
	return errors.Wrap(o.db.Create(data).Error, "u.db.Create")
}

func (o *Shield) GetShieldByIdAndStatus(id uint, status []models.ShieldStatus) (*models.Shield, error) {
	var result models.Shield
	if err := o.db.Where("id = ? and status in (?)", id, status).First(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, errors.Wrap(err, "u.db.Where.First")
	}

	return &result, nil
}

func (o *Shield) GetShieldById(id uint) (*models.Shield, error) {
	var result models.Shield
	if err := o.db.Where("id = ?", id).First(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, errors.Wrap(err, "u.db.Where.First")
	}

	return &result, nil
}

func (o *Shield) GetShieldHistoryById(id uint) ([]*models.ShieldHistory, error) {
	var result []*models.ShieldHistory
	if err := o.db.Where("job_id = ?", id).Order("id DESC").Find(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, errors.Wrap(err, "u.db.Where.Find")
	}

	return result, nil
}

func (o *Shield) ListShield(page int, limit int, fields map[string]string) ([]*models.Shield, uint, error) {
	var results []*models.Shield
	var count uint
	offset := page*limit - limit

	query := o.db.Table("shield_unshield_sols u")
	query = database.BuildSearchQuery(query, fields, nil, "u", (*models.Shield)(nil))
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrap(err, "db.Count")
	}

	query = query.Order("u.id desc").Limit(limit).Offset(offset)
	if err := query.Scan(&results).Error; err != nil {
		return nil, 0, errors.Wrap(err, "c.db.Where")
	}

	return results, count, nil
}

func (o *Shield) GetSOLToken() (*models.PToken, error) {
	queryString := "select * from p_tokens where currency_type = ? limit 1"

	var result models.PToken
	if err := o.db.Raw(queryString, models.SOL).Scan(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "o.db.GetSOLToken")
	}
	return &result, nil
}

func (o *Shield) GetPTokenByContractID(contractId string) (*models.PToken, error) {
	var result models.PToken
	if err := o.db.Where("LOWER(contract_id) = ? and currency_type=?", strings.ToLower(contractId), models.SOL_SPL).First(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

func (o *Shield) GetPTokenByID(tokenID string) (*models.PToken, error) {

	var result models.PToken
	if err := o.db.Where("token_id = ?", tokenID).First(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

// ShieldSolWallet
func (o *Shield) GetItemByIncWallet(incWallet string, tokenID string) (*models.ShieldSolWallet, error) {
	var result models.ShieldSolWallet
	if err := o.db.Where("LOWER(inc_address) = ? and spl_token_id=?", strings.ToLower(incWallet), tokenID).First(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, errors.Wrap(err, "u.db.Where.First")
	}
	return &result, nil
}
func (o *Shield) SaveShieldSolWallet(model *models.ShieldSolWallet) error {
	if err := o.db.Save(model).Error; err != nil {
		return errors.Wrap(err, "tx.UpdateAddress")
	}
	return nil
}

func (u *Shield) CreateShieldSolWallet(item *models.ShieldSolWallet) error {
	return errors.Wrap(u.db.Create(item).Error, "o.db.CreateShieldSolWallet")
}

func (o *Shield) GetShieldSolWalletByIncAddress(incAddress string) ([]*models.ShieldSolWallet, error) {
	var result []*models.ShieldSolWallet
	if err := o.db.Where("inc_address = ?", incAddress).Find(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "u.db.Where.Find")
	}
	return result, nil
}

// get vault token address:
func (o *Shield) GetVaultTokenByTokenID(tokenID string) (*models.VaultSolToken, error) {
	var result models.VaultSolToken
	if err := o.db.Where("LOWER(spl_token)=?", strings.ToLower(tokenID)).First(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, errors.Wrap(err, "u.db.Where.First")
	}
	return &result, nil
}

func (u *Shield) CreateVaultAddress(model *models.VaultSolToken) error {
	return errors.Wrap(u.db.Create(model).Error, "o.db.CreateVaultAddress")
}
