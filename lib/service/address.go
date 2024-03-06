package service

import (
	"context"
	//"encoding/hex"
	//"errors"

	"github.com/getAlby/lndhub.go/common"
	"github.com/getAlby/lndhub.go/db/models"
	"github.com/uptrace/bun"
)

type AddressResponse struct {
	AssetName  string `json:"asset_name"`
	TaAssetID  string `json:"ta_asset_id"`
	Addr       string `json:"addr"`  
}


func (svc *LndhubService) CreateAddress(ctx context.Context, address string, userId uint64, taAssetId string, amt uint64, createAccounts bool) (addr *models.Address, err error) {
	addrObj := &models.Address{}

	addrObj.Addr = address
	addrObj.UserId = userId
	addrObj.TaAssetID = taAssetId
	addrObj.Amount = amt
	// add accounts for address
	err = svc.DB.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// insert new address
		if _, err := tx.NewInsert().Model(addrObj).Exec(ctx); err != nil {
			return err
		}
		// check if we need to create accounts
		if createAccounts {
			// get account types - ex fees for non-bitcoin accounts
			accountTypes := []string{
				common.AccountTypeIncoming,
				common.AccountTypeCurrent,
				common.AccountTypeOutgoing,
				//common.AccountTypeFees,
			}
			// for each account type
			for _, accountType := range accountTypes {
				// create account - TODO ensure joint uniqueness on type/ta_asset_id
				account := models.Account{
					UserID: int64(userId), 
					Type: accountType, 
					TaAssetID: taAssetId,
				}
				// insert account
				if _, err := tx.NewInsert().Model(&account).Exec(ctx); err != nil {
					return err
				}
			}
		}
		// exit db transaction
		return nil
	})
	return addr, err
}

func (svc * LndhubService) UpdateAddress(ctx context.Context, taAssetId string, userId uint64, address string, amt uint64) (addr *models.Address, err error) {
	addr, err = svc.FindAddress(ctx, userId, taAssetId, amt)
	// check error
	if err != nil {
		return nil, err
	}
	// check not found (this is likely redundant to the first check)
	if addr == nil {
		return nil, nil
	}
	// var encAddr []byte
	// // hex encode address for sql length issue
	// val := hex.Encode(encAddr, []byte(address))
	// // if wrote 0 bytes, return error
	// if val == 0 {
	// 	return nil, err
	// }
	addr.Addr   = address
	addr.Amount = amt
	_, err = svc.DB.NewUpdate().Model(addr).WherePK().Exec(ctx)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func (svc *LndhubService) GetAddresses(ctx context.Context, userId uint64) ([]models.Address, error) {
	/// NOTE: THIS DOES THE HEX DECODE
	addresses := []models.Address{}
	//response  := []AddressResponse{}

	// get all addresses for user
	err := svc.DB.NewSelect().Model(&addresses).Where("user_id = ?", userId).Distinct().Scan(ctx)
	if err != nil {
		return nil, err
	}

	// TODO use bulk helper here
	// for _, addr := range addresses {
	// 	// decode address
	// 	var decodedAddr []byte
	// 	_, err := hex.Decode(decodedAddr, addr.Addr)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	// convert address bytes to string
	// 	decodedAddrStr := string(decodedAddr)
	// 	// create AddressResponse object for the address
	// 	addrResp := AddressResponse{
	// 		AssetName: addr.Asset.AssetName,
	// 		TaAssetID: addr.TaAssetID,
	// 		Addr: addr.Addr,
	// 	}
	// 	// append to result set
	// 	response = append(response, addrResp)
	// }
	//return response, err

	return addresses, err
}

func (svc *LndhubService) FindAddress(ctx context.Context, userId uint64, taAssetId string, amt uint64) (*models.Address, error) {
	var addr models.Address

	err := svc.DB.NewSelect().Model(&addr).Where("user_id = ? AND amount = ? AND ta_asset_id = ?", userId, amt, taAssetId).Limit(1).Scan(ctx)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		// sql error or not found
		return &addr, err
	}
	// success
	return &addr, nil
}

func (svc *LndhubService) FindAddresses(ctx context.Context, userId uint64, taAssetId string) ([]models.Address, error) {
	addresses := []models.Address{}
	err := svc.DB.NewSelect().Model(&addresses).Where("user_id = ? AND ta_asset_id = ?", userId, taAssetId).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return addresses, nil
}

func (svc *LndhubService) LookupUserByAddr(ctx context.Context, addr string) (*models.Address, error) {
	var address models.Address
	// get user by address
	err := svc.DB.
		NewSelect().
		Model(&address).
		Where("addr = ?", addr).Relation("User").Relation("Asset").Limit(1).Scan(ctx)
	if err != nil {
		return &address, err
	}
	return &address, nil
}

// func decodedAddresses(addresses []models.Address) ([]AddressResponse, error) {
// 	response := []AddressResponse{}
// 	// TODO decode all addresses - look for more efficient way to do this
// 	for _, addr := range addresses {
// 		// decode address
// 		var decodedAddr []byte
// 		_, err := hex.Decode(decodedAddr, addr.Addr)
// 		if err != nil {
// 			return nil, err
// 		}
// 		// convert address bytes to string
// 		decodedAddrStr := string(decodedAddr)
// 		// create AddressResponse object for the address
// 		addrResp := AddressResponse{
// 			AssetName: addr.Asset.AssetName,
// 			TaAssetID: addr.TaAssetID,
// 			Addr: decodedAddrStr,
// 		}
// 		// append to result set
// 		response = append(response, addrResp)
// 	}
// 	return response, nil
// }

// func decodeDbAddressToString(address []byte) (string, error) {
// 	var decodedAddr []byte
// 	_, err := hex.Decode(decodedAddr, address)
// 	if err != nil {
// 		return string(decodedAddr), err
// 	}
// 	// convert address bytes to string
// 	decodedAddrStr := string(decodedAddr)
// 	return decodedAddrStr, nil
// }

// func encodeDbAddressHex(address string) ([]byte, error) {
// 	var encAddr []byte
// 	// hex encode address for sql length issue
// 	val := hex.Encode(encAddr, []byte(address))
// 	// if wrote 0 bytes, return error
// 	if val == 0 {
// 		return encAddr, errors.New("no bytes written")
// 	}
// 	return encAddr, nil
// }
