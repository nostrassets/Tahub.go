package service

import (
	"sync"
	//"github.com/lightninglabs/taproot-assets/taprpc"
	//"github.com/getAlby/lndhub.go/db/models"
)

// This should give enough space to allow for some spike in traffic without flooding memory to much.
const DefaultTapdChannelBufSize = 20

type TapdPubsub struct {
	mu   sync.RWMutex
	subs map[string]map[string]chan bool
}

func NewTapdPubsub() *TapdPubsub {
	ps := &TapdPubsub{}
	ps.subs = make(map[string]map[string]chan bool)
	return ps
}

func (ps *TapdPubsub) TapdSubscribe(topic string) (chan bool, string, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if ps.subs[topic] == nil {
		ps.subs[topic] = make(map[string]chan bool)
	}
	//re-use preimage code for a uuid
	preImageHex, err := makePreimageHex()
	if err != nil {
		return nil, "", err
	}
	subId := string(preImageHex)
	ch := make(chan bool, DefaultTapdChannelBufSize)
	ps.subs[topic][subId] = ch
	return ch, subId, nil
}

func (ps *TapdPubsub) TapdUnsubscribe(id string, topic string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if ps.subs[topic] == nil {
		return
	}
	if ps.subs[topic][id] == nil {
		return
	}
	close(ps.subs[topic][id])
	delete(ps.subs[topic], id)
}

func (ps *TapdPubsub) TapdPublish(topic string, msg bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if ps.subs[topic] == nil {
		return
	}

	for _, ch := range ps.subs[topic] {
		ch <- msg
	}
}