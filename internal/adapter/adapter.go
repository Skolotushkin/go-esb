package adapter

import (
	"context"
	"fmt"

	"go-esb/internal/models"
)

// ProtocolAdapter интерфейс для адаптеров протоколов
type ProtocolAdapter interface {
	Send(ctx context.Context, endpoint string, action string, headers map[string]string, body []byte) ([]byte, int, error)
	Authenticate(auth *models.ConnectionAuthentication, endpoint string) (map[string]string, error)
}

// AdapterFactory создает адаптеры по типу протокола
type AdapterFactory struct {
	restAdapter *RESTAdapter
	soapAdapter *SOAPAdapter
	amqpAdapter *AMQPAdapter
}

// NewAdapterFactory создает фабрику адаптеров
func NewAdapterFactory() *AdapterFactory {
	return &AdapterFactory{
		restAdapter: NewRESTAdapter(),
		soapAdapter: NewSOAPAdapter(),
		amqpAdapter: NewAMQPAdapter(),
	}
}

// GetAdapter возвращает адаптер для указанного протокола
func (f *AdapterFactory) GetAdapter(protocol models.ProtocolType) (ProtocolAdapter, error) {
	switch protocol {
	case models.ProtocolREST:
		return f.restAdapter, nil
	case models.ProtocolSOAP:
		return f.soapAdapter, nil
	case models.ProtocolAMQP:
		return f.amqpAdapter, nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

