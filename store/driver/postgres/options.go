package postgres

type QueryOptions struct {
	// Ignore return will cause the query to ignore the return value from a
	// Insert, Upsert or Update query. Normally these queries will return/update
	// in place the new value it is returning.
	IgnoreReturn bool
}

type QueryOption func(opt *QueryOptions) error

var DefaultQueryOptions = QueryOptions{
	IgnoreReturn: false,
}

func QueryOptionIgnoreReturn(v bool) QueryOption {
	return func(opt *QueryOptions) error {
		opt.IgnoreReturn = v
		return nil
	}
}
