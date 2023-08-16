package middleware

import (
	"github.com/kataras/iris/v12/context"
)

type (
	Option  func(*options)
	options struct {
		resourceExtract func(*context.Context) string
		blockFallback   func(*context.Context)
	}
)

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	for _, opt := range opts {
		opt(optCopy)
	}

	return optCopy
}

// WithResourceExtractor sets the resource extractor of the web requests.
func WithResourceExtractor(fn func(*context.Context) string) Option {
	return func(opts *options) {
		opts.resourceExtract = fn
	}
}

// WithBlockFallback sets the fallback handler when requests are blocked.
func WithBlockFallback(fn func(ctx *context.Context)) Option {
	return func(opts *options) {
		opts.blockFallback = fn
	}
}
