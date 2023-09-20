package eventv1

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// NewEvent returns a new instance of cloud.event.v1.Event message.
func NewEvent() *Event {
	id, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}

	event := &Event{
		Id:          id.String(),
		Attributes:  make(map[string]*EventAttributeValue),
		SpecVersion: "1.0",
	}

	return event
}

// SetId sets the ID attribute.
func (x *Event) SetId(value string) {
	x.Id = value
}

// GetSubject returns the Subject attribute. It describes the subject of the
// event in the context of the event producer (identified by source). In
// publish-subscribe scenarios, a subscriber will typically subscribe to events
// emitted by a source, but the source identifier alone might not be sufficient
// as a qualifier for any specific event if the source context has internal
// sub-structure.
func (x *Event) GetSubject() string {
	if attr, ok := x.Attributes["subject"]; ok {
		return attr.GetCeString()
	}

	return ""
}

// SetSubject sets the Subject attribute.
func (x *Event) SetSubject(value string) {
	x.Attributes["subject"] = &EventAttributeValue{
		Attr: &EventAttributeValue_CeString{
			CeString: value,
		},
	}
}

// SetType sets the Type attribute.
func (x *Event) SetType(value string) {
	x.Type = value
}

// SetSource sets the Source attribute.
func (x *Event) SetSource(value string) {
	x.Source = value
}

// SetSpecVersion sets the SpecVersion attribute.
func (x *Event) SetSpecVersion(value string) {
	x.SpecVersion = value
}

// SetTime sets the Time attribute.
func (x *Event) SetTime(value time.Time) {
	x.Attributes["time"] = &EventAttributeValue{
		Attr: &EventAttributeValue_CeTimestamp{
			CeTimestamp: timestamppb.New(value),
		},
	}
}

// GetTime returns the Time attribute. Timestamp of when the occurrence
// happened. If the time of the occurrence cannot be determined then this
// attribute MAY be set to some other time (such as the current time) by the
// CloudEvents producer, however all producers for the same source MUST be
// consistent in this respect. In other words, either they all use the actual
// time of the occurrence or they all use the same algorithm to determine the
// value used.
func (x *Event) GetTime() time.Time {
	if attr, ok := x.Attributes["time"]; ok {
		if value := attr.GetCeTimestamp(); value != nil {
			return value.AsTime()
		}
	}

	return time.Time{}
}

// GetDataSchema returns the DataSchema attribute.  It contains the schema that
// data adheres to. Incompatible changes to the schema SHOULD be reflected by a
// different URI.
func (x *Event) GetDataSchema() string {
	if attr, ok := x.Attributes["dataschema"]; ok {
		return attr.GetCeUri()
	}

	return ""
}

// SetDataSchema sets the DataSchema attribute.
func (x *Event) SetDataSchema(value string) {
	if v, err := url.Parse(value); err == nil {
		x.Attributes["dataschema"] = &EventAttributeValue{
			Attr: &EventAttributeValue_CeUri{
				CeUri: v.String(),
			},
		}
	}
}

// GetDataContentType returns the DataContentType attribute.
func (x *Event) GetDataContentType() string {
	if attr, ok := x.Attributes["datacontenttype"]; ok {
		return attr.GetCeString()
	}

	return ""
}

// SetDataContentType sets the DataContentType attribute.
func (x *Event) SetDataContentType(value string) {
	x.Attributes["datacontenttype"] = &EventAttributeValue{
		Attr: &EventAttributeValue_CeString{
			CeString: value,
		},
	}
}

// SetExtension sets the Extension attribute.
func (x *Event) SetExtension(name string, value interface{}) {
	attributes := x.Attributes

	switch v := value.(type) {
	case bool:
		value = &EventAttributeValue_CeBoolean{
			CeBoolean: v,
		}
	case string:
		value = &EventAttributeValue_CeString{
			CeString: v,
		}
	case int32:
		value = &EventAttributeValue_CeInteger{
			CeInteger: v,
		}
	case []byte:
		value = &EventAttributeValue_CeString{
			CeString: base64.RawStdEncoding.EncodeToString(v),
		}
	case *url.URL:
		value = &EventAttributeValue_CeString{
			CeString: v.String(),
		}
	case time.Time:
		value = &EventAttributeValue_CeTimestamp{
			CeTimestamp: timestamppb.New(v),
		}
	case *timestamppb.Timestamp:
		value = &EventAttributeValue_CeTimestamp{
			CeTimestamp: v,
		}
	}

	switch attr := value.(type) {
	case *EventAttributeValue_CeBoolean:
		attributes[name] = &EventAttributeValue{
			Attr: attr,
		}
	case *EventAttributeValue_CeBytes:
		attributes[name] = &EventAttributeValue{
			Attr: attr,
		}
	case *EventAttributeValue_CeInteger:
		attributes[name] = &EventAttributeValue{
			Attr: attr,
		}
	case *EventAttributeValue_CeString:
		attributes[name] = &EventAttributeValue{
			Attr: attr,
		}
	case *EventAttributeValue_CeTimestamp:
		attributes[name] = &EventAttributeValue{
			Attr: attr,
		}
	case *EventAttributeValue_CeUri:
		attributes[name] = &EventAttributeValue{
			Attr: attr,
		}
	case *EventAttributeValue_CeUriRef:
		attributes[name] = &EventAttributeValue{
			Attr: attr,
		}
	}
}

// GetDataAs attempts to populate the provided data object with the event
// payload. The object should be a pointer type.
func (x *Event) GetDataAs(value interface{}) error {
	ctype := x.GetDataContentType()

	switch {
	case strings.HasPrefix(ctype, "text"):
		data := x.GetTextData()
		// unmarshal the data
		value := reflect.ValueOf(value)
		value.Elem().SetString(data)
		// done!
		return nil
	case strings.EqualFold(ctype, "application/json"):
		data := x.GetBinaryData()
		// unmarshal the data
		return json.Unmarshal(data, value)
	case strings.EqualFold(ctype, "application/cloudevents+protobuf"):
		if value, ok := value.(proto.Message); ok {
			// unmarshal the data
			return x.GetProtoData().UnmarshalTo(value)
		}
	}

	return fmt.Errorf("cannot get the data with content-type %v as data-type %T", ctype, value)
}

// SetData encodes the given payload with the given content type. If the
// provided payload is a byte array, when marshalled to json it will be encoded
// as base64. If the provided payload is different from byte array,
// datacodec.Encode is invoked to attempt a marshalling to byte array.
func (x *Event) SetData(value interface{}) error {
	switch data := value.(type) {
	case string:
		x.SetDataContentType("text/plain")
		// set the data
		x.Data = &Event_TextData{
			TextData: data,
		}
	case []byte:
		x.SetDataContentType("application/octet-stream")
		// set the data
		x.Data = &Event_BinaryData{
			BinaryData: data,
		}
	case json.Marshaler:
		x.SetDataContentType("application/json")
		// marshal the payload
		payload, err := json.Marshal(data)
		if err != nil {
			return err
		}
		// set the data
		x.Data = &Event_BinaryData{
			BinaryData: payload,
		}

		// If the provided payload is a proto.Message, when marshalled to json it will use the TypeUrl as a data schema.
		if data, ok := data.(proto.Message); ok {
			if message, err := anypb.New(data); err == nil {
				x.SetDataSchema(message.TypeUrl)
			}
		}
	case proto.Message:
		message, ok := data.(*anypb.Any)

		if !ok {
			var err error
			// create a new entity
			message, err = anypb.New(data)
			if err != nil {
				return err
			}
		}

		x.SetDataSchema(message.TypeUrl)
		x.SetDataContentType("application/cloudevents+protobuf")
		// set the data
		x.Data = &Event_ProtoData{
			ProtoData: message,
		}
	default:
		return fmt.Errorf("cannot set the data with data-type %T", value)
	}

	return nil
}

// GetOrderingKey returns the ordering key.
func (x *PushEventRequest) GetOrderingKey() string {
	return x.Event.GetId()
}

// GetAttributes returns the attributes.
func (x *PushEventRequest) GetAttributes() map[string]string {
	// WithPrefix returns the key with a prefix.
	WithPrefix := func(key string) string {
		return strings.ToLower("ce-" + key)
	}

	attributes := make(map[string]string)
	attributes[WithPrefix("id")] = x.Event.GetId()
	attributes[WithPrefix("type")] = x.Event.GetType()
	attributes[WithPrefix("source")] = x.Event.GetSource()
	attributes[WithPrefix("specversion")] = x.Event.GetSpecVersion()

	for name, attribute := range x.Event.GetAttributes() {
		// prepare the name
		name = WithPrefix(name)
		// prepare the value
		switch attr := attribute.Attr.(type) {
		case *EventAttributeValue_CeBoolean:
			attributes[name] = strconv.FormatBool(attr.CeBoolean)
		case *EventAttributeValue_CeInteger:
			attributes[name] = strconv.FormatInt(int64(attr.CeInteger), 10)
		case *EventAttributeValue_CeBytes:
			attributes[name] = base64.StdEncoding.EncodeToString(attr.CeBytes)
		case *EventAttributeValue_CeUri:
			attributes[name] = attr.CeUri
		case *EventAttributeValue_CeUriRef:
			attributes[name] = attr.CeUriRef
		case *EventAttributeValue_CeTimestamp:
			attributes[name] = attr.CeTimestamp.AsTime().UTC().Format(time.RFC3339Nano)
		case *EventAttributeValue_CeString:
			attributes[name] = attr.CeString
		}
	}

	return attributes
}

// SetAttributes sets the attributes.
func (x *PushEventRequest) SetAttributes(attributes map[string]string) error {
	// WithPrefix returns the key without a prefix.
	WithoutPrefix := func(key string) string {
		key = strings.ToLower(key)
		key = strings.TrimPrefix(key, "ce-")
		return key
	}

	for name, value := range attributes {
		// preapre the name
		name = WithoutPrefix(name)
		// prepare the value
		switch name {
		case "id":
			x.Event.SetId(value)
		case "type":
			x.Event.SetType(value)
		case "subject":
			x.Event.SetSubject(value)
		case "source":
			x.Event.SetSource(value)
		case "specversion":
			x.Event.SetSpecVersion(value)
		case "dataschema":
			x.Event.SetDataSchema(value)
		case "datacontenttype":
			x.Event.SetDataContentType(value)
		case "time":
			timestamp, err := time.Parse(time.RFC3339Nano, value)
			if err != nil {
				return err
			}
			// set the value
			x.Event.SetTime(timestamp)
		default:
			x.Event.SetExtension(name, value)
		}
	}

	return nil
}

// GetData returns the data.
func (x *PushEventRequest) GetData() []byte {
	switch payload := x.Event.GetData().(type) {
	case *Event_TextData:
		return []byte(payload.TextData)
	case *Event_ProtoData:
		data, _ := protojson.Marshal(payload.ProtoData)
		return data
	case *Event_BinaryData:
		return payload.BinaryData
	default:
		return nil
	}
}

// SetData sets the data.
func (x *PushEventRequest) SetData(data []byte) error {
	ctype := x.Event.GetDataContentType()

	switch {
	case strings.EqualFold(ctype, "application/cloudevents+protobuf"):
		entity := &anypb.Any{}
		// unmarshal the entity
		if err := protojson.Unmarshal(data, entity); err != nil {
			return err
		}
		// set the data
		if err := x.Event.SetData(entity); err != nil {
			return err
		}
	case strings.HasPrefix(ctype, "text"):
		// set the data
		if err := x.Event.SetData(string(data)); err != nil {
			return err
		}
	default:
		// set the data
		if err := x.Event.SetData(data); err != nil {
			return err
		}
	}

	return nil
}

// GetEventID returns the event id from the incoming context.
func GetEventID(ctx context.Context) string {
	if values := metadata.ValueFromIncomingContext(ctx, "ce-id"); len(values) > 0 {
		return values[0]
	}

	return ""
}

// GetEventType returns the event type from the incoming context.
func GetEventType(ctx context.Context) string {
	if values := metadata.ValueFromIncomingContext(ctx, "ce-type"); len(values) > 0 {
		return values[0]
	}

	return ""
}

// GetEventSource returns the event source from the incoming context.
func GetEventSource(ctx context.Context) string {
	if values := metadata.ValueFromIncomingContext(ctx, "ce-source"); len(values) > 0 {
		return values[0]
	}

	return ""
}

// GetEventSubject returns the event subject from the incoming context.
func GetEventSubject(ctx context.Context) string {
	if values := metadata.ValueFromIncomingContext(ctx, "ce-subject"); len(values) > 0 {
		return values[0]
	}

	return ""
}

// GetEventDataSchema returns the event data schema from the incoming context.
func GetEventDataSchema(ctx context.Context) string {
	if values := metadata.ValueFromIncomingContext(ctx, "ce-dataschema"); len(values) > 0 {
		return values[0]
	}

	return ""
}

// GetEventDataContentType returns the event data schema from the incoming context.
func GetEventDataContentType(ctx context.Context) string {
	if values := metadata.ValueFromIncomingContext(ctx, "ce-datacontenttype"); len(values) > 0 {
		return values[0]
	}

	return ""
}

// GetEventSpecVersion returns the event spec version from the incoming context.
func GetEventSpecVersion(ctx context.Context) string {
	if values := metadata.ValueFromIncomingContext(ctx, "ce-specversion"); len(values) > 0 {
		return values[0]
	}

	return ""
}

// GetEventTime returns the event time from the incoming context.
func GetEventTime(ctx context.Context) time.Time {
	if values := metadata.ValueFromIncomingContext(ctx, "ce-time"); len(values) > 0 {
		tvalue, _ := time.Parse(time.RFC3339, values[0])
		return tvalue
	}

	return time.Time{}
}
