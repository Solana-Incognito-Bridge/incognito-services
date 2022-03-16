package dao

import (
	"strings"

	"github.com/inc-backend/api-dapp/serializers"

	"github.com/inc-backend/api-dapp/database"
	"github.com/inc-backend/api-dapp/models"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type Raydium struct {
	db *gorm.DB
}

func (o *Raydium) GetDB() *gorm.DB {
	return o.db
}

func NewRaydium(db *gorm.DB) *Raydium {
	return &Raydium{db}
}

func (o *Raydium) Create(data *models.Raydium) error {
	return errors.Wrap(o.db.Create(data).Error, "u.db.Create")
}

func (o *Raydium) GetById(id uint) (*models.Raydium, error) {
	var result models.Raydium
	if err := o.db.Where("id = ?", id).First(&result).Error; err != nil {
		return nil, errors.Wrap(err, "u.db.Where.First")
	}
	return &result, nil
}

func (o *Raydium) GetByIdToSubmitBurnTx(id uint) (*models.Raydium, error) {
	var result models.Raydium
	if err := o.db.Where("id = ? and status = ?", id, models.EstimateFee).First(&result).Error; err != nil {
		return nil, errors.Wrap(err, "u.db.Where.First")
	}

	return &result, nil
}

func (o *Raydium) List(page int, limit int, fields map[string]string) ([]*models.Raydium, uint, error) {
	var results []*models.Raydium
	var count uint
	offset := page*limit - limit

	query := o.db.Table("trades")
	query = database.BuildSearchQuery(query, fields, nil, (*models.Raydium)(nil))
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrap(err, "db.Count")
	}

	query = query.Order("id desc").Limit(limit).Offset(offset)
	if err := query.Scan(&results).Error; err != nil {
		return nil, 0, errors.Wrap(err, "c.db.Where")
	}

	return results, count, nil
}

func (o *Raydium) Update(model *models.Raydium) error {
	if err := o.db.Save(model).Error; err != nil {
		return errors.Wrap(err, "tx.Update")
	}
	return nil
}

// get exit item:
func (o *Raydium) GetEstTradingFeeExists(
	walletAddress,
	srcTokens,
	destTokens string,
	tradeType int,
) (*models.Raydium, error) {
	var result models.Raydium
	if err := o.db.Where("wallet_address = ? and src_tokens=? and dest_tokens = ? and status = ? and trade_type=? and updated_at >= DATE_SUB(NOW(), INTERVAL 10 HOUR) ", walletAddress, srcTokens, destTokens, int(models.EstimateFee), tradeType).Order("updated_at DESC").First(&result).Error; err != nil {
		return nil, errors.Wrap(err, "u.db.Where.Find")
	}
	return &result, nil
}

func (o *Raydium) GetPTokenByID(tokenID string) (*models.PToken, error) {
	var result models.PToken
	if err := o.db.Where("token_id = ?", tokenID).First(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (o *Raydium) GetSolToken() (*models.PToken, error) {
	queryString := "select * from p_tokens where currency_type = 23 limit 1"

	var result models.PToken
	if err := o.db.Raw(queryString).Scan(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "o.db.GetSolToken")
	}
	return &result, nil
}

func (o *Raydium) GetAllRaydiumTokenExistOnChain() ([]*serializers.PTokenRes, error) {
	var result []*serializers.PTokenRes

	queryString := "select p.*, d.is_popular, d.priority from dapp_tokens d join p_tokens p on LOWER(d.contract_id) = LOWER(p.contract_id) where p.status = 1 and p.currency_type = 24 and source='Raydium' order by d.priority"
	if err := o.db.Raw(queryString).Scan(&result).Error; err != nil {
		return nil, errors.Wrap(err, "u.db.Scan")
	}

	return result, nil
}

func (o *Raydium) TrackHistory(item *models.Raydium, status models.HistoryStatus, requestMsg string, responseMsg string) error {
	trackData := &models.RaydiumHistory{
		JobId:         item.ID,
		JobStatus:     int(item.Status),
		JobStatusName: models.ConvertRaydiumToString(item.Status),
		Status:        status,
		RequestMsg:    requestMsg,
		ResponseMsg:   responseMsg,
	}
	return errors.Wrap(o.db.Create(trackData).Error, "u.db.Create")
}

func (o *Raydium) GetJobHistory(id uint) ([]*models.RaydiumHistory, error) {
	var result []*models.RaydiumHistory
	if err := o.db.Where("job_id = ?", id).Order("id DESC").Find(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, errors.Wrap(err, "u.db.Where.Find")
	}

	return result, nil
}

func (o *Raydium) GetListTimeout() ([]*models.Raydium, error) {
	var results []*models.Raydium
	if err := o.db.Where("status != ? and err_count >= ?", models.Invalid, models.MaxErr+10).Find(&results).Error; err != nil {
		return nil, errors.Wrap(err, "c.db.Where")
	}
	return results, nil
}

func (o *Raydium) GetDAppTokenBySource(contractAddress, source string) (*models.DappToken, error) {
	var result models.DappToken
	if err := o.db.Where("LOWER(contract_id) = ? and source=?", strings.ToLower(contractAddress), source).First(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &result, nil
}
