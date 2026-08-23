package propagator

import "context"

type Carrier interface {
	Get(key string) string
	Set(key, value string)
}

type Context struct {
	TraceID      string
	ParentSpanID string
	Sampled      bool
	Remote       bool
}

type key struct{}

func WithContext(ctx context.Context, value Context) context.Context {
	return context.WithValue(ctx, key{}, value)
}

func FromContext(ctx context.Context) (Context, bool) {
	value, ok := ctx.Value(key{}).(Context)
	return value, ok
}

type Propagator interface {
	Inject(Carrier, Context)
	Extract(Carrier) (Context, bool)
}
