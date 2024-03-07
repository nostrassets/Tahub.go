package service

import (
	"context"
	//"fmt"
	"errors"
	//"github.com/getAlby/lndhub.go/common"
	//"github.com/getAlby/lndhub.go/db/models"
	"github.com/getAlby/lndhub.go/tapd"
	"github.com/lightninglabs/taproot-assets/taprpc"
)

var AlreadyProcessedTapdSendEventError = errors.New("already processed tapd event")

func (svc *LndhubService) ConnectSendSubscription(ctx context.Context) (tapd.SubscribeSendAssetEventWrapper, error) {
	// start tapd send asset subcription
	svc.Logger.Info("starting tapd send asset subscription")
	return svc.TapdClient.SubscribeSendAssetEvent(ctx, &taprpc.SubscribeSendAssetEventNtfnsRequest{})
}

func (svc *LndhubService) TapdSendSubscription(ctx context.Context) (err error) {
	sendSubscriptionStream, err := svc.ConnectSendSubscription(ctx)
	if err != nil {
		// TODO apply sentry
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// receive event
			_, err := sendSubscriptionStream.Recv()
			if err != nil {
				// TODO apply sentry
				return err
			}
			// handle event
			//err = svc.HandleTapdSendEvent(ctx, sendEvent)
			// if err != nil {
			// 	// TODO apply sentry
			// 	return nil
			// }
		}
	}
}

// func (svc *LndhubService) HandleTapdSendEvent(ctx context.Context, sendEvent *taprpc.SendAssetEvent) (err error) {

// }
