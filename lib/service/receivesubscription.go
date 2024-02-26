package service

import (
	"context"
	"errors"

	"github.com/getAlby/lndhub.go/tapd"
	"github.com/lightninglabs/taproot-assets/taprpc"
)

var AlreadyProcessedTapdEventError = errors.New("already processed tapd event")

func (svc *LndhubService) ConnectReceiveSubscription(ctx context.Context) (tapd.SubscribeReceiveAssetEventWrapper, error) {
	// start tapd receive asset subcription
	svc.Logger.Info("starting tapd receive asset subscription")
	return svc.TapdClient.SubscribeReceiveAssetEvent(ctx, &taprpc.SubscribeReceiveAssetEventNtfnsRequest{})
}

func (svc *LndhubService) HandleTapdReceiveEvent(ctx context.Context) (err error) {

}