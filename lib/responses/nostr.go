package responses

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nbd-wtf/go-nostr"
)

/// generic error 
type NostrErrorResponseBody struct {
	Message string `json:"message"`
}
/// create user with nostr event
type NostrCreateUserResponseBody struct {
	UserID      int64  `json:"user_id"`
}
/// get server public key
type NostrServerPubkeyResponseBody struct {
	NpubHex 	  string `json:"npub_hex"`
}

type NostrBalanceResponseBody struct {
	Balances map[string]int64 `json:"balances"`
}
/// universe response
type NostrUniAssetResponseBody struct {
	Assets map[string]string `json:"assets"`
}
/// transfer response
type NostrTransferResponseBody struct {
	Message string `json:"message"`
}
/// new address response
type NostrAddressResponseBody struct {
	Address string `json:"address"`
}
/// auth response
type AuthResponseBody struct {
	Pubkey       string `json:"pubkey"`
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}


type RelayResponder struct {} 
// relay-compatible responder
func (responder *RelayResponder) NostrErrorResponse(c echo.Context, errMsg string) error {
	msg := fmt.Sprintf("error: %s", errMsg)
	res := []interface{}{"OK", -1, false, msg}

	return c.JSON(http.StatusOK, res)
}

func (responder *RelayResponder) NostrErrorJson(c echo.Context, errMsg string) error {
	var ErrorResponse NostrErrorResponseBody
	ErrorResponse.Message = errMsg

	return c.JSON(http.StatusBadRequest, &ErrorResponse)
}

func (responder *RelayResponder) CreateUserOk(c echo.Context, event nostr.Event, userId int64, isError bool, errMsg string) error {
	// TODO these messages should end up in a central location
	var status = false
	var msg = fmt.Sprintf("error: %s", errMsg)

	if !isError {
		msg = fmt.Sprintf("userid: %d", userId)
		status = true
	}
	
	res := []interface{} {"OK", event.ID, status, msg}
	return c.JSON(http.StatusOK, res)
}

func (responder *RelayResponder) CreateUserJson(c echo.Context, userId int64) error {
	var res NostrCreateUserResponseBody
	res.UserID = userId

	return c.JSON(http.StatusOK, &res)
}

func (responder *RelayResponder) UniverseAssetsJson(c echo.Context, assetMap map[string]string) error {
	var res NostrUniAssetResponseBody
	res.Assets = assetMap

	return c.JSON(http.StatusOK, &res)
}

func (responder *RelayResponder) GetAddressJson(c echo.Context, address string) error {
	var res NostrAddressResponseBody
	res.Address = address

	return c.JSON(http.StatusOK, &res)
}

func (responder *RelayResponder) GetBalancesJson(c echo.Context, balances map[string]int64) error {
	var res NostrBalanceResponseBody
	res.Balances = balances

	return c.JSON(http.StatusOK, &res)
}

func (responder *RelayResponder) TransferAssetsJson(c echo.Context, msg string) error {
	var res NostrTransferResponseBody
	res.Message = msg

	return c.JSON(http.StatusOK, &res)
}

func (responder *RelayResponder) AuthJson(c echo.Context, pubkey string, accessToken string, refreshToken string) error {
	var res AuthResponseBody
	res.Pubkey = pubkey
	res.AccessToken = accessToken
	res.RefreshToken = refreshToken

	return c.JSON(http.StatusOK, &res)
}

func (responder *RelayResponder) GetServerPubkeyOk(c echo.Context, event nostr.Event, serverNpub string, isError bool, errMsg string) error {
	// TODO these messages should end up in a central location
	var status = false
	var msg = fmt.Sprintf("error: %s", errMsg)

	if !isError {
		msg = fmt.Sprintf("pubkey: %s", serverNpub)
		status = true
	}

	res := []interface{} {"OK", event.ID, status, msg}
	return c.JSON(http.StatusOK, res)
}

func (responder *RelayResponder) GetServerPubkeyJson(c echo.Context, serverNpub string) error {
	var res NostrServerPubkeyResponseBody
	res.NpubHex = serverNpub

	return c.JSON(http.StatusOK, &res)
}

func (responder *RelayResponder) GenericOk(c echo.Context, event nostr.Event, msg string, status bool) error {
	// TODO these messages should end up in a central location
	res := []interface{} {"OK", event.ID, status, msg}
	return c.JSON(http.StatusOK, res)
}


type CreateUserEventResponseBody struct {
	// internal tahub user id
	ID     int64 `json:"id"`
	// nostr public key, discovered via the event
	Pubkey string `json:"pubkey"`
}

type GetServerPubkeyResponseBody struct {
	TahubPubkeyHex   string `json:"tahub_pubkey"`
	TahubNpub        string `json:"tahub_npub"`
}