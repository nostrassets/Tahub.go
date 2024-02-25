package service

import (
	"errors"
	//"github.com/lightninglabs/taproot-assets/taprpc"
)

var AlreadyProcessedTapdEventError = errors.New("already processed tapd event")
// * TODO insert receive addresses for user as they request them

// * TODO find user by receive address
