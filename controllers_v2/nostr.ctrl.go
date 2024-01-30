package v2controllers

import (
	"net/http"
	"github.com/nbd-wtf/go-nostr"
	"github.com/getAlby/lndhub.go/lib/responses"
	"github.com/getAlby/lndhub.go/lib/service"
	"github.com/labstack/echo/v4"
)

// NostrController : Add NoStr Event controller struct
type NostrController struct {
	svc *service.LndhubService
}

func NewNostrController(svc *service.LndhubService) *NostrController {
	return &NostrController{svc: svc}
}

type CreateUserEventResponseBody struct {
	// internal tahub user id
	ID     int64 `json:"id"`
	// nostr public key, discovered via the event
	Pubkey string `json:"pubkey"`
}

type GetTahubPublicKey struct {
	TaHubPublicKey   string `json:"tahubpublickey"`
	Pubkey string `json:"pubkey"`
}


func (controller *NostrController) AddNostrEvent(c echo.Context) error {
	
	var body service.EventRequestBody

	if err := c.Bind(&body); err != nil {
		c.Logger().Errorf("Failed to load AddNostrEvent request body: %v", err)
		return c.JSON(http.StatusBadRequest, responses.BadArgumentsError)
	}

	if err := c.Validate(&body); err != nil {
		c.Logger().Errorf("Invalid AddNostrEvent request body: %v", err)
		return c.JSON(http.StatusBadRequest, responses.BadArgumentsError)
	}
	switch body.Content {
		
	case "TAHUB_CREATE_USER":
		user, err := controller.svc.CreateUser(c.Request().Context(), body.Pubkey)
		if err != nil {
			// create user error response
			c.Logger().Errorf("Failed to create user via Nostr event: %v", err)
			return c.JSON(http.StatusInternalServerError, responses.GeneralServerError)
		}
	
		// create user success response
		var ResponseBody CreateUserEventResponseBody
		ResponseBody.ID = user.ID
		ResponseBody.Pubkey = user.Pubkey
		return c.JSON(http.StatusOK, &ResponseBody)
	
	case "GET_SERVER_PUBKEY":
		var ResponseBody GetTahubPublicKey
		ResponseBody.Pubkey = body.Pubkey
		ResponseBody.TaHubPublicKey = controller.svc.Config.TaHubPublicKey
		return c.JSON(http.StatusOK, &ResponseBody)
	
	default:
		// TODO handle next events
		return c.JSON(http.StatusBadRequest, responses.UnimplementedError)
	}

}

