package eventv1sdk

import (
	context "context"
	http "net/http"
	time "time"

	connect "connectrpc.com/connect"
	interceptor "github.com/connect-sdk/interceptor"
	middleware "github.com/connect-sdk/middleware"
	chi "github.com/go-chi/chi/v5"
	metadata "google.golang.org/grpc/metadata"

	eventv1 "github.com/connect-sdk/event-api/proto/connect/event/v1"
	eventv1connect "github.com/connect-sdk/event-api/proto/connect/event/v1/eventv1connect"
)

var _ eventv1.EventServiceClient = &EventServiceClient{}

// EventServiceClient is a client for the cloud.event.v1.EventService service.
type EventServiceClient struct {
	client eventv1connect.EventServiceClient
}

// NewEventServiceClient creates a new cloud.event.v1.EventServiceClient client.
func NewEventServiceClient(uri string, options ...connect.ClientOption) eventv1.EventServiceClient {
	// prepare the options
	options = append(options, interceptor.WithContext())
	options = append(options, interceptor.WithTracer())
	options = append(options, interceptor.WithLogger())
	// prepare the clinet
	client := &EventServiceClient{
		client: eventv1connect.NewEventServiceClient(http.DefaultClient, uri, options...),
	}

	return client
}

// PushEventEvent implements cloud.event.v1.EventServiceClient.
func (x *EventServiceClient) PushEvent(ctx context.Context, r *eventv1.PushEventRequest) (*eventv1.PushEventResponse, error) {
	response, err := x.client.PushEvent(ctx, connect.NewRequest(r))
	if err != nil {
		return nil, err
	}
	// done
	return response.Msg, nil
}

var _ eventv1connect.EventServiceHandler = &EventServiceHandler{}

// EventServiceHandler represents an instance of cloud.event.v1.EventServiceHandler handler.
type EventServiceHandler struct {
	// EventService contains an instance of cloud.event.v1.EventService service.
	EventService eventv1.EventService
}

// Mount mounts the controller to a given router.
func (x *EventServiceHandler) Mount(r chi.Router) {
	var options []connect.HandlerOption
	// prepare the options
	options = append(options, interceptor.WithContext())
	options = append(options, interceptor.WithTracer())
	options = append(options, interceptor.WithLogger())
	options = append(options, interceptor.WithRecovery())
	options = append(options, interceptor.WithValidator())

	r.Group(func(r chi.Router) {
		// mount the middleware
		r.Use(middleware.WithLogger())
		// create the handler
		path, handler := eventv1connect.NewEventServiceHandler(x, options...)
		// mount the handler
		r.Mount(path, handler)
	})
}

// PushEvent pushes a given event to connect.runtime.v1.EventService service.
func (x *EventServiceHandler) PushEvent(ctx context.Context, r *connect.Request[eventv1.PushEventRequest]) (*connect.Response[eventv1.PushEventResponse], error) {
	response, err := x.EventService.PushEvent(ctx, r.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(response), nil
}

var _ eventv1.EventService = &EventService{}

// EventService represents a handler of cloud.event.v1.EventService service.
type EventService struct {
	// EventService contains an instance of cloud.event.v1.EventHandler handler.
	EventHandler eventv1.EventHandler
}

// PushEvent implements eventv1.EventService.
func (x *EventService) PushEvent(ctx context.Context, r *eventv1.PushEventRequest) (*eventv1.PushEventResponse, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		meta = metadata.Pairs()
	}

	// add the event attributes to the context
	meta.Append("ce-id", r.Event.GetId())
	meta.Append("ce-type", r.Event.GetType())
	meta.Append("ce-source", r.Event.GetSource())
	meta.Append("ce-subject", r.Event.GetSubject())
	meta.Append("ce-dataschema", r.Event.GetDataSchema())
	meta.Append("ce-specversion", r.Event.GetSpecVersion())
	meta.Append("ce-datacontenttype", r.Event.GetDataContentType())
	meta.Append("ce-time", r.Event.GetTime().Format(time.RFC3339))
	// override the context
	ctx = metadata.NewOutgoingContext(ctx, meta)

	// push the event
	if err := x.EventHandler.HandleEvent(ctx, r.Event); err != nil {
		return nil, err
	}

	return &eventv1.PushEventResponse{}, nil
}
