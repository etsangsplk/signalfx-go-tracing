package buntdb

import "context"

type config struct {
	ctx           context.Context
	serviceName   string
	analyticsRate float64
}

func defaults(cfg *config) {
	cfg.serviceName = "buntdb"
	cfg.ctx = context.Background()
	// cfg.analyticsRate = globalconfig.AnalyticsRate()
}

// An Option customizes the config.
type Option func(cfg *config)

// WithContext sets the context for the transaction.
func WithContext(ctx context.Context) Option {
	return func(cfg *config) {
		cfg.ctx = ctx
	}
}

// WithServiceName sets the given service name for the transaction.
func WithServiceName(serviceName string) Option {
	return func(cfg *config) {
		cfg.serviceName = serviceName
	}
}

// WithAnalytics enables Trace Analytics for all started spans.
func WithAnalytics(on bool) Option {
	if on {
		return WithAnalyticsRate(1.0)
	}
	return WithAnalyticsRate(0.0)
}

// WithAnalyticsRate sets the sampling rate for Trace Analytics events
// correlated to started spans.
func WithAnalyticsRate(rate float64) Option {
	return func(cfg *config) {
		cfg.analyticsRate = rate
	}
}
