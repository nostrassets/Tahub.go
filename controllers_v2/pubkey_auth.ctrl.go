package v2controllers

import (
	"net/http"

	"github.com/getAlby/lndhub.go/lib/responses"
	"github.com/getAlby/lndhub.go/lib/service"
	"github.com/labstack/echo/v4"
)
// PubkeyAuthController : PubkeyAuthController struct
type PubkeyAuthController struct {
	svc *service.LndhubService
	responder responses.RelayResponder
}

func NewPubkeyAuthController(svc *service.LndhubService) *PubkeyAuthController {
	return &PubkeyAuthController{svc: svc}
}
/// auth request 
type AuthRequestBody struct {
	Pubkey       string `json:"pubkey"`
	Content	     string `json:"content"`
	RefreshToken string `json:"refresh_token"`
}
/// auth response
type AuthResponseBody struct {
	Pubkey       string `json:"pubkey"`
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}
/// PubkeyAuthentication godoc
/// @Summary      Authenticate by pubkey
/// @Description  Authenticate by pubkey
/// @Accept       json
/// @Produce      json
/// @Tags         Auth
/// @Success      200  {object}  AuthResponseBody
/// @Failure      400  {object}  responses.ErrorResponse
/// @Failure      500  {object}  responses.ErrorResponse
/// @Router       /v2/auth [post]
func (controller *PubkeyAuthController) Auth(c echo.Context) error {
	// payload is an event
	var body AuthRequestBody
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
	// decode content using public key 
	result, err := controller.svc.DecodeNip4Msg(body.Pubkey, body.Content)
	if err != nil || result == "" {
		c.Logger().Errorf("Invalid Nostr Event content: %v", err)
		return controller.responder.NostrErrorJson(c, responses.InvalidTahubContentError.Message)
	}
	// * NOTE this is the stub for acquiring a JWT with a pubkey without a full Nostr implementation
	if result != "TAHUB_AUTH" {
		c.Logger().Errorf("Improper signed content for authentication material: %v", err)
		// TODO this is not a nostr error responder
		return controller.responder.NostrErrorJson(c, responses.InvalidTahubContentError.Message)
	}
	// create access and refresh token
	accessToken, refreshToken, err := controller.svc.GenerateToken(c.Request().Context(), body.Pubkey, body.RefreshToken)
	if err != nil {
		c.Logger().Errorf("Failed to generate token: %v", err)
		// TODO this is not a nostr error responder
		return controller.responder.NostrErrorJson(c, responses.InvalidTahubContentError.Message)
	}
	// respond
	return c.JSON(http.StatusOK, &AuthResponseBody{
		Pubkey:       body.Pubkey,
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	})
}
