package testutils

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// WithAbortedTransaction is a helper wrapper that allows to perform idempotent actions on the mongoDB.
func WithAbortedTransaction(ctx context.Context, client *mongo.Client, f func(ctx context.Context)) {
	wc := writeconcern.New(writeconcern.WMajority())
	rc := readconcern.Snapshot()
	txnOpts := options.Transaction().SetWriteConcern(wc).SetReadConcern(rc)

	session, err := client.StartSession()
	if err != nil {
		panic(err)
	}
	defer session.EndSession(ctx)

	mongo.WithSession(ctx, session, func(sessCtx mongo.SessionContext) error {
		if err = session.StartTransaction(txnOpts); err != nil {
			panic(err)
		}
		f(sessCtx)
		if err = session.AbortTransaction(sessCtx); err != nil {
			panic(err)
		}
		return nil
	})
}
