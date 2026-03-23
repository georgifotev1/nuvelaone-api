package txmanager

import "context"

type contextKey string

const txKey contextKey = "tx"

func WithTx(ctx context.Context, tx DBTX) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

func TxFromContext(ctx context.Context) (DBTX, bool) {
	tx, ok := ctx.Value(txKey).(DBTX)
	return tx, ok
}
