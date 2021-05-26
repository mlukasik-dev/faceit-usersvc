package store

import "go.mongodb.org/mongo-driver/mongo/options"

type Pagination struct {
	Page uint
	Size uint
}

// findOpts creates an options for mongodb's Find method.
func (p *Pagination) findOpts() *options.FindOptions {
	opts := options.Find()
	if p != nil {
		opts.SetLimit(int64(p.Size))
		opts.SetSkip(int64((p.Page - 1) * p.Size))
	}
	return opts
}
