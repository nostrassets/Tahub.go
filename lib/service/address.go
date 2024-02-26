package service

import (
	"context"
	//"encoding/hex"
	//"errors"

	"github.com/getAlby/lndhub.go/db/models"
)

type AddressResponse struct {
	AssetName  string `json:"asset_name"`
	TaAssetID  string `json:"ta_asset_id"`
	Addr       string `json:"addr"`  
}


func (svc *LndhubService) CreateAddress(ctx context.Context, address string, userId uint64, taAssetId string, amt uint64) (addr *models.Address, err error) {
	addrObj := &models.Address{}
	// encAddr := make([]byte, hex.EncodedLen(len(address)))
	// // hex encode address for sql length issue
	// val := hex.Encode(encAddr, []byte(address))
	// // if wrote 0 bytes, return error
	// if val == 0 {
	// 	return nil, err
	// }
	addrObj.Addr = address
	addrObj.UserId = userId
	addrObj.TaAssetID = taAssetId
	addrObj.Amount = amt

	_, err = svc.DB.NewInsert().Model(addrObj).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return addr, nil
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
