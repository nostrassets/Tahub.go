package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/getAlby/lndhub.go/common"
	"github.com/getAlby/lndhub.go/db/models"
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
	// get user by address
	tahubUser, err := svc.FindUserByAddress(ctx, completeEvent.Address.Encoded)
	if err != nil {
		svc.Logger.Error("error finding user by address")
		// TODO apply sentry
		return err
	}
	svc.Logger.Infof("tahub user found: %s", tahubUser.Pubkey)
	// decode the asset id
	assetId, err := decodeAssetIdBytes(completeEvent.Address)
	if err != nil {
		svc.Logger.Error("error decoding asset id")
		// TODO apply sentry
		return err
	}
	svc.Logger.Infof("asset id decoded: %s", assetId)
	// get user incoming account (it will go negative which is acceptable per notes in db migration)
	// - this will be the debit_account
	debitAccount, err := svc.AccountFor(ctx, common.AccountTypeIncoming, assetId, tahubUser.ID)
	if err != nil {
		svc.Logger.Error("error getting user incoming account")
		// TODO apply sentry
		return err
	}
	// get user current account
	// - this will be the credit_account
	creditAccount, err := svc.AccountFor(ctx, common.AccountTypeCurrent, assetId, tahubUser.ID)
	if err != nil {
		svc.Logger.Error("error getting user current account")
		// TODO apply sentry
		return err
	}
	// 	transaction entry
	entry := models.TransactionEntry{
		UserID: tahubUser.ID,
		DebitAccountID: debitAccount.ID,
		CreditAccountID: creditAccount.ID,
		Amount: int64(completeEvent.Address.Amount),
		EntryType: models.EntryTypeIncoming,
		Outpoint: completeEvent.Outpoint,
		Addr: completeEvent.Address.Encoded,
	}

	if completeEvent.Timestamp > 0 {
		// TODO ensure that completed Event is always populated even if backoff is too
		// TODO confirm this is the best indication the event has been processed

		// insert the tx entry
		_, err = svc.DB.NewInsert().Model(&entry).Exec(ctx)
		// check error on insertion
		if err != nil {
			svc.Logger.Error("error inserting transaction entry")
			// TODO apply sentry
			return err
		}
		// create message
		message := "received: " + completeEvent.Address.Encoded + " " + fmt.Sprint(completeEvent.Address.Amount) + " " + assetId
		// broadcast the notice to the user
        _ = svc.SendNip4Notification(ctx, message, tahubUser.Pubkey)
	} else {
		// TODO what is this condition?
		// TODO apply sentry
		svc.Logger.Error("event not completed, execution path needs exploration")
	}
	return nil
}
