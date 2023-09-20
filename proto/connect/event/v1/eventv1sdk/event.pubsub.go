package eventv1sdk

import (
	context "context"
	fmt "fmt"
	slog "log/slog"

	pubsub "cloud.google.com/go/pubsub"
	connect "connectrpc.com/connect"
	pubsubv1 "github.com/connect-sdk/pubsub-api/proto/connect/pubsub/v1"
	slogr "github.com/ralch/slogr"
	option "google.golang.org/api/option"

	eventv1 "github.com/connect-sdk/event-api/proto/connect/event/v1"
)

var _ pubsubv1.PubsubService = &EventPubsubService{}

// EventPubsubService is an implementation of the google.pubsub.v1.EventPubsubService service.
type EventPubsubService struct {
	// EventService contains an instance of cloud.event.v1.EventService service.
	EventService eventv1.EventService
}

// PushPubsubMessage implements google.pubsub.v1.PubsubService.
func (x *EventPubsubService) PushPubsubMessage(ctx context.Context, r *pubsubv1.PushPubsubMessageRequest) (*pubsubv1.PushPubsubMessageResponse, error) {
	// prepare the arg
	args := &eventv1.PushEventRequest{
		Event: &eventv1.Event{},
	}

	// set the event attributes
	if err := args.SetAttributes(r.Message.Attributes); err != nil {
		return nil, err
	}

	// set the event data
	if err := args.SetData(r.Message.Data); err != nil {
		return nil, err
	}

	// push the event
	if _, err := x.EventService.PushEvent(ctx, args); err != nil {
		return nil, err
	}

	response := &pubsubv1.PushPubsubMessageResponse{}
	// done!
	return response, nil
}

var (
	// ErrMissingTopic is returned by NewPubsubEventServiceClient when the topic argument is not provided.
	ErrMissingTopic = fmt.Errorf("no topic")
	// ErrMissingProject is returned by NewPubsubEventServiceClient when the project argument is not provided.
	ErrMissingProject = fmt.Errorf("no project")
)

// PubsubEventServiceClientConfig represents a configuration for the cloud.event.v1.PubsubServiceClient client.
type PubsubEventServiceClientConfig struct {
	// Project is the Google Cloud Project
	Project string
	// Topic is the Google Pub/Sub Topic
	Topic string
	// Options contains the client Options
	Options []option.ClientOption
}

var _ eventv1.EventServiceClient = &PubsubEventServiceClient{}

// EventServiceConn is a client for the cloud.event.v1.EventService service.
type PubsubEventServiceClient struct {
	client *pubsub.Client
	topic  string
}

// NewPubsubEventServiceClient creates a new cloud.event.v1.EventServiceClient client.
func NewPubsubEventServiceClient(ctx context.Context, config *PubsubEventServiceClientConfig) (eventv1.EventServiceClient, error) {
	if config == nil {
		return &eventv1.NopEventServiceClient{}, nil
	}

	if config.Project == "" {
		return nil, ErrMissingProject
	}

	if config.Topic == "" {
		return nil, ErrMissingTopic
	}

	// prepare the client
	client, err := pubsub.NewClient(ctx, config.Project, config.Options...)
	if err != nil {
		return nil, err
	}

	// prepare the broker
	connector := &PubsubEventServiceClient{
		topic:  config.Topic,
		client: client,
	}

	// done!
	return connector, nil
}

// PushEvent pushes a given event to cloud.event.v1.EventService service.
func (x *PubsubEventServiceClient) PushEvent(ctx context.Context, r *eventv1.PushEventRequest) (*eventv1.PushEventResponse, error) {
	// prepare the message
	message := &pubsub.Message{
		Data:        r.GetData(),
		Attributes:  r.GetAttributes(),
		OrderingKey: r.GetOrderingKey(),
	}

	// prepare the logger attr
	attr := slog.Group("event",
		slog.String("id", r.Event.Id),
		slog.String("type", r.Event.GetType()),
		slog.String("source", r.Event.GetSource()),
		slog.String("subject", r.Event.GetSubject()),
	)

	logger := slogr.FromContext(ctx)
	// prepare the logger message
	logger.Info("push an event", attr)

	topic := x.client.Topic(x.topic)
	topic.EnableMessageOrdering = true
	// publish the message
	if _, err := topic.Publish(ctx, message).Get(ctx); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &eventv1.PushEventResponse{}
	// done!
	return response, nil
}
