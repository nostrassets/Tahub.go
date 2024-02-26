package service

import (
	"context"
	b64 "encoding/base64"
	//"encoding/hex"
	"fmt"
	//"slices"
	//"strings"

	"github.com/lightninglabs/taproot-assets/taprpc"
	//"github.com/lightninglabs/taproot-assets/taprpc/universerpc"
)

type AssetRoot struct {
	AssetName  string `json:"asset_name"`
	AssetID    string `json:"asset_id"`
	GroupKey   string `json:"group_key"`  
}

// TODO most of this logic was moved to the universe_to_assets go migration
//      abstract that to a utility function
func (svc *LndhubService) GetUniverseAssets(ctx context.Context) (okMsg string, success bool) {
	// req := universerpc.AssetRootRequest{}
	// universeRoots, err := svc.TapdClient.GetUniverseAssets(ctx, &req)
	universeRoots, err := svc.GetAssets(ctx)
	if err != nil {
		// TODO OK Relay-Compatible messages need a central location
		return "error: no assets found, possible disconnect.", false
	}
	var okSuccessMsg = "uniassets: "
	// since there can be two root entries per asset (one for issuance and one for transfer: https://lightning.engineering/api-docs/api/taproot-assets/universe/query-asset-roots#universerpcqueryrootresponse)
	// the observedAssetIds array helps us return something the user expects to see i.e. joins asset/transfer entry if both exist
	
	//var observedAssetIds = []string{}

	// TODO confirm when the key may be the group key hash instead of the assetId

	// for assetId, root := range universeRoots.UniverseRoots {
	// 	rawAssetId := strings.Split(assetId, "-")[1]
	// 	seen := slices.Contains(observedAssetIds, rawAssetId)

	// 	if !seen {
	// 		decoded, err := hex.DecodeString(rawAssetId)

	// 		if err != nil {
	// 			// TODO OK Relay-Compatible messages need a central location
	// 			return "error: failed to parse assetID.", false				
	// 		}

	// 		final := b64.StdEncoding.EncodeToString(decoded)

	// 		appendAsset := fmt.Sprintf("%s %s,", final, root.AssetName)
	// 		okSuccessMsg = okSuccessMsg + appendAsset

	// 		observedAssetIds = append(observedAssetIds, rawAssetId)
	// 	}
	// }
	for _, asset := range universeRoots {
		appendAsset := fmt.Sprintf("%s %s,", asset.TaAssetID, asset.AssetName)

		okSuccessMsg = okSuccessMsg + appendAsset
	}

	return okSuccessMsg, true
}

func (svc *LndhubService) GetAddressByAssetId(ctx context.Context, assetId string, amt uint64) (okMsg string, success bool) {
	decoded, err := b64.StdEncoding.DecodeString(assetId)
	if err != nil {
		// TODO OK Relay-Compatible messages need a central location
		return "error: failed to parse assetID.", false	
	}

	req := taprpc.NewAddrRequest{
		AssetId: decoded,
		Amt: amt,
	}
	newAddr, err := svc.TapdClient.NewAddress(ctx, &req)
	if err != nil {
		// TODO OK Relay-Compatible messages need a central location
		return "error: failed to create receive address.", false
	}
	return fmt.Sprintf("address: %s", newAddr.Encoded), true
}

func (svc *LndhubService) FetchOrCreateAssetAddr(ctx context.Context, userId uint64, assetId string, amt uint64) (string, error) {
	addr, err := svc.FindAddress(ctx, userId, assetId, amt)
	// check db error
	if err != nil {
		return "error: failed to check on existing address.", err
	}
	// return if existing address found
	if addr.ID > 0 {
		return addr.Address, nil
	}
	// decode assetId for tapd request
	decoded, err := b64.StdEncoding.DecodeString(assetId)
	if err != nil {
		// TODO OK Relay-Compatible messages need a central location
		return "error: failed to parse assetID.", err	
	}
	// create new address
	req := taprpc.NewAddrRequest{
		AssetId: decoded,
		Amt: amt,
	}
	newAddr, err := svc.TapdClient.NewAddress(ctx, &req)
	if err != nil {
		// TODO OK Relay-Compatible messages need a central location
		return "error: failed to create receive address.", err
	}
	// save new address to db
	_, err = svc.CreateAddress(ctx, newAddr.Encoded, userId, assetId)
	if err != nil {
		// TODO OK Relay-Compatible messages need a central location
		return "error: failed to save receive address.", err
	}
	// return success message
	return fmt.Sprintf("address: %s", newAddr.Encoded), nil
}