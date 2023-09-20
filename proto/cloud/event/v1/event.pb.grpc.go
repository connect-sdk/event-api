package eventv1

import (
	"context"
)

//go:generate counterfeiter -generate

//counterfeiter:generate -o ./eventv1fake . EventService

// The service that an application uses to consume events from a subscription via the Push method.
type EventService interface {
	// PushEvent pushes a given event to connect.runtime.v1.EventService service.
	PushEvent(context.Context, *PushEventRequest) (*PushEventResponse, error)
}

//counterfeiter:generate -o ./eventv1fake . EventServiceClient

// The service client that an application uses to consume events from a subscription via the Push method.
type EventServiceClient interface {
	// PushEvent pushes a given event to connect.runtime.v1.EventService service.
	PushEvent(context.Context, *PushEventRequest) (*PushEventResponse, error)
}
