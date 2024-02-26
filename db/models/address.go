package models

import (
	"context"
	"time"
	"github.com/uptrace/bun"
)

// Address : Address Model
type Address struct {
	ID         uint64 `bun:",pk,autoincrement"`
	Address    string `bun:",notnull,unique"`
	UserId     uint64 `bun:",notnull"`
	TaAssetID  string `bun:",notnull"`
	Amount     uint64 
	CreatedAt  time.Time `bun:",notnull,default:current_timestamp"`
	UpdatedAt  bun.NullTime `bun:",nullzero"`
	// relationship
	Asset 	*Asset `bun:"rel:has-one,join:asset_id=ta_asset_id"`
}

func (a *Address) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.UpdateQuery:
		a.UpdatedAt = bun.NullTime{Time: time.Now()}
	}
	return nil
}

var _ bun.BeforeAppendModelHook = (*Address)(nil)

