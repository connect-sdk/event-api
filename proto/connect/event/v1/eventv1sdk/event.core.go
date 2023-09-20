package eventv1sdk

import (
	context "context"
	slog "log/slog"
	http "net/http"

	connect "connectrpc.com/connect"
	idtoken "google.golang.org/api/idtoken"
)

// EventServiceClientConfig represents the config for cloud.event.v1.EventServiceClient client.
type EventServiceClientConfig struct {
	Client        *http.Client
	ClientURL     string
	ClientOptions []connect.ClientOption
}

// EventServiceClientOption represents an option for cloud.event.v1.EventServiceClient client.
type EventServiceClientOption interface {
	// Apply applies the option.
	Apply(config *EventServiceClientConfig)
}

var _ EventServiceClientOption = EventServiceClientOptionFunc(nil)

// EventServiceClientOptionFunc represent a function that implementes cloud.event.v1.EventServiceClientOption option.
type EventServiceClientOptionFunc func(*EventServiceClientConfig)

// Apply applies the option.
func (fn EventServiceClientOptionFunc) Apply(config *EventServiceClientConfig) {
	fn(config)
}

// EventServiceClientWithAuthorization enables an oidc authorization.
func EventServiceClientWithAuthorization() EventServiceClientOption {
	fn := func(config *EventServiceClientConfig) {
		// client uri
		uri := config.ClientURL
		// prepare the options
		options := []idtoken.ClientOption{}
		options = append(options, idtoken.WithHTTPClient(config.Client))
		// prepare the client
		client, err := idtoken.NewClient(context.Background(), uri, options...)
		if err != nil {
			slog.Error("unable to create an id token", err)
		}
		// set the client
		config.Client = client
	}

	return EventServiceClientOptionFunc(fn)
}

// EventServiceClientWithProtocol enables a given protocol.
func EventServiceClientWithProtocol(name string) EventServiceClientOption {
	fn := func(config *EventServiceClientConfig) {
		// prepare the protocol
		switch name {
		case "grpc":
			config.ClientOptions = append(config.ClientOptions, connect.WithGRPC())
		case "grpc+web":
			config.ClientOptions = append(config.ClientOptions, connect.WithGRPCWeb())
		}
	}

	return EventServiceClientOptionFunc(fn)
}

// EventServiceClientWithCodec enables a given codec.
func EventServiceClientWithCodec(name string) EventServiceClientOption {
	fn := func(config *EventServiceClientConfig) {
		// prepare the protocol
		switch name {
		case "json":
			config.ClientOptions = append(config.ClientOptions, connect.WithProtoJSON())
		}
	}

	return EventServiceClientOptionFunc(fn)
}
