package service

import (
	"context"
	"github.com/getAlby/lndhub.go/db/models"
)

func (svc *LndhubService) CreateAddress(ctx context.Context, address string, userId uint64, taAssetId string) (addr *models.Address, err error) {
	addr = &models.Address{}
	addr.Address = address
	addr.UserId = userId
	addr.TaAssetID = taAssetId

	_, err = svc.DB.NewInsert().Model(addr).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func (svc * LndhubService) UpdateAddress(ctx context.Context, taAssetId string, userId uint64, address string, amt uint64) (addr *models.Address, err error) {
	addr, err = svc.FindAddress(ctx, userId, taAssetId, amt)
	if err != nil {
		return nil, err
	}
	addr.Address = address
	_, err = svc.DB.NewUpdate().Model(addr).WherePK().Exec(ctx)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func (svc *LndhubService) GetAddresses(ctx context.Context, userId uint64) ([]models.Address, error) {
	addr := []models.Address{}

	err := svc.DB.NewSelect().Model(&addr).Where("user_id = ?", userId).Distinct().Scan(ctx)
	return addr, err
}

func (svc *LndhubService) FindAddress(ctx context.Context, userId uint64, taAssetId string, amt uint64) (*models.Address, error) {
	var addr models.Address

	err := svc.DB.NewSelect().Model(&addr).Where("user_id = ? AND amount = ? AND ta_asset_id = ?", userId, amt, taAssetId).Limit(1).Scan(ctx)
	if err != nil {
		return &addr, err
	}

	return &addr, nil
}
