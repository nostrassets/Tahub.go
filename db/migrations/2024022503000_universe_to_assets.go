package migrations

import (
	"context"
	"encoding/hex"
	b64 "encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/getAlby/lndhub.go/db/models"
	"github.com/getAlby/lndhub.go/lib"
	"github.com/getAlby/lndhub.go/lib/service"
	"github.com/getAlby/lndhub.go/tapd"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/lightninglabs/taproot-assets/taprpc/universerpc"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// setup
		c := &service.Config{}
		// Load configruation from environment variables
		err := godotenv.Load(".env")
		if err != nil {
			fmt.Println("Failed to load .env file")
		}
		err = envconfig.Process("", c)
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
			TaAssetID: "BTC",
			AssetType: 0,
		}
		assets = append(assets, bitcoin)
		// iterate over universe assets
		for assetId, root := range res.UniverseRoots {
			// get human-readable asset ID
			rawAssetId := strings.Split(assetId, "-")[1]
			decodedAssetId, err := hex.DecodeString(rawAssetId)
			// check error
			if err != nil {
				return err
			}
			// final form
			base64AssetId := b64.StdEncoding.EncodeToString(decodedAssetId)
			// populate asset model
			asset := models.Asset{
				AssetName: root.AssetName,
				TaAssetID: base64AssetId,
				// TODO this is not on the universe rpc. hard-code for now
				AssetType: 0,
			}
			// append for bulk upsert
			assets = append(assets, asset)
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
