package db

import (
	"context"
	"github.com/google/uuid"
	"token-based-payment-service-api/types"
)

func CreateImage(image *types.Image) error {
	_, err := Bun.NewInsert().Model(image).Exec(context.Background())
	return err
}

func GetImagesByUserId(userId uuid.UUID) ([]types.Image, error) {
	var images []types.Image
	err := Bun.NewSelect().
		Model(&images).
		Where("user_id = ?", userId).
		Where("deleted = ?", false).
		Order("created_at DESC").
		Scan(context.Background())

	return images, err
}

func GetImageById(userId uuid.UUID, imageId int) (types.Image, error) {
	var image types.Image
	err := Bun.NewSelect().
		Model(&image).
		Where("user_id = ?", userId).
		Where("id = ?", imageId).
		Where("deleted = ?", false).
		Order("created_at DESC").
		Scan(context.Background())

	return image, err
}

func GetAccountByUserId(userId uuid.UUID) (types.Account, error) {
	account := types.Account{}
	err := Bun.NewSelect().Model(&account).Where("user_id = ?", userId).Scan(context.Background())

	return account, err
}

func CreateAccount(account *types.Account) error {
	_, err := Bun.NewInsert().Model(account).Exec(context.Background())
	return err
}

func UpdateAccount(account *types.Account) error {
	_, err := Bun.NewUpdate().Model(account).WherePK().Exec(context.Background())
	return err
}
