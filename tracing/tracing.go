package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/opentracer"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
	"os"
	"strings"
)

const (
	signalfxServiceName = "SIGNALFX_SERVICE_NAME"
	signalfxEndpointURL = "SIGNALFX_ENDPOINT_URL"
	signalfxAccessToken = "SIGNALFX_ACCESS_TOKEN"
	signalfxSpanTags    = "SIGNALFX_SPAN_TAGS"

	// Set of default keys added to every span.
	SignalfxLibraryKey  = "signalfx.tracing.library"
	SignalfxLibraryValue= "go.tracing"
	// When cutting a new release, update the version value.
	SignalfxVersionKey   = "signalfx.tracing.version"
	SignalfxVersionValue = "v1.1.0"
)

var defaults = map[string]string{
	signalfxServiceName: "unnamed-go-service",
	signalfxEndpointURL: "http://localhost:9080/v1/trace",
	signalfxAccessToken: "",
}

type config struct {
	serviceName string
	accessToken string
	url         string
	// Because there can be multiple global tags added via environment variable
	// or calls to WithGlobalTag, store them in the required StartOption format to
	// call tracer.Start() in the variadic format.
	globalTags []tracer.StartOption
}

// StartOption is a function that configures an option for Start
type StartOption = func(*config)

func defaultConfig() *config {
	return &config{
		serviceName: envOrDefault(signalfxServiceName),
		accessToken: envOrDefault(signalfxAccessToken),
		url:         envOrDefault(signalfxEndpointURL),
		globalTags:  envDefaultAndGlobalTags(),
	}
}

// envDefaultAndGlobalTags extract global tags from the environment variable and parses the
// value in the expected format key1:value1, ...
// Also, it adds the Version and Library tags
func envDefaultAndGlobalTags() []tracer.StartOption {
	var globalTags []tracer.StartOption

	// Add the default Library and Version tags.
	defaultLibraryTag := tracer.WithGlobalTag(SignalfxLibraryKey, SignalfxLibraryValue)
	globalTags = append(globalTags, defaultLibraryTag)
	defaultVersionTag := tracer.WithGlobalTag(SignalfxVersionKey, SignalfxVersionValue)
	globalTags = append(globalTags, defaultVersionTag)

	// Add environment tags if they are passed in.
	var val string
	if val = os.Getenv(signalfxSpanTags); val == "" {
		return globalTags
	}

	tags := strings.Split(val, ",")
	for _, tag := range tags {
		// TODO: Currently this assumes "<stringb>" where "<stringa>:<stringb>" has no ":" in the
		// string. The TODO is to fix this logic to allow for "<stringb> to have colons, ":', in it.
		pair :=strings.Split(tag, ":")
		if len(pair) == 2 {
			key := strings.TrimSpace(pair[0])
			value := strings.TrimSpace(pair[1])
			// Empty keys aren't valid in Zipkin.
			// https://github.com/openzipkin/zipkin-api/blob/d3324ac79d1aa8f5c6f0ea4febb299402e50480f/zipkin-jsonv2.proto#L50-L51
			if key == "" {
				continue
			}
			globalTag := tracer.WithGlobalTag(key, value)
			globalTags = append(globalTags, globalTag)
		}
	}

	return globalTags
}

// envOrDefault gets the given environment variable if set otherwise a default value.
func envOrDefault(envVar string) string {
	if val := os.Getenv(envVar); val != "" {
		return val
	}
	return defaults[envVar]
}

// WithServiceName changes the reported service name
func WithServiceName(serviceName string) StartOption {
	return func(c *config) {
		c.serviceName = serviceName
	}
}

// WithAccessToken configures the SignalFx access token to use when reporting
func WithAccessToken(accessToken string) StartOption {
	return func(c *config) {
		c.accessToken = accessToken
	}
}

// WithEndpointURL configures the URL to send traces to
func WithEndpointURL(url string) StartOption {
	return func(c *config) {
		c.url = url
	}
}

// WithGlobalTag sets a tag with the given key/value on all spans created by the
// tracer. This option may be used multiple times.
// Note: Since the underlying transport is Zipkin, only values with strings
// are accepted.
func WithGlobalTag(k string, v string) StartOption {
	return func(c *config) {
		globalTag := tracer.WithGlobalTag(k, v)
		c.globalTags = append(c.globalTags, globalTag)
	}
}

// Start tracing globally
func Start(opts ...StartOption) {
	c := defaultConfig()
	for _, fn := range opts {
		fn(c)
	}

	startOptions := append(c.globalTags, tracer.WithServiceName(c.serviceName))
	startOptions = append(startOptions, tracer.WithZipkin(c.serviceName, c.url, c.accessToken))
	tracer.Start(
		startOptions...
	)

	opentracing.SetGlobalTracer(opentracer.New())
}

// Stop tracing globally
func Stop() {
	tracer.Stop()
	opentracing.SetGlobalTracer(&opentracing.NoopTracer{})
}
