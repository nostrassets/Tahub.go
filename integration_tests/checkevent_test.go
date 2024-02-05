package integration_tests

import (
// 	"errors"
	"testing"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/nbd-wtf/go-nostr"
	"github.com/getAlby/lndhub.go/lib/service"
)

type LndhubServiceTestSuite struct {
	suite.Suite
	service    *service.LndhubService
}

func (suite *LndhubServiceTestSuite) TestCheckEvent() {
	
	svc := suite.service

	payload1 := nostr.Event{
		Kind:    1,
		Content: "TAHUB_CREATE_USER",
	}
	result, err := svc.CheckEvent(payload1)
	assert.True(suite.T(), result)
	assert.Nil(suite.T(), err)

	// Test case 2: TAHUB_RECEIVE_ADDRESS_FOR_ASSET
	payload2 := nostr.Event{
		Kind:    1,
		Content: "TAHUB_RECEIVE_ADDRESS_FOR_ASSET:34ttthgfhg:45",
	}
	result, err = svc.CheckEvent(payload2)
	fmt.Println(result)
	assert.True(suite.T(), result)
	assert.Nil(suite.T(), err)

	// Test case 3: Invalid content for TAHUB_RECEIVE_ADDRESS_FOR_ASSET
	payload3 := nostr.Event{
		Kind:    1,
		Content: "TAHUB_RECEIVE_ADDRESS_FOR_ASSET",
	}
	result, err = svc.CheckEvent(payload3)
	assert.False(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.EqualError(suite.T(), err, "Invalid 'Content' for TAHUB_RECEIVE_ADDRESS_FOR_ASSET.")

	payload4 := nostr.Event{
		Kind:    1,
		Content: "DEMO_RECEIVE_ADDRESS_FOR_ASSET",
	}
	result, err = svc.CheckEvent(payload4)
	assert.False(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.EqualError(suite.T(), err, "Undefined 'Content' Name")

	 

}


func TestSuiteCheckEvent(t *testing.T) {
	suite.Run(t, new(LndhubServiceTestSuite))
}

