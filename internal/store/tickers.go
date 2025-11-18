package store

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

func InsertTickers(ctx context.Context, db *sql.DB, tickers []string) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	//Create temp list of new tickers
	if _, err := tx.ExecContext(ctx, `CREATE TEMP TABLE _new(symbol text) ON COMMIT DROP`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO _new(symbol) SELECT UNNEST($1::text[])`, pq.Array(tickers)); err != nil {
		return err
	}

	//Upsert and reset queue state for incoming tickers
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO tickers (symbol, updated_at, is_processed)
		SELECT symbol, now(), false FROM _new
		ON CONFLICT (symbol) DO UPDATE SET
			updated_at   = EXCLUDED.updated_at,
			is_processed = false;
	`); err != nil {
		return err
	}

	//Delete any ticker that has been processed (regardless of snapshot existence)
	//and is not in the new upload list.
	// napshots remain safe because the FK is now RESTRICT (no cascade).
	if _, err := tx.ExecContext(ctx, `
    DELETE FROM tickers t
    WHERE t.is_processed = true
      AND NOT EXISTS (SELECT 1 FROM _new n WHERE n.symbol = t.symbol)
      AND NOT EXISTS (SELECT 1 FROM snapshots s WHERE s.symbol = t.symbol);
`); err != nil {
		return err
	}

	return tx.Commit()
}
