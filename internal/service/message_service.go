package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"go-esb/internal/adapter"
	"go-esb/internal/converter"
	"go-esb/internal/models"
	"go-esb/internal/repository"

	"github.com/google/uuid"
)

// MessageService обрабатывает маршрутизацию и трансформацию сообщений
type MessageService interface {
	ProcessMessage(ctx context.Context, threadID string, direction models.Directions, messageData []byte) error
	RouteMessage(ctx context.Context, threadID uuid.UUID, direction models.Directions, messageData []byte) error
}

type messageService struct {
	threadRouteRepo  repository.ThreadRouteRepository
	routeRepo        repository.RouteRepository
	connectionRepo   repository.ConnectionRepository
	systemRepo       repository.SystemRepository
	adapterFactory   *adapter.AdapterFactory
	formatConverter  *converter.Converter
}

func NewMessageService(
	threadRouteRepo repository.ThreadRouteRepository,
	routeRepo repository.RouteRepository,
	connectionRepo repository.ConnectionRepository,
	systemRepo repository.SystemRepository,
) MessageService {
	return &messageService{
		threadRouteRepo:  threadRouteRepo,
		routeRepo:        routeRepo,
		connectionRepo:   connectionRepo,
		systemRepo:       systemRepo,
		adapterFactory:   adapter.NewAdapterFactory(),
		formatConverter:  converter.NewConverter(),
	}
}

// ProcessMessage обрабатывает входящее сообщение через thread
func (s *messageService) ProcessMessage(ctx context.Context, threadID string, direction models.Directions, messageData []byte) error {
	threadUUID, err := uuid.Parse(threadID)
	if err != nil {
		return fmt.Errorf("invalid thread ID: %w", err)
	}

	return s.RouteMessage(ctx, threadUUID, direction, messageData)
}

// RouteMessage маршрутизирует сообщение по конфигурации thread
func (s *messageService) RouteMessage(ctx context.Context, threadID uuid.UUID, direction models.Directions, messageData []byte) error {
	// Получаем thread и group для определения протокола
	thread, group, err := s.threadRouteRepo.GetThreadWithGroup(ctx, threadID)
	if err != nil {
		return fmt.Errorf("failed to get thread: %w", err)
	}

	// Получаем маршруты для данного направления
	routes, err := s.threadRouteRepo.GetThreadRouteByDirection(ctx, threadID, direction)
	if err != nil {
		return fmt.Errorf("failed to get routes: %w", err)
	}

	if len(routes) == 0 {
		return fmt.Errorf("no routes found for thread %s with direction %s", threadID, direction)
	}

	// Обрабатываем каждый маршрут
	for _, threadRoute := range routes {
		if err := s.processRoute(ctx, thread, group, threadRoute, messageData); err != nil {
			log.Printf("⚠️ Error processing route %s: %v", threadRoute.Route, err)
			// Продолжаем обработку других маршрутов
		}
	}

	return nil
}

func (s *messageService) processRoute(
	ctx context.Context,
	thread *models.Thread,
	group *models.ThreadGroup,
	threadRoute models.ThreadRoute,
	messageData []byte,
) error {
	// Получаем route для получения информации о системе
	routeID := threadRoute.Route
	// Получаем route через repository (нужно добавить метод GetByID)
	// Пока используем прямой запрос через routeRepo

	// Получаем информацию о системе из route
	route, err := s.getRouteByID(ctx, routeID)
	if err != nil {
		return fmt.Errorf("failed to get route: %w", err)
	}

	// Получаем connection settings для системы
	connSettings, err := s.connectionRepo.GetConnectionSettings(ctx, route.System)
	if err != nil {
		return fmt.Errorf("failed to get connection settings: %w", err)
	}

	// Конвертируем формат данных если нужно
	convertedData := messageData
	if threadRoute.FileFormat != "JSON" {
		// Конвертируем из JSON в требуемый формат
		convertedData, err = s.formatConverter.Convert(messageData, "JSON", string(threadRoute.FileFormat))
		if err != nil {
			return fmt.Errorf("failed to convert format: %w", err)
		}
	}

	// Получаем адаптер для протокола
	protocolAdapter, err := s.adapterFactory.GetAdapter(group.Protocol)
	if err != nil {
		return fmt.Errorf("unsupported protocol: %w", err)
	}

	// Подготавливаем заголовки аутентификации
	var headers map[string]string
	if connSettings.AuthRef != uuid.Nil {
		auth, err := s.connectionRepo.GetConnectionAuth(ctx, connSettings.AuthRef)
		if err == nil {
			headers, _ = protocolAdapter.Authenticate(auth, connSettings.Path)
		}
	}

	// Формируем endpoint
	endpoint := s.buildEndpoint(connSettings, route)

	// Отправляем сообщение
	action := ""
	switch group.Protocol {
	case models.ProtocolSOAP:
		// SOAP action может быть из route или конфигурации
		action = route.Path
	case models.ProtocolREST:
		// Для REST action содержит HTTP method
		action = string(route.Method)
	case models.ProtocolAMQP:
		// Для AMQP action содержит exchange name (если нужно)
		action = "" // или из конфигурации
	}
	_, statusCode, err := protocolAdapter.Send(ctx, endpoint, action, headers, convertedData)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("✅ Message sent to %s via %s (status: %d)", route.Name, group.Protocol, statusCode)
	return nil
}

func (s *messageService) getRouteByID(ctx context.Context, routeID uuid.UUID) (*models.Route, error) {
	return s.routeRepo.GetByID(ctx, routeID)
}

func (s *messageService) buildEndpoint(connSettings *models.ConnectionSetting, route *models.Route) string {
	// Базовая часть endpoint из connection settings
	basePath := connSettings.Path
	if basePath == "" {
		basePath = route.Path
	}

	// Добавляем порт если указан
	if connSettings.Port > 0 && connSettings.Port != 80 && connSettings.Port != 443 {
		if !strings.Contains(basePath, ":") {
			// Определяем протокол
			if strings.HasPrefix(basePath, "https://") {
				if connSettings.Port != 443 {
					parts := strings.SplitN(basePath, "://", 2)
					if len(parts) == 2 {
						basePath = fmt.Sprintf("%s://%s:%d", parts[0], parts[1], connSettings.Port)
					}
				}
			} else if strings.HasPrefix(basePath, "http://") {
				if connSettings.Port != 80 {
					parts := strings.SplitN(basePath, "://", 2)
					if len(parts) == 2 {
						basePath = fmt.Sprintf("%s://%s:%d", parts[0], parts[1], connSettings.Port)
					}
				}
			}
		}
	}

	// Объединяем base path с route path если нужно
	if !strings.HasPrefix(route.Path, "http") && !strings.HasSuffix(basePath, route.Path) {
		basePath = strings.TrimSuffix(basePath, "/") + route.Path
	}

	return basePath
}

