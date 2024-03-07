package models

// Account : Account Model
type Account struct {
	ID      int64  `bun:",pk,autoincrement"`
	UserID  int64  `bun:",notnull"`
	User    *User  `bun:"rel:belongs-to,join:user_id=id"`
	TaAssetID string  `bun:",notnull"`
	Asset   *Asset `bun:"rel:has-one,join:ta_asset_id=ta_asset_id"`
	Type    string `bun:",notnull"`
}
