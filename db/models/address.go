package models

import (
	"context"
	"time"
	"github.com/uptrace/bun"
)

// Address : Address Model
type Address struct {
	ID         int64 `bun:",pk,autoincrement"`
	Address    string `bun:",notnull,unique"`
	UserId     int64 `bun:",notnull"`
	AssetId	   int64 `bun:",notnull"`
	CreatedAt  time.Time `bun:",notnull,default:current_timestamp"`
	UpdatedAt  bun.NullTime `bun:",nullzero"`
}

func (a *Address) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.UpdateQuery:
		a.UpdatedAt = bun.NullTime{Time: time.Now()}
	}
	return nil
}

var _ bun.BeforeAppendModelHook = (*Address)(nil)

