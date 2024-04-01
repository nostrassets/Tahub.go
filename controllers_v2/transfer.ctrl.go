package v2controllers

import (
	"net/http"
	"github.com/getAlby/lndhub.go/lib/responses"
	"github.com/getAlby/lndhub.go/lib/service"
	"github.com/labstack/echo/v4"
)

type TransferController struct {
	svc *service.LndhubService
	responder responses.RelayResponder
}

func NewTransferController(svc *service.LndhubService) *TransferController {
	return &TransferController{svc: svc, responder: responses.RelayResponder{}}
}
// transfer request
type TransferRequestBody struct {
	Address string `json:"address"`
}
// transfer response
type TransferResponseBody struct {
	Message string `json:"message"`
	Status  bool  `json:"status"`
}
// Transfer godoc
// @Summary      Transfer assets
// @Description  Transfer assets to an address
// @Accept       json
// @Produce      json
// @Tags         Transfer
// @Param        transfer  body      TransferRequestBody  false  "Transfer"
// @Success      200      {object}  TransferResponseBody
// @Failure      400      {object}  responses.ErrorResponse
// @Failure      500      {object}  responses.ErrorResponse
// @Router       /v2/transfer [post]
// @Security     NIP 4

func (controller *TransferController) Transfer(c echo.Context) error {
	userId := c.Get("UserID").(uint64)
	// payload is an event
	var body TransferRequestBody
	// load request payload, params into nostr.Event struct
	if err := c.Bind(&body); err != nil {
		c.Logger().Errorf("Failed to bind transfer request body: %v", err)
		// TODO this is not a nostr error responder
		return controller.responder.NostrErrorJson(c, responses.BadArgumentsError.Message)
	}
	// potentially redundant validation
	if err := c.Validate(&body); err != nil {
		c.Logger().Errorf("Failed to validate transfer request body: %v", err)
		// TODO this is not a nostr error responder
		return controller.responder.NostrErrorJson(c, responses.BadArgumentsError.Message)
	}
	// transfer assets
	msg, status := controller.svc.TransferAssets(c.Request().Context(), userId, body.Address)
	if !status {
		c.Logger().Errorf("Failed to send assets: %v", msg)
		// TODO improve error response
		return controller.responder.NostrErrorJson(c, responses.BadArgumentsError.Message)
	}
	return c.JSON(http.StatusOK, &TransferResponseBody{
		Message: msg,
		Status:  status,
	})
}

