package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go-esb/internal/models"
	"go-esb/internal/repository"

	"github.com/google/uuid"
)

// Orchestrator управляет бизнес-процессами (оркестрация)
type Orchestrator interface {
	ExecuteProcess(ctx context.Context, processName string, initialData []byte) error
}

type orchestrator struct {
	messageService   MessageService
	threadRouteRepo  repository.ThreadRouteRepository
	routeRepo        repository.RouteRepository
	connectionRepo   repository.ConnectionRepository
	systemRepo       repository.SystemRepository
}

// NewOrchestrator создает новый оркестратор
func NewOrchestrator(
	messageService MessageService,
	threadRouteRepo repository.ThreadRouteRepository,
	routeRepo repository.RouteRepository,
	connectionRepo repository.ConnectionRepository,
	systemRepo repository.SystemRepository,
) Orchestrator {
	return &orchestrator{
		messageService:   messageService,
		threadRouteRepo:  threadRouteRepo,
		routeRepo:        routeRepo,
		connectionRepo:   connectionRepo,
		systemRepo:       systemRepo,
	}
}

// OrderProcessingFlow представляет поток обработки заказа
type OrderProcessingFlow struct {
	StripePaymentData map[string]interface{} `json:"stripe_payment_data"`
	SAPOrderData      map[string]interface{} `json:"sap_order_data"`
	SalesforceData    map[string]interface{} `json:"salesforce_data"`
}

// ExecuteProcess выполняет бизнес-процесс
func (o *orchestrator) ExecuteProcess(ctx context.Context, processName string, initialData []byte) error {
	log.Printf("🎯 Starting process: %s", processName)

	switch processName {
	case "order_payment_flow":
		return o.orderPaymentFlow(ctx, initialData)
	default:
		return fmt.Errorf("unknown process: %s", processName)
	}
}

// orderPaymentFlow реализует поток: Stripe → SAP → Salesforce
func (o *orchestrator) orderPaymentFlow(ctx context.Context, stripeData []byte) error {
	flowStartTime := time.Now()
	log.Printf("🔄 Order Payment Flow started")

	flow := &OrderProcessingFlow{}

	// Шаг 1: Парсим данные от Stripe
	if err := json.Unmarshal(stripeData, &flow.StripePaymentData); err != nil {
		return fmt.Errorf("failed to parse Stripe data: %w", err)
	}

	log.Printf("📥 Received payment data from Stripe: %+v", flow.StripePaymentData)

	// Шаг 2: Находим thread для отправки в SAP
	// Предполагаем, что thread ID известен (можно получить из конфигурации)
	// Для примера используем поиск по имени системы
	
	// Получаем систему SAP (предполагаем, что она создана в БД)
	systems, err := o.systemRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get systems: %w", err)
	}

	var sapSystemID, salesforceSystemID uuid.UUID
	for _, sys := range systems {
		if sys.Name == "SAP" {
			sapSystemID = sys.Ref
		}
		if sys.Name == "Salesforce" {
			salesforceSystemID = sys.Ref
		}
	}

	if sapSystemID == uuid.Nil {
		return fmt.Errorf("SAP system not found")
	}

	// Шаг 3: Преобразуем данные Stripe в формат для SAP
	// Преобразуем JSON в XML для SAP SOAP
	flow.SAPOrderData = o.transformStripeToSAP(flow.StripePaymentData)
	
	sapData, err := json.Marshal(flow.SAPOrderData)
	if err != nil {
		return fmt.Errorf("failed to marshal SAP data: %w", err)
	}

	// Шаг 4: Отправляем в SAP через thread (нужно найти thread для SAP)
	// Пока используем прямую отправку через MessageService
	// В реальном сценарии нужно найти thread по конфигурации
	
	// Временно: создаем контекст с таймаутом для SAP запроса
	sapCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Ищем thread для отправки в SAP
	// Для примера создадим логику поиска thread
	threadID, err := o.findThreadForSystem(ctx, sapSystemID, models.DirectionOut)
	if err != nil {
		return fmt.Errorf("failed to find SAP thread: %w", err)
	}

	log.Printf("📤 Sending to SAP via thread: %s", threadID)
	if err := o.messageService.RouteMessage(sapCtx, threadID, models.DirectionOut, sapData); err != nil {
		return fmt.Errorf("failed to send to SAP: %w", err)
	}

	log.Printf("✅ SAP confirmed order update")

	// Шаг 5: После подтверждения SAP отправляем в Salesforce
	if salesforceSystemID == uuid.Nil {
		log.Printf("⚠️ Salesforce system not found, skipping")
		return nil
	}

	// Преобразуем данные для Salesforce
	flow.SalesforceData = o.transformSAPToSalesforce(flow.SAPOrderData)
	salesforceData, err := json.Marshal(flow.SalesforceData)
	if err != nil {
		return fmt.Errorf("failed to marshal Salesforce data: %w", err)
	}

	// Ищем thread для Salesforce
	salesforceThreadID, err := o.findThreadForSystem(ctx, salesforceSystemID, models.DirectionOut)
	if err != nil {
		return fmt.Errorf("failed to find Salesforce thread: %w", err)
	}

	salesforceCtx, cancel2 := context.WithTimeout(ctx, 5*time.Second)
	defer cancel2()

	log.Printf("📤 Sending to Salesforce via thread: %s", salesforceThreadID)
	if err := o.messageService.RouteMessage(salesforceCtx, salesforceThreadID, models.DirectionOut, salesforceData); err != nil {
		return fmt.Errorf("failed to send to Salesforce: %w", err)
	}

	duration := time.Since(flowStartTime)
	log.Printf("🎉 Order Payment Flow completed in %v", duration)

	if duration > 5*time.Second {
		log.Printf("⚠️ Flow took longer than 5 seconds target")
	}

	return nil
}

// transformStripeToSAP преобразует данные Stripe в формат SAP
func (o *orchestrator) transformStripeToSAP(stripeData map[string]interface{}) map[string]interface{} {
	sapData := make(map[string]interface{})
	
	// Извлекаем данные из Stripe и мапим на SAP формат
	if orderID, ok := stripeData["order_id"].(string); ok {
		sapData["OrderNumber"] = orderID
	}
	if amount, ok := stripeData["amount"].(float64); ok {
		sapData["Amount"] = amount / 100 // Stripe хранит в центах
	}
	if currency, ok := stripeData["currency"].(string); ok {
		sapData["Currency"] = currency
	}
	if status, ok := stripeData["status"].(string); ok {
		sapData["PaymentStatus"] = status
	}
	if customerID, ok := stripeData["customer_id"].(string); ok {
		sapData["CustomerID"] = customerID
	}

	sapData["PaymentGateway"] = "Stripe"
	sapData["Timestamp"] = time.Now().Format(time.RFC3339)

	return sapData
}

// transformSAPToSalesforce преобразует данные SAP в формат Salesforce
func (o *orchestrator) transformSAPToSalesforce(sapData map[string]interface{}) map[string]interface{} {
	salesforceData := make(map[string]interface{})
	
	if orderNum, ok := sapData["OrderNumber"].(string); ok {
		salesforceData["OrderId"] = orderNum
	}
	if status, ok := sapData["PaymentStatus"].(string); ok {
		// Маппинг статусов
		if status == "succeeded" {
			salesforceData["Status"] = "Paid"
		} else {
			salesforceData["Status"] = status
		}
	}
	if amount, ok := sapData["Amount"].(float64); ok {
		salesforceData["Amount"] = amount
	}
	if currency, ok := sapData["Currency"].(string); ok {
		salesforceData["Currency"] = currency
	}
	if customerID, ok := sapData["CustomerID"].(string); ok {
		salesforceData["AccountId"] = customerID
	}

	salesforceData["LastModifiedDate"] = time.Now().Format(time.RFC3339)

	return salesforceData
}

// findThreadForSystem находит thread для системы
func (o *orchestrator) findThreadForSystem(ctx context.Context, systemID uuid.UUID, direction models.Directions) (uuid.UUID, error) {
	// Получаем все routes для системы
	routes, err := o.routeRepo.GetBySystem(ctx, systemID)
	if err != nil {
		return uuid.Nil, err
	}

	if len(routes) == 0 {
		return uuid.Nil, fmt.Errorf("no routes found for system %s", systemID)
	}

	// Берем первый route
	targetRoute := routes[0]

	// Ищем thread route по route ID
	threadRoute, err := o.threadRouteRepo.GetThreadRouteByRouteID(ctx, targetRoute.Ref)
	if err != nil {
		return uuid.Nil, fmt.Errorf("no thread route found for route %s: %w", targetRoute.Ref, err)
	}

	if threadRoute.Direction != direction {
		return uuid.Nil, fmt.Errorf("thread route direction mismatch: expected %s, got %s", direction, threadRoute.Direction)
	}

	return threadRoute.Thread, nil
}

