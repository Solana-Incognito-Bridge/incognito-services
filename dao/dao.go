package dao

import (
	"github.com/incognito-services/models"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

var (
	userTables        = []interface{}{(*models.User)(nil), (*models.UserVerification)(nil), (*models.IncognitoConfig)(nil), (*models.BlackListIP)(nil)}
	pTokenTables      = []interface{}{(*models.PToken)(nil), (*models.PCustomToken)(nil), (*models.PCustomToken)(nil)}
	StakingPoolTables = []interface{}{(*models.StakingPoolStaker)(nil), (*models.StakingOrder)(nil), (*models.StakingPoolStakeNode)(nil)}
)

func AutoMigrate(db *gorm.DB) error {
	allTables := make([]interface{}, 0, len(userTables))

	allTables = append(allTables, userTables...)
	allTables = append(allTables, pTokenTables...)

	allTables = append(allTables, StakingPoolTables...)

	allTables = append(allTables)

	if err := db.AutoMigrate(allTables...).Error; err != nil {
		return errors.Wrap(err, "db.AutoMigrate")
	}

	return nil
}
