package service

import (
	"context"
	//"fmt"
	"errors"
	//"github.com/getAlby/lndhub.go/common"
	//"github.com/getAlby/lndhub.go/db/models"
	"github.com/getAlby/lndhub.go/db/models"
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
			sendEvent, err := sendSubscriptionStream.Recv()
			if err != nil {
				// TODO apply sentry
				return err
			}
			// get pending transfers
			pending, err := svc.GetAllPendingTaprootTransfers(ctx)
			if err != nil {
				// failed to retreive pending transfers
				return err
			} 
			// handle event
			err = svc.HandleTapdSendEvent(ctx, sendEvent, pending)
			if err != nil {
				// TODO apply sentry
				return err
			}
		}
	}
}

func (svc *LndhubService) HandleTapdSendEvent(ctx context.Context, sendEvent *taprpc.SendAssetEvent, pending []models.TransactionEntry) (err error) {
	// check backoff first
	backoffEvent := sendEvent.GetProofTransferBackoffWaitEvent()
	if backoffEvent != nil {
		// handle backoff event
		svc.Logger.Error("backoff event received")
		// wait for completed 
		return nil
	}
	// check if pending transaction entries exist
	if len(pending) < 1 {
		// no pending transactions
		svc.Logger.Error("no pending transactions found but receiving send asset event. Do not have enough information to map to database schema.")
		return nil
	}
	// check send asset event
	event := sendEvent.GetExecuteSendStateEvent()
	if event != nil {
		// handle send asset event
		svc.Logger.Info("send asset event received")
		// TODO match pending to incoming by something found in the broadcast state string
		tx := pending[0]
		// update transaction entry
		success := svc.UpdateTapdTransactionEntry(
			ctx,
			tx.ID,
			tx.TaAssetID,
			tx.UserID,
			event.SendState,
		)
		// check success updating transaction
		if !success {
			// TODO apply sentry
			svc.Logger.Error("error updating transaction entry in database. issue will be handled by daily reconciliation script.")
			return nil
		}
		// TODO attempt to provide informative updates to sender about transaction
		return nil
	}

	return nil
}
