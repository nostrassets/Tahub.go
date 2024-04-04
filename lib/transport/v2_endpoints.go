package transport

import (
	v2controllers "github.com/getAlby/lndhub.go/controllers_v2"
	"github.com/getAlby/lndhub.go/lib/service"
	"github.com/labstack/echo/v4"
)
// TODO for the purpose of Tahub it is worth considering trimming/refactoring this
//		to be for Admin endpoints only.
func RegisterV2Endpoints(svc *service.LndhubService, e *echo.Echo, secured *echo.Group, securedWithStrictRateLimit *echo.Group, strictRateLimitMiddleware echo.MiddlewareFunc, adminMw echo.MiddlewareFunc, logMw echo.MiddlewareFunc) {
	// get server public key endpoint for signing
	e.GET("/v2/pubkey", v2controllers.NewNostrController(svc).GetServerPubkey, strictRateLimitMiddleware, logMw)
	// get universe assets
	e.GET("/v2/universe-assets", v2controllers.NewUniverseController(svc).UniverseAssets, strictRateLimitMiddleware, logMw)
	// since tahub users register by pubkey, v2 auth returns tokens if a message
	// is signed by the pubkey of our user to the server pubkey
	e.POST("/v2/auth", v2controllers.NewPubkeyAuthController(svc).PubkeyAuth, strictRateLimitMiddleware, adminMw, logMw)
	if svc.Config.AllowAccountCreation {
		/// TAHUB_CREATE_USER / N.S. register modified endpoint
		e.POST("/v2/users", v2controllers.NewCreateUserController(svc).CreateUser, strictRateLimitMiddleware, adminMw, logMw)
	}
	//require admin token for update user endpoint
	if svc.Config.AdminToken != "" {
		e.PUT("/v2/admin/users", v2controllers.NewUpdateUserController(svc).UpdateUser, strictRateLimitMiddleware, adminMw)
	}
	// invoiceCtrl := v2controllers.NewInvoiceController(svc)
	// keysendCtrl := v2controllers.NewKeySendController(svc)
	nostrEventCtrl := v2controllers.NewNostrController(svc)

	// NOSTR EVENT Request - single endpoint that takes Nostr Events
	e.POST("/v2/event", nostrEventCtrl.HandleNostrEvent, strictRateLimitMiddleware, logMw)
	// REST Tahub actions with nostr abstracted
	secured.GET("/v2/balances/all", v2controllers.NewBalanceController(svc).Balances, strictRateLimitMiddleware, logMw)
	secured.POST("/v2/create-address", v2controllers.NewAddressController(svc).CreateAddress, strictRateLimitMiddleware, logMw)
	secured.POST("/v2/transfer", v2controllers.NewTransferController(svc).Transfer, strictRateLimitMiddleware, logMw)

	// secured.POST("/v2/invoices", invoiceCtrl.AddInvoice)
	// secured.GET("/v2/invoices/incoming", invoiceCtrl.GetIncomingInvoices)
	// secured.GET("/v2/invoices/outgoing", invoiceCtrl.GetOutgoingInvoices)
	// secured.GET("/v2/invoices/:payment_hash", invoiceCtrl.GetInvoice)
	// securedWithStrictRateLimit.POST("/v2/payments/bolt11", v2controllers.NewPayInvoiceController(svc).PayInvoice)
	// securedWithStrictRateLimit.POST("/v2/payments/keysend", keysendCtrl.KeySend)
	// securedWithStrictRateLimit.POST("/v2/payments/keysend/multi", keysendCtrl.MultiKeySend)
	secured.GET("/v2/balance/:asset_id", v2controllers.NewBalanceController(svc).Balance)
}
