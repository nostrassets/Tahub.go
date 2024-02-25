package service

import (
	"context"
	"github.com/getAlby/lndhub.go/db/models"
)

func (svc *LndhubService) CreateAsset(ctx context.Context, name string, tapdAssetId string, tapdAssetType int64) (asset *models.Asset, err error) {
	asset = &models.Asset{}
	asset.AssetName = name
	asset.TaAssetID = tapdAssetId
	asset.AssetType = tapdAssetType
	
	_, err = svc.DB.NewInsert().Model(asset).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return asset, nil
}

func (svc *LndhubService) GetAssets(ctx context.Context) ([]models.Asset, error) {
	asset := []models.Asset{}

	err := svc.DB.NewSelect().Model(&asset).Distinct().Scan(ctx)
	return asset, err
}

func (svc *LndhubService) FindAsset(ctx context.Context, assetId int64) (*models.Asset, error) {
	var asset models.Asset

	err := svc.DB.NewSelect().Model(&asset).Where("id = ?", assetId).Limit(1).Scan(ctx)
	if err != nil {
		return &asset, err
	}
	return &asset, nil
}

func (svc *LndhubService) FindAssetByName(ctx context.Context, assetName string) (*models.Asset, error) {
	var asset models.Asset

	err := svc.DB.NewSelect().Model(&asset).Where("name = ?", assetName).Limit(1).Scan(ctx)
	if err != nil {
		return &asset, err
	}
	return &asset, nil
}

func (svc *LndhubService) UpdateAsset(ctx context.Context, assetId int64) (asset *models.Asset, err error) {
	asset, err = svc.FindAsset(ctx, assetId)
	if err != nil {
		return nil, err
	}
	_, err = svc.DB.NewUpdate().Model(asset).WherePK().Exec(ctx)
	if err != nil {
		return nil, err
	}
	return asset, nil
}

