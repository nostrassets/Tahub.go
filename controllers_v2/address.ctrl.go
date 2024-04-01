package v2controllers

import (
	"net/http"
	"strconv"

	"github.com/getAlby/lndhub.go/lib/responses"
	"github.com/getAlby/lndhub.go/lib/service"
	"github.com/labstack/echo/v4"
)

// AddressController : AddressController struct
type AddressController struct {
	svc *service.LndhubService
	responder responses.RelayResponder
}

func NewAddressController(svc *service.LndhubService) *AddressController {
	return &AddressController{svc: svc, responder: responses.RelayResponder{}}
}

// get address request
type AddressRequestBody struct {
	AssetId string `json:"asset_id"`
	Amt     string  `json:"amt"`
}

type AddressResponseBody struct {
	Address string `json:"address"`
}

/// Address godoc
// @Summary      Get or create address
// @Description  Get or create address for deposit
// @Accept       json
// @Produce      json
// @Tags         Address

// @Success      200      {object}  AddressResponseBody
// @Failure      400      {object}  responses.ErrorResponse
// @Failure      500      {object}  responses.ErrorResponse
// @Router       /v2/create-address [post]
func (controller *AddressController) CreateAddress(c echo.Context) error {
	userId := c.Get("UserID").(uint64)
	// payload is an event
	var body AddressRequestBody
	// load request payload, params into nostr.Event struct
	if err := c.Bind(&body); err != nil {
		c.Logger().Errorf("Failed to load Nostr Event request body: %v", err)
		// TODO this is not a nostr error responder
		return controller.responder.NostrErrorJson(c, responses.BadArgumentsError.Message)
	}
	// potentially redundant validation
	if err := c.Validate(&body); err != nil {
		c.Logger().Errorf("Invalid Nostr Event request body: %v", err)
		// TODO this is not a nostr error responder
		return controller.responder.NostrErrorJson(c, responses.BadArgumentsError.Message)
	}
	// convert string amount to uint
	amt, err := strconv.ParseUint(body.Amt, 10, 64)
	if err != nil {
		c.Logger().Errorf("Invalid amount. Pass value as a string: %v", err)
		return c.JSON(http.StatusBadRequest, responses.BadArgumentsError)
	}
	result, err := controller.svc.FetchOrCreateAssetAddr(c.Request().Context(), userId, body.AssetId, amt)
	if err != nil {
		c.Logger().Errorf("error creating address: %v", err)
		return c.JSON(http.StatusBadRequest, responses.BadArgumentsError)
	}
	return c.JSON(http.StatusOK, &AddressResponseBody{
		Address: result,
	})
}