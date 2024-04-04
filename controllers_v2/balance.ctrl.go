package v2controllers

import (
	"net/http"
	"github.com/getAlby/lndhub.go/common"
	"github.com/getAlby/lndhub.go/lib/responses"
	"github.com/getAlby/lndhub.go/lib/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

// BalanceController : BalanceController struct
type BalanceController struct {
	svc *service.LndhubService
}

func NewBalanceController(svc *service.LndhubService) *BalanceController {
	return &BalanceController{svc: svc}
}

type BalanceResponse struct {
	Balance  int64  `json:"balance"`
	Currency string `json:"currency"`
	Unit     string `json:"unit"`
}

/// get all balances
type BalancesResponse struct {
	Balances map[string]int64 `json:"balances"`
}
// Balance godoc
// @Summary      Retrieve balance
// @Description  Current user's balance in satoshi
// @Accept       json
// @Produce      json
// @Tags         Account
// @Success      200  {object}  BalanceResponse
// @Failure      400  {object}  responses.ErrorResponse
// @Failure      500  {object}  responses.ErrorResponse
// @Router       /v2/balance/:asset_id [get]
// @Security     OAuth2Password
func (controller *BalanceController) Balance(c echo.Context) error {
	userId := c.Get("UserID").(int64)
	assetParam := c.Param("asset_id")
	// default to bitcoin if error parsing the param
	if  assetParam == "" {
		assetParam = common.BTC_TA_ASSET_ID
	}
	balance, err := controller.svc.CurrentUserBalance(c.Request().Context(), assetParam, userId)
	if err != nil {
		c.Logger().Errorj(
			log.JSON{
				"message":        "failed to retrieve user balance",
				"lndhub_user_id": userId,
				"error":          err,
			},
		)
		return c.JSON(http.StatusBadRequest, responses.BadArgumentsError)
	}
	return c.JSON(http.StatusOK, &BalanceResponse{
		Balance:  balance,
		Currency: "BTC",
		Unit:     "sat",
	})
}

/// Balances godoc
/// @Summary      Retrieve all balances
/// @Description  Retrieve all user balances
/// @Accept       json
/// @Produce      json
/// @Tags         Account
/// @Success      200  {object}  BalancesResponse
/// @Failure      400  {object}  responses.ErrorResponse
/// @Failure      500  {object}  responses.ErrorResponse
/// @Router       /v2/balances/all [get]
func (controller *BalanceController) Balances(c echo.Context) error {
	userId := c.Get("UserID").(int64)

	balances, err := controller.svc.GetAllCurrentBalancesJson(c.Request().Context(), userId)
	if err != nil {
		c.Logger().Errorj(
			log.JSON{
				"message":        "failed to retrieve user balances",
				"lndhub_user_id": userId,
				"error":          err,
			},
		)
		return c.JSON(http.StatusBadRequest, responses.BadArgumentsError)
	}
	return c.JSON(http.StatusOK, &BalancesResponse{
		Balances: balances,
	})
}
