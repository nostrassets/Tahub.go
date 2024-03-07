-- TODO factor in AssetID to account ledgers
CREATE VIEW account_ledgers(
	ta_asset_id,
	account_id,
	transaction_entry_id,
	amount
) AS
    SELECT
		transaction_entries.ta_asset_id,
		transaction_entries.credit_account_id,
		transaction_entries.id,
		transaction_entries.amount
	FROM
		transaction_entries
	UNION ALL
	SELECT
		transaction_entries.ta_asset_id,
		transaction_entries.debit_account_id,
		transaction_entries.id,
		(0 - transaction_entries.amount)
	FROM
		transaction_entries;
