package eventv1sdk

import (
	context "context"
	fmt "fmt"
	slog "log/slog"
	time "time"

	pubsub "cloud.google.com/go/pubsub"
	connect "connectrpc.com/connect"
	pubsubv1 "github.com/connect-sdk/pubsub-api/proto/google/pubsub/v1"
	slogr "github.com/ralch/slogr"
	metadata "google.golang.org/grpc/metadata"

	eventv1 "github.com/connect-sdk/event-api/proto/cloud/event/v1"
)

var _ pubsubv1.PubsubService = &PubsubService{}

// PubsubService is an implementation of the google.pubsub.v1.PubsubService service.
type PubsubService struct {
	// EventService contains an instance of cloud.event.v1.EventService service.
	EventService eventv1.EventService
}

// PushPubsubMessage implements google.pubsub.v1.PubsubService.
func (x *PubsubService) PushPubsubMessage(ctx context.Context, r *pubsubv1.PushPubsubMessageRequest) (*pubsubv1.PushPubsubMessageResponse, error) {
	// prepare the argument
	argument := &eventv1.PushEventRequest{
		Event: &eventv1.Event{},
	}

	// set the event attributes
	if err := argument.SetAttributes(r.Message.Attributes); err != nil {
		return nil, err
	}

	// set the event data
	if err := argument.SetData(r.Message.Data); err != nil {
		return nil, err
	}

	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		meta = metadata.Pairs()
	}

	// add the event attributes to the context
	meta.Append("ce-id", argument.Event.GetId())
	meta.Append("ce-type", argument.Event.GetType())
	meta.Append("ce-source", argument.Event.GetSource())
	meta.Append("ce-subject", argument.Event.GetSubject())
	meta.Append("ce-dataschema", argument.Event.GetDataSchema())
	meta.Append("ce-specversion", argument.Event.GetSpecVersion())
	meta.Append("ce-datacontenttype", argument.Event.GetDataContentType())
	meta.Append("ce-time", argument.Event.GetTime().Format(time.RFC3339))
	// override the context
	ctx = metadata.NewOutgoingContext(ctx, meta)

	// push the event
	if _, err := x.EventService.PushEvent(ctx, argument); err != nil {
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

// PubsubServiceClientConfig represents a configuration for the cloud.event.v1.PubsubServiceClient client.
type PubsubServiceClientConfig struct {
	// Project is the Google Cloud Project
	Project string
	// Topic is the Google Pub/Sub Topic
	Topic string
}

var _ eventv1.EventServiceClient = &PubsubServiceClient{}

// EventServiceConn is a client for the cloud.event.v1.EventService service.
type PubsubServiceClient struct {
	client *pubsub.Client
	topic  string
}

// NewPubsubServiceClient creates a new cloud.event.v1.EventServiceClient client.
func NewPubsubServiceClient(ctx context.Context, config *PubsubServiceClientConfig) (eventv1.EventServiceClient, error) {
	if config.Project == "" {
		return nil, ErrMissingProject
	}

	if config.Topic == "" {
		return nil, ErrMissingTopic
	}

	// prepare the client
	client, err := pubsub.NewClient(ctx, config.Project)
	if err != nil {
		return nil, err
	}

	// prepare the broker
	connector := &PubsubServiceClient{
		topic:  config.Topic,
		client: client,
	}

	// done!
	return connector, nil
}

// PushEvent pushes a given event to cloud.event.v1.EventService service.
func (x *PubsubServiceClient) PushEvent(ctx context.Context, r *eventv1.PushEventRequest) (*eventv1.PushEventResponse, error) {
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
