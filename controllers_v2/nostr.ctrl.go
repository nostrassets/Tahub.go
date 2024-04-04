package v2controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/getAlby/lndhub.go/lib/responses"
	"github.com/getAlby/lndhub.go/lib/service"
	"github.com/labstack/echo/v4"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

// NostrController : Add NoStr Event controller struct
type NostrController struct {
	svc *service.LndhubService
	responder responses.RelayResponder
}

func NewNostrController(svc *service.LndhubService) *NostrController {
	return &NostrController{svc: svc, responder: responses.RelayResponder{}}
}
// A utility endpoint to recover the server pubkey w/o creating a nostr event
func (controller *NostrController) GetServerPubkey(c echo.Context) error {
	res, err := controller.HandleGetPublicKey()
	if err != nil {
		c.Logger().Errorf("Failed to handle / encode public key: %v", err)
		return c.JSON(http.StatusInternalServerError, responses.NostrServerError)
	}

	return c.JSON(http.StatusOK, &res)
}

func (controller *NostrController) HandleNostrEvent(c echo.Context) error {
	// The main nostr event handler
	var body nostr.Event
	// load request payload, params into nostr.Event struct
	if err := c.Bind(&body); err != nil {
		c.Logger().Errorf("Failed to load Nostr Event request body: %v", err)
		return controller.responder.NostrErrorJson(c, responses.BadArgumentsError.Message)
	}
	// potentially redundant validation
	if err := c.Validate(&body); err != nil {
		c.Logger().Errorf("Invalid Nostr Event request body: %v", err)
		return controller.responder.NostrErrorJson(c, responses.BadArgumentsError.Message)
	}
	// check signature
	if result, err := body.CheckSignature(); (err != nil || !result) {
		c.Logger().Errorf("Signature is not valid for the event... Consider monitoring this user if issue persists: %v", err)
		return controller.responder.NostrErrorJson(c, responses.BadAuthError.Message)
	}
	// call our payload validator 
	result, decodedPayload, err := controller.svc.CheckEvent(body)
	if err != nil || !result {
		c.Logger().Errorf("Invalid Nostr Event content: %v", err)
		return controller.responder.NostrErrorJson(c, responses.InvalidTahubContentError.Message)
	}
	// check that this event is not a duplicate, even though it is from the REST API
	status, err := controller.svc.InsertEvent(c.Request().Context(), body)
	if err != nil || !status {
		// * specifically handle duplicate events
		dupEvent := strings.Contains(err.Error(), "unique constraint")
		if dupEvent {
			// * NOTE we are not responding to duplicate events and trusting the filter
			//   minimizes the workload we have on a given restart
			c.Logger().Errorf("Duplicate event detected: %v", err)
			return nil
		} else {
			// * likely db connectivity issue, since payload is valid
			c.Logger().Errorf("Failed to insert event into database: %v", err)
		}
	}
	// Split event content
	data := strings.Split(decodedPayload.Content, ":")
	// handle create user event - can assume valid thanks to middleware
	if data[0] == "TAHUB_CREATE_USER" {
		// TODO determine if a check against config is required
		// 		in Tahub's case: https://github.com/nostrassets/Tahub.go/blob/a798601f63d5847b045360e45e8090081bb4cd85/lib/transport/v2_endpoints.go#L12
		// check if user exists
		existingUser, err := controller.svc.FindUserByPubkey(c.Request().Context(), decodedPayload.PubKey)
		// check if user was found
		if existingUser.ID > 0 {
			c.Logger().Errorf("Cannot create user that has already registered this pubkey")
			return controller.responder.NostrErrorJson(c, "this pubkey has already been registered.")
		}
		// confirm no error occurred in checking if the user exists
		if err != nil {
			msg := err.Error()
			// TODO consider this and try to make more robust
			if msg != "sql: now rows in result set" {
				c.Logger().Info("Error is related to no results in the dataset, which is acceptable.")
			} else {
				c.Logger().Errorf("Unable to verify the pubkey has not already been registered: %v", err)
				return controller.responder.NostrErrorJson(c, "failed to check pubkey.")
			}
		}
		// create the user, by public key
		user, err := controller.svc.CreateUser(c.Request().Context(), decodedPayload.PubKey)
		if err != nil {
			// create user error response
			c.Logger().Errorf("Failed to create user via Nostr event: %v", err)
			return controller.responder.NostrErrorJson(c, "failed to insert user into database.")
		}
		// create user success response
		return controller.responder.CreateUserJson(c, user.ID)
	} else if data[0] == "TAHUB_AUTH" {
		// authentication required
		existingUser, isAuthenticated := controller.svc.GetUserIfExists(c.Request().Context(), decodedPayload)
		if existingUser == nil || !isAuthenticated {
			controller.svc.Logger.Errorf("Failed to authenticate user for get rcv addr.")

			return controller.responder.NostrErrorJson(c, responses.BadAuthError.Message)
		}
		// create table with optional relationship for access_token and refresh_token

		// assign access_token and refresh_token to user
		
		// return access_token and refresh_token in Responder method
		return nil
	} else if data[0] == "TAHUB_GET_SERVER_PUBKEY" {
		// get server npub
		res, err := controller.HandleGetPublicKey()
		if err != nil {
			c.Logger().Errorf("Failed to handle / encode public key: %v", err)
			return controller.responder.NostrErrorJson(c, responses.NostrServerError.Message)
		}
		// return server npub
		return controller.responder.GetServerPubkeyJson(c, res.TahubPubkeyHex)

	} else if data[0] == "TAHUB_GET_UNIVERSE_ASSETS" {
		// get universe known assets 
		data, status := controller.svc.GetUniverseAssetsJson(c.Request().Context())
		if status != nil {
			controller.svc.Logger.Errorf("Failed to get universe assets: %v", status)
			return controller.responder.NostrErrorJson(c, responses.GeneralServerError.Message)
		}
		// * NOTE status is passed to an isError flag
		return controller.responder.UniverseAssetsJson(c, data)
	} else if data[0] == "TAHUB_GET_RCV_ADDR" {
		// authentication required
		existingUser, isAuthenticated := controller.svc.GetUserIfExists(c.Request().Context(), decodedPayload)
		if existingUser == nil || !isAuthenticated {
			controller.svc.Logger.Errorf("Failed to authenticate user for get rcv addr.")

			return controller.responder.NostrErrorJson(c, responses.BadAuthError.Message)
		}
		// given an asset_id and amt, return the address
		// these values are prevalidated by CheckEvent
		assetId := data[1]
		amt, err := strconv.ParseUint(data[2], 10, 64)
		if err != nil {
			c.Logger().Errorf("Failed to parse amt field in content: %v", err)
			return controller.responder.NostrErrorJson(c, responses.GeneralServerError.Message)
		}
		msgContent, err := controller.svc.FetchOrCreateAssetAddr(c.Request().Context(), uint64(existingUser.ID), assetId, amt)
		if err != nil {
			// set isError status to true for the error response
			status = true
			controller.svc.Logger.Errorf("Failed to fetch or create asset address: %v", err)
			return controller.responder.NostrErrorJson(c, responses.GeneralServerError.Message)
		}
		// create msg
		msg := "address: " + msgContent
		/// * NOTE status is passed to an isError flag
		return controller.responder.GetAddressJson(c, msg)
	} else if data[0] == "TAHUB_GET_BALANCES" {
		// authentication required
		existingUser, isAuthenticated := controller.svc.GetUserIfExists(c.Request().Context(), decodedPayload)
		if existingUser == nil || !isAuthenticated {
			controller.svc.Logger.Errorf("Failed to authenticate user for get balances.")
			return controller.responder.NostrErrorJson(c, responses.BadAuthError.Message)
		}
		// pull all accounts
		// group by assets, total current accounts - outgoing accounts
		data, err := controller.svc.GetAllCurrentBalancesJson(c.Request().Context(), existingUser.ID)
		if err != nil {
			controller.svc.Logger.Errorf("Failed to get all current balances: %v", err)
			return controller.responder.NostrErrorJson(c, responses.GeneralServerError.Message)
		}
		// respond
		return controller.responder.GetBalancesJson(c, data)
	} else if data[0] == "TAHUB_SEND_ASSET" {
		// authentication required
		existingUser, isAuthenticated := controller.svc.GetUserIfExists(c.Request().Context(), decodedPayload)
		if existingUser == nil || !isAuthenticated {
			controller.svc.Logger.Errorf("Failed to authenticate user for send asset.")
			return controller.responder.NostrErrorJson(c, responses.BadAuthError.Message)
		}
		// check balance and send
		msg, status := controller.svc.TransferAssets(c.Request().Context(), uint64(existingUser.ID), data[1])
		if !status {
			controller.svc.Logger.Errorf("Failed to transfer assets: %v", msg)
			return controller.responder.NostrErrorJson(c, responses.GeneralServerError.Message)
		} else {
			// success
			return controller.responder.TransferAssetsJson(c, msg)
		}
	} else {
		// catch all - unimplemented
		controller.svc.Logger.Errorf("Unimplemented Nostr Event content: %v", decodedPayload.Content)
		return controller.responder.NostrErrorJson(c, "unimplemented.")
	}
}

func (controller *NostrController) HandleGetPublicKey() (responses.GetServerPubkeyResponseBody, error) {
	var ResponseBody responses.GetServerPubkeyResponseBody
	ResponseBody.TahubPubkeyHex = controller.svc.Config.TahubPublicKey
	npub, err := nip19.EncodePublicKey(controller.svc.Config.TahubPublicKey)
	// TODO improve this
	if err != nil {
		return ResponseBody, err
	}
	ResponseBody.TahubNpub = npub
	return ResponseBody, nil
}