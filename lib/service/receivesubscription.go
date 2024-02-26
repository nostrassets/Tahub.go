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

func (svc *LndhubService) TapdReceiveSubscription(ctx context.Context) (err error) {
	rcvSubscriptionStream, err := svc.ConnectReceiveSubscription(ctx)
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
			rcvEvent, err := rcvSubscriptionStream.Recv()
			if err != nil {
				// TODO apply sentry
				return err
			}
			// handle event
			err = svc.HandleTapdReceiveEvent(ctx, rcvEvent)
			if err != nil {
				// TODO apply sentry
				return err
			}
		}
	}
}

func (svc *LndhubService) HandleTapdReceiveEvent(ctx context.Context, rcvEvent *taprpc.ReceiveAssetEvent) (err error) {
	// check backoff first
	backoffEvent := rcvEvent.GetProofTransferBackoffWaitEvent()
	if backoffEvent.TriesCounter > 0 {
		// TODO something is going wrong
		svc.Logger.Error("backoff event received")
		// TODO apply sentry
		return nil
	}
	// check complete event
	completeEvent := rcvEvent.GetAssetReceiveCompleteEvent()
	// TODO could this be null
	if completeEvent.Timestamp > 0 {
		// TODO confirm this is the best indication the event has been processed

		// TODO create current account entry
		
		// event is completed
	}


	// TODO check if event has already been processed
	// if err != nil {
	// 	return err
	// }
	// process event
	// TODO process event
	return nil
}