package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/getAlby/lndhub.go/db/models"
	"github.com/getAlby/lndhub.go/lib/responses"
	"github.com/getAlby/lndhub.go/lib/tokens"
	"github.com/getAlby/lndhub.go/lnd"
	"github.com/getAlby/lndhub.go/rabbitmq"
	"github.com/getAlby/lndhub.go/tapd"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/random"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/uptrace/bun"
	"github.com/ziflex/lecho/v3"
)

const alphaNumBytes = random.Alphanumeric

type LndhubService struct {
	Config         *Config
	DB             *bun.DB
	TapdClient     tapd.TapdClientWrapper
	LndClient      lnd.LightningClientWrapper
	RabbitMQClient rabbitmq.Client
	Logger         *lecho.Logger
	InvoicePubSub  *Pubsub
	TaprootAssetPubSub *TapdPubsub
}

func (svc *LndhubService) ParseInt(value interface{}) (int64, error) {
	switch v := value.(type) {
	case float64:
		return int64(v), nil
	case string:
		c, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, err
		}
		return c, nil
	default:
		return 0, fmt.Errorf("conversion to int from %T not supported", v)
	}
}

func (svc *LndhubService) CheckEvent(payload nostr.Event) (bool, nostr.Event, error) {
	if payload.Kind != 4 {
		return false, payload, errors.New("Field 'kind' must be 4")
	}
	// TODO perform checks on content
	// check the length of the content
	if len(payload.Content) == 0 {
		return false, payload, errors.New("Field 'Content' must have a value")
	}
	
	sharedSecret, err := nip04.ComputeSharedSecret(payload.PubKey, svc.Config.TahubPrivateKey)
	if err != nil {
		return false, payload, errors.New("Failed to compute shared secret for sender.")
	}
	decodedContent, err := nip04.Decrypt(payload.Content, sharedSecret)
	if err != nil {
		return false, payload, errors.New("Failed to decode payload with shared secret")
	}
	payload.Content = decodedContent
	// Split event content
	data := strings.Split(payload.Content, ":")
	if len(data) == 0 {
		return false, payload, errors.New("Field 'Content' must at least specify the action.")
	}

	switch data[0] {

	case "TAHUB_CREATE_USER":
		return true, payload, nil
	case "TAHUB_GET_SERVER_PUBKEY":
		return true, payload, nil
	case "TAHUB_GET_UNIVERSE_ASSETS":
		return true, payload, nil
	case "TAHUB_GET_BALANCES":
		return true, payload, nil
	case "TAHUB_AUTH":
		return true, payload, nil
	case "TAHUB_GET_RCV_ADDR":
		// this action must have three parts to the content
		if len(data) != 3 {
			return false, payload, errors.New("Invalid 'Content' for TAHUB_GET_RCV_ADDR.")
		}
		// Validate specific fields for TAHUB_GET_RCV_ADDR event

		// TODO come up with further validations for this asset_id i.e. a Taproot Asset AssetID or 'btc'
		// validate asset ID
		if data[1] == "" {
			return false, payload, errors.New("Field 'Asset ID' must have a value")
		}
		// validate amt
		amt, err := strconv.ParseUint(data[2], 10, 64)
		if err != nil || amt == 0 {
			return false, payload, errors.New("Field 'amt' must be a valid number and non-zero")
		}

		return true, payload, nil

	case "TAHUB_SEND_ASSET":
		// this action must have three parts to the content
		if len(data) != 2 {
			return false, payload, errors.New("Invalid 'Content' for TAHUB_SEND_ASSET.")
		}
		// Validate specific fields for TAHUB_SEND_ASSET event
		// TODO consider other validation on the address
		if data[1] == "" {
			return false, payload, errors.New("Field 'ADDR' must have a value")
		}

		return true, payload, nil

	default:
		return false, payload, errors.New("Undefined 'Content' Name")
	}

}

func (svc *LndhubService) DecodeNip4Msg(pubKey string, encryptedContent string) (string, error) {
	// check the length of the content
	if len(encryptedContent) == 0 {
		return "", errors.New("Field 'Content' must have a value")
	}
	// compute the shared secret
	sharedSecret, err := nip04.ComputeSharedSecret(pubKey, svc.Config.TahubPrivateKey)
	if err != nil {
		return "", errors.New("Failed to compute shared secret for sender.")
	}
	// decode content
	decodedContent, err := nip04.Decrypt(encryptedContent, sharedSecret)
	if err != nil {
		return "", errors.New("Failed to decode payload with shared secret")
	}
	/// return decoded payload
	return decodedContent, nil
}

func (svc *LndhubService) OneAssetInMultiKeysend(arr []string) bool {
	for i := 1; i < len(arr); i++ {
		// compare every item to the first positioned item
		if arr[i] != arr[0] {
			return false
		}
	}
	return true
}

func (svc *LndhubService) ValidateUserMiddleware() echo.MiddlewareFunc {
	// TODO update ValidateUserMiddlware
	// * it has already performed a check on the pubkey for the event passed to endpoint
	// * it must know ensure that pubkey returns a user in the database
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userId := c.Get("UserID").(int64)
			if userId == 0 {
				return echo.ErrUnauthorized
			}
			user, err := svc.FindUser(c.Request().Context(), userId)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
					"error":   true,
					"code":    1,
					"message": "bad auth",
				})
			}
			if user.Deactivated || user.Deleted {
				return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
					"error":   true,
					"code":    1,
					"message": "bad auth",
				})
			}
			return next(c)
		}
	}
}


func (svc *LndhubService) GenerateToken(ctx context.Context, pubkey string, inRefreshToken string) (accessToken, refreshToken string, err error) {
	var user models.User

	switch {
	// * NOTE this needs to be gated by the authentication ckecks in auth.ctrl or pubkey_auth.ctrl
	case inRefreshToken == "":
		{
			if err := svc.DB.NewSelect().Model(&user).Where("pubkey = ?", pubkey).Scan(ctx); err != nil {
				return "", "", fmt.Errorf("bad auth")
			}
		}
	case inRefreshToken != "":
		{
			userId, err := tokens.GetUserIdFromToken(svc.Config.JWTSecret, inRefreshToken)
			if err != nil {
				return "", "", fmt.Errorf("bad auth")
			}

			if err := svc.DB.NewSelect().Model(&user).Where("id = ?", userId).Scan(ctx); err != nil {
				return "", "", fmt.Errorf("bad auth")
			}
		}
	default:
		{
			return "", "", fmt.Errorf("login and password or refresh token is required")
		}
	}

	if user.Deactivated || user.Deleted {
		return "", "", fmt.Errorf(responses.AccountDeactivatedError.Message)
	}

	accessToken, err = tokens.GenerateAccessToken(svc.Config.JWTSecret, svc.Config.JWTAccessTokenExpiry, &user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = tokens.GenerateRefreshToken(svc.Config.JWTSecret, svc.Config.JWTRefreshTokenExpiry, &user)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}
