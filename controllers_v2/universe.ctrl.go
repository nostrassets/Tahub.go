package v2controllers

import (
	"net/http"
	"github.com/getAlby/lndhub.go/lib/responses"
	"github.com/getAlby/lndhub.go/lib/service"
	"github.com/labstack/echo/v4"
)

type UniverseController struct {
	svc *service.LndhubService
	responder responses.RelayResponder
}

func NewUniverseController(svc *service.LndhubService) *UniverseController {
	return &UniverseController{svc: svc, responder: responses.RelayResponder{}}
}

// universe assets response
/// universe response
type UniverseAssetsResponseBody struct {
	Assets map[string]string `json:"assets"`
}

// Universe godoc
// @Summary      Retrieve universe assets
// @Description  Retrieve universe assets
// @Accept       json
// @Produce      json
// @Tags         Universe
// @Success      200  {object}  UniverseAssetsResponseBody
// @Failure      400  {object}  responses.ErrorResponse
// @Failure      500  {object}  responses.ErrorResponse
// @Router       /v2/universe-assets [get]
func (controller *UniverseController) UniverseAssets(c echo.Context) error {
	data, err := controller.svc.GetUniverseAssetsJson(c.Request().Context())
	if err != nil {
		c.Logger().Errorf("Failed to retrieve universe assets: %v", err)
		return c.JSON(http.StatusBadRequest, responses.BadArgumentsError)
	}

	return c.JSON(http.StatusOK, &UniverseAssetsResponseBody{
		Assets: data,
	})
}