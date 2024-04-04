package service

import (
	"context"
	b64 "encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/getAlby/lndhub.go/common"
	"github.com/lightninglabs/taproot-assets/taprpc"
)

type AssetRoot struct {
	AssetName  string `json:"asset_name"`
	AssetID    string `json:"asset_id"`
	GroupKey   string `json:"group_key"`  
}

func (svc *LndhubService) GetUniverseAssets(ctx context.Context) (okMsg string, success bool) {
	universeRoots, err := svc.GetAssets(ctx)
	if err != nil {
		// TODO OK Relay-Compatible messages need a central location
		return "error: no assets found, possible disconnect.", false
	}
	var okSuccessMsg = "uniassets: "
	// since there can be two root entries per asset (one for issuance and one for transfer: https://lightning.engineering/api-docs/api/taproot-assets/universe/query-asset-roots#universerpcqueryrootresponse)
	// the observedAssetIds array helps us return something the user expects to see i.e. joins asset/transfer entry if both exist
	for _, asset := range universeRoots {
		appendAsset := fmt.Sprintf("%s %s,", asset.TaAssetID, asset.AssetName)

		okSuccessMsg = okSuccessMsg + appendAsset
	}

	return okSuccessMsg, true
}

func (svc *LndhubService) GetUniverseAssetsJson(ctx context.Context) (map[string]string, error) {
	// asset map
	assetMap := make(map[string]string)	
	// get asset data
	assets, err := svc.GetAssets(ctx)
	if err != nil {
		return assetMap, err
	}
	// build success msg
	for _, asset := range assets {
		assetMap[asset.TaAssetID] = asset.AssetName
	}
	return assetMap, nil
}

func (svc *LndhubService) GetAllCurrentBalances(ctx context.Context, userId int64) (string, error) {
	// success message string
	msg := "balances: "
	// get balance data
	balances, err := svc.CurrentUserBalanceByAsset(ctx, userId)
	if err != nil {
		return "error: failed to fetch balances.", err
	}
	// build success msg
	for asset, balance := range balances {
		assetMsg := fmt.Sprintf("%s - %d,", asset, balance)
		msg = msg + assetMsg
	}
	return msg, nil
}

func (svc*LndhubService) GetAllCurrentBalancesJson(ctx context.Context, userId int64) (map[string]int64, error) {
	// balance map
	balanceMap := make(map[string]int64)
	// get balance data
	balances, err := svc.CurrentUserBalanceByAsset(ctx, userId)
	if err != nil {
		return balanceMap, err
	}
	// build success msg
	for asset, balance := range balances {
		balanceMap[asset] = balance
	}
	return balanceMap, nil
}

func  (svc *LndhubService) BalanceByAsset(ctx context.Context) (okMsg string, success bool) {
	/// * TODO this is a placeholder for now use account_ledger population on RecieveNotification Subscription
	filter := taprpc.ListBalancesRequest_AssetId{AssetId: true}
	req := taprpc.ListBalancesRequest{GroupBy: &filter}
	balances, err := svc.TapdClient.ListBalances(ctx, &req)
	if err != nil {
		// TODO OK Relay-Compatible messages need a central location
		return "error: failed to fetch balances.", false
	}
	aggBalances := make(map[string]uint64)
	var okSuccessMsg = "balances: "
	for _, balance := range balances.AssetBalances {

		// seen a group of this asset already
		bal, ok := aggBalances[balance.AssetGenesis.Name]
		if ok {
			aggBalances[balance.AssetGenesis.Name] = bal + balance.Balance
		} else {
			// add to map
			name := balance.AssetGenesis.Name
			aggBalances[name] = balance.Balance
		}
	}
	// check for no len
	if len(aggBalances) == 0 {
		// TODO OK Relay-Compatible messages need a central location
		return "balance: 0", false
	}
	// success message 
	for asset, balance := range aggBalances {
		okSuccessMsg = okSuccessMsg + fmt.Sprintf("%s %d,", asset, balance)
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

func (svc *LndhubService) TransferAssets(ctx context.Context, userId uint64, addr string) (string, bool) {
	// has funding flag
	hasFunding := false
	// decode addr
	req := taprpc.DecodeAddrRequest{Addr: addr}
	decodedAddr, err := svc.TapdClient.GetDecodedAddress(ctx, &req)
	if err != nil {
		return "error: failed to decode address.", false
	}
	sendAmt := decodedAddr.Amount
	sendAssetId := hex.EncodeToString(decodedAddr.AssetId)
	// pull balance for asset - TODO fix this awkward conversion on type mismatch
	balance, err := svc.CurrentUserBalanceForAsset(ctx, sendAssetId, int64(userId))
	if err != nil {
		// TODO OK Relay-Compatible messages need a central location
		return "error: failed to read balance for asset, ensure you own the asset.", false
	}
	// TODO estimate fee rate
	fee := 1
	// TODO apply GetLimits on various configurable limits (see User service for original lndhub reference)
	totalLimits := 1
	// TODO apply service fee
	serviceFee := 1
	// compare current account to send request and limits
	hasFunding = uint64(balance) >= sendAmt + uint64(fee) + uint64(totalLimits) + uint64(serviceFee)
	if !hasFunding {
		// TODO OK Relay-Compatible messages need a central location
		return "error: insufficient funds.", false
	} else {
		// insert pending transaction entry
		debitAccount, err := svc.AccountFor(ctx, common.AccountTypeCurrent, sendAssetId, int64(userId))
		if err != nil {
			svc.Logger.Errorf("Could not find current account user_id:%v", userId)
			return "error: failed to find debit account for send", false
		}
		creditAccount, err := svc.AccountFor(ctx, common.AccountTypeOutgoing, sendAssetId, int64(userId))
		if err != nil {
			svc.Logger.Errorf("Could not find outgoing account user_id:%v", userId)
			return "error: failed to find credit account for send", false
		}
		_, err = svc.InsertTapdTransactionEntry(ctx, int64(userId), creditAccount, debitAccount, sendAmt)
		if err != nil {
			// TODO OK Relay-Compatible messages need a central location
			return "error: failed to create transaction entry. your send was processed but we lost connectivity to our DB. we will reconcile things ASAP.", false
		}		
		// send asset so tx is already in db for status updates
		sendReq := taprpc.SendAssetRequest{
			TapAddrs: []string{addr},
		}
		_, err = svc.TapdClient.SendAsset(ctx, &sendReq)
		if err != nil {
			// TODO OK Relay-Compatible messages need a central location
			return "error: failed to send asset.", false
		}
		// return success message
		msg := fmt.Sprintf("success: sent %s", sendAssetId)
		return msg, true
	}
}

func (svc *LndhubService) FetchOrCreateAssetAddr(ctx context.Context, userId uint64, assetId string, amt uint64) (string, error) {
	assetMatch := false
	amtMatch   := false
	// fetch all addresses for asset, will attempt to match on amount later
	addrs, err := svc.FindAddresses(ctx, userId, assetId)
	// check db error
	if err != nil {
		return "error: failed to check on existing address.", err
	}
	// decode assetId for tapd request
	decoded, err := hex.DecodeString(assetId)
	if err != nil {
		return "error: failed to parse assetID.", err	
	}
	// addrs is nil - return 
	if len(addrs) >= 1 {
		// setting flag to prevent creation of additional accounts for the asset
		assetMatch = true
		// attempt to match on amount
		for _, addr := range addrs {
			if addr.Amount == amt {
				// setting flag to prevent creation of any additional receiver resources. address exists for exact amount.
				amtMatch = true
				return addr.Addr, nil
			}
		}
	}
	// definitely need to create an address, and may also need to create accounts if no addresses are found for the asset (of any amount)
	req := taprpc.NewAddrRequest{
		AssetId: decoded,
		Amt: amt,
	}
	newAddr, err := svc.TapdClient.NewAddress(ctx, &req)
	if err != nil {
		return "error: tapd failed to create receive address.", err
	}
	// determine if new accounts should be created
	createAccounts := !assetMatch && !amtMatch
	_, err = svc.CreateAddress(ctx, newAddr.Encoded, userId, assetId, amt, createAccounts)
	// note the defensive return
	if err == nil {
		return newAddr.Encoded, nil
	} 
	return "error: failed to create or fetch address.", nil
}
