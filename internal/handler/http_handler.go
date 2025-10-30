package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go-esb/internal/middleware"
	"go-esb/internal/models"
	"go-esb/internal/service"

	"github.com/gorilla/mux"
)

// HTTPHandler обрабатывает HTTP запросы
type HTTPHandler struct {
	messageService service.MessageService
	orchestrator   service.Orchestrator
}

// NewHTTPHandler создает новый HTTP обработчик
func NewHTTPHandler(messageService service.MessageService, orchestrator service.Orchestrator) *HTTPHandler {
	return &HTTPHandler{
		messageService: messageService,
		orchestrator:   orchestrator,
	}
}

// SetupRoutes настраивает маршруты HTTP API
func (h *HTTPHandler) SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Применяем middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recovery)

	// Health check
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")

	// API endpoints
	api := router.PathPrefix("/api/v1").Subrouter()

	// Обработка сообщений через thread
	api.HandleFunc("/messages/process/{threadId}", h.ProcessMessage).Methods("POST")

	// Оркестрация бизнес-процессов
	api.HandleFunc("/orchestrate/{processName}", h.OrchestrateProcess).Methods("POST")

	// Webhook для Stripe (специальный endpoint)
	api.HandleFunc("/webhooks/stripe", h.StripeWebhook).Methods("POST")

	return router
}

// HealthCheck проверка работоспособности
func (h *HTTPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"service":   "Go ESB",
	})
}

// ProcessMessage обрабатывает сообщение через thread
func (h *HTTPHandler) ProcessMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadID := vars["threadId"]
	
	direction := r.URL.Query().Get("direction")
	if direction == "" {
		direction = "In"
	}

	var messageData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&messageData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(messageData)
	if err != nil {
		http.Error(w, "Failed to marshal data", http.StatusInternalServerError)
		return
	}

	if err := h.messageService.ProcessMessage(r.Context(), threadID, models.Directions(direction), data); err != nil {
		log.Printf("❌ Error processing message: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Message processed successfully",
	})
}

// OrchestrateProcess запускает бизнес-процесс
func (h *HTTPHandler) OrchestrateProcess(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	processName := vars["processName"]

	var processData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&processData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(processData)
	if err != nil {
		http.Error(w, "Failed to marshal data", http.StatusInternalServerError)
		return
	}

	if err := h.orchestrator.ExecuteProcess(r.Context(), processName, data); err != nil {
		log.Printf("❌ Error executing process: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Process executed successfully",
		"process": processName,
	})
}

// StripeWebhook обрабатывает webhook от Stripe
func (h *HTTPHandler) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	log.Println("📥 Received Stripe webhook")

	var stripeEvent map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&stripeEvent); err != nil {
		log.Printf("❌ Failed to parse Stripe webhook: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Извлекаем тип события
	eventType, ok := stripeEvent["type"].(string)
	if !ok {
		http.Error(w, "Missing event type", http.StatusBadRequest)
		return
	}

	// Обрабатываем только payment events
	if eventType != "payment_intent.succeeded" && eventType != "charge.succeeded" {
		log.Printf("ℹ️ Skipping event type: %s", eventType)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Извлекаем данные события
	var paymentData map[string]interface{}
	if data, ok := stripeEvent["data"].(map[string]interface{}); ok {
		if obj, ok := data["object"].(map[string]interface{}); ok {
			paymentData = obj
		}
	}

	// Добавляем метаданные
	paymentData["event_type"] = eventType
	paymentData["timestamp"] = time.Now().Unix()

	// Запускаем оркестрацию процесса обработки заказа
	processData, err := json.Marshal(paymentData)
	if err != nil {
		http.Error(w, "Failed to prepare process data", http.StatusInternalServerError)
		return
	}

	if err := h.orchestrator.ExecuteProcess(r.Context(), "order_payment_flow", processData); err != nil {
		log.Printf("❌ Error in order payment flow: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("✅ Stripe webhook processed successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"event":  eventType,
	})
}

