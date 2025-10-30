package models

import (
	"github.com/google/uuid"
)

//
// === Базовые справочники ===
//

type Directions string

const (
	DirectionIn  Directions = "In"
	DirectionOut Directions = "Out"
)

type ValueType string

const (
	ValueTypeString    ValueType = "String"
	ValueTypeDate      ValueType = "Date"
	ValueTypeInteger   ValueType = "Integer"
	ValueTypeBoolean   ValueType = "Boolean"
	ValueTypeNull      ValueType = "Null"
	ValueTypeStructure ValueType = "Structure"
	ValueTypeArray     ValueType = "Array"
)

type FileFormat string

const (
	FileFormatXML  FileFormat = "XML"
	FileFormatJSON FileFormat = "JSON"
	FileFormatDBF  FileFormat = "DBF"
	FileFormatCSV  FileFormat = "CSV"
	FileFormatTXT  FileFormat = "TXT"
)

type RoutineType string

const (
	RoutineBefore RoutineType = "Before"
	RoutineAfter  RoutineType = "After"
)

type MessageConvertType string

const (
	ConvertMultiplex MessageConvertType = "Multiplex"
	ConvertSplit     MessageConvertType = "Split"
	ConvertNone      MessageConvertType = "None"
)

type ProtocolType string

const (
	ProtocolTCP  ProtocolType = "TCP"
	ProtocolREST ProtocolType = "REST"
	ProtocolSOAP ProtocolType = "SOAP"
	ProtocolAMQP ProtocolType = "AMQP"
)

type MessageBrokerType string

const (
	BrokerKafka  MessageBrokerType = "Kafka"
	BrokerRabbit MessageBrokerType = "Rabbit"
)

type AuthenticationType string

const (
	AuthBasic       AuthenticationType = "Basic"
	AuthBearerToken AuthenticationType = "BearerToken"
)

type RestMethod string

const (
	MethodGet    RestMethod = "Get"
	MethodPost   RestMethod = "Post"
	MethodPatch  RestMethod = "Patch"
	MethodPut    RestMethod = "Put"
	MethodDelete RestMethod = "Delete"
)

//
// === Основные сущности ===
//

type User struct {
	Ref      uuid.UUID `db:"ref" json:"ref"`
	Username string    `db:"username" json:"username"`
	Password string    `db:"password" json:"password"`
}

type Global struct {
	Name  string      `db:"name" json:"name"`
	Value interface{} `db:"value" json:"value"`
}

type System struct {
	Ref  uuid.UUID `db:"ref" json:"ref"`
	Name string    `db:"name" json:"name"`
}

type Route struct {
	Ref    uuid.UUID  `db:"ref" json:"ref"`
	Name   string     `db:"name" json:"name"`
	Path   string     `db:"path" json:"path"`
	System uuid.UUID  `db:"system" json:"system"`
	Method RestMethod `db:"method" json:"method"`
}

type ConnectionSetting struct {
	Ref     uuid.UUID `db:"ref" json:"ref"`
	Name    string    `db:"name" json:"name"`
	System  uuid.UUID `db:"system" json:"system"`
	Path    string    `db:"path" json:"path"`
	Port    int       `db:"port" json:"port"`
	AuthRef uuid.UUID `db:"auth" json:"auth"`
}

type ConnectionAuthentication struct {
	Ref      uuid.UUID          `db:"ref" json:"ref"`
	Name     string             `db:"name" json:"name"`
	System   uuid.UUID          `db:"system" json:"system"`
	Type     AuthenticationType `db:"type" json:"type"`
	Username string             `db:"username" json:"username"`
	Password string             `db:"password" json:"password"`
	Token    string             `db:"token" json:"token"`
}

type ThreadObject struct {
	Ref        uuid.UUID  `db:"ref" json:"ref"`
	Name       string     `db:"name" json:"name"`
	NameObject string     `db:"name_object" json:"name_object"`
	Type       ValueType  `db:"type" json:"type"`
	Parent     *uuid.UUID `db:"parent" json:"parent,omitempty"`
}

type Routine struct {
	Ref  uuid.UUID   `db:"ref" json:"ref"`
	Name string      `db:"name" json:"name"`
	Type RoutineType `db:"type" json:"type"`
	Code string      `db:"code" json:"code"`
}

type ThreadGroup struct {
	Ref           uuid.UUID         `db:"ref" json:"ref"`
	Name          string            `db:"name" json:"name"`
	Protocol      ProtocolType      `db:"protocol" json:"protocol"`
	Parent        *uuid.UUID        `db:"parent" json:"parent,omitempty"`
	MessageBroker MessageBrokerType `db:"message_broker" json:"message_broker"`
}

type Thread struct {
	Ref                uuid.UUID          `db:"ref" json:"ref"`
	Name               string             `db:"name" json:"name"`
	Group              uuid.UUID          `db:"group" json:"group"`
	MessageConvertType MessageConvertType `db:"message_convert_type" json:"message_convert_type"`
}

type ThreadRoute struct {
	Thread     uuid.UUID  `db:"thread" json:"thread"`
	Direction  Directions `db:"direction" json:"direction"`
	Route      uuid.UUID  `db:"route" json:"route"`
	FileFormat FileFormat `db:"file_format" json:"file_format"`
	Object     uuid.UUID  `db:"object" json:"object"`
	Routine    uuid.UUID  `db:"routine" json:"routine"`
}
