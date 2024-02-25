package service

import (
	"context"
	"github.com/getAlby/lndhub.go/db/models"
)

func (svc *LndhubService) CreateAddress(ctx context.Context, address string, userId int64, assetId int64) (addr *models.Address, err error) {
	addr = &models.Address{}
	addr.Address = address
	addr.UserId = userId
	addr.AssetId = assetId

	_, err = svc.DB.NewInsert().Model(addr).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func (svc * LndhubService) UpdateAddress(ctx context.Context, assetId int64, userId int64, address string) (addr *models.Address, err error) {
	addr, err = svc.FindAddress(ctx, userId, assetId)
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

func (svc *LndhubService) GetAddresses(ctx context.Context, userId int64) ([]models.Address, error) {
	addr := []models.Address{}

	err := svc.DB.NewSelect().Model(&addr).Where("user_id = ?", userId).Distinct().Scan(ctx)
	return addr, err
}

func (svc *LndhubService) FindAddress(ctx context.Context, userId int64, assetId int64) (*models.Address, error) {
	var addr models.Address

	err := svc.DB.NewSelect().Model(&addr).Where("user_id = ? AND asset_id = ?", userId, assetId).Limit(1).Scan(ctx)
	if err != nil {
		return &addr, err
	}
	return &addr, nil
}
