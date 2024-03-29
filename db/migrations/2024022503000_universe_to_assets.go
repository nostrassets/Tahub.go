package migrations

import (
	"context"
	"encoding/hex"
	"log"
	"github.com/getAlby/lndhub.go/db/models"
	"github.com/getAlby/lndhub.go/lib"
	"github.com/getAlby/lndhub.go/lib/service"
	"github.com/getAlby/lndhub.go/tapd"
	"github.com/kelseyhightower/envconfig"
	"github.com/lightninglabs/taproot-assets/taprpc/universerpc"
	"github.com/uptrace/bun"
	"golang.org/x/exp/slices"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// setup
		c := &service.Config{}
		// Load configruation from environment variables
		c.LoadEnv()

		err := envconfig.Process("", c)
		if err != nil {
			log.Fatalf("Error loading environment variables: %v", err)
		}
		// setup logger - although not used in other go migrations
		logger := lib.Logger(c.LogFilePath)
		// initialize tapd config
		tapdConfig, err := tapd.LoadConfig()
		// check for error
		if err != nil {
			return err
		}
		// initialize tapd client
		tapdClient, err := tapd.InitTAPDClient(tapdConfig, logger, ctx)
		// check for error
		if err != nil {
			return err
		}
		// get universe assets
		req := universerpc.AssetRootRequest{}
		res, err := tapdClient.GetUniverseAssets(ctx, &req)
		// check for error
		if err != nil {
			return err
		}
		// for asset in universe, insert row to database if non exists
		assets := []models.Asset{}
		// hardcode bitcoin row
		bitcoin := models.Asset{
			AssetName: "bitcoin",
			TaAssetID: "btc",
			AssetType: 0,
		}
		assets = append(assets, bitcoin)
		// defend bulk list of assets from having multiple groups of the same asset ID
		assetIds := []string{}
		// iterate over universe assets
		for _, root := range res.UniverseRoots {
			// get asset stats, to avoid group ID use in ID spot on GetUniverseAssets
			req := universerpc.AssetStatsQuery{
				AssetNameFilter: root.AssetName,
			}
			assetStats, err := tapdClient.GetAssetStats(ctx, &req)
			// check for error or non-specifc filter on asset name
			if err != nil {
				return err
			}
			assetIdBytes := assetStats.AssetStats[0].GroupAnchor.AssetId
			decodedAssetId := hex.EncodeToString(assetIdBytes)
			// check for duplicate asset
			if !slices.Contains(assetIds, decodedAssetId) {
				// populate asset model
				asset := models.Asset{
					AssetName: root.AssetName,
					TaAssetID: decodedAssetId,
					// TODO this is not on the universe rpc. hard-code for now
					AssetType: 0,
				}
				// append for bulk upsert
				assets = append(assets, asset)
				// add to checklist
				assetIds = append(assetIds, decodedAssetId)
			}
		}
		// bulk upsert of assets
		_, err = db.NewInsert().Model(&assets).On("conflict (asset_name) do nothing").Exec(ctx)
		if err != nil {
			return err
		}
		// success reached
		return nil
	}, nil)
}
