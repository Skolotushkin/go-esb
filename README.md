<<<<<<< HEAD
# 🚀 Go ESB - Enterprise Service Bus

**Go ESB** — это корпоративная сервисная шина (ESB) на языке Go, предназначенная для интеграции разнородных систем с разными протоколами и форматами данных.

## 📋 Описание

Go ESB решает проблему интеграции систем, которые используют разные протоколы (REST, SOAP, AMQP) и форматы данных (JSON, XML, CSV). Система обеспечивает:

- ✅ **Маршрутизацию сообщений** между системами
- ✅ **Преобразование форматов** (JSON ↔ XML ↔ CSV)
- ✅ **Трансформацию протоколов** (REST → SOAP, REST → AMQP)
- ✅ **Безопасность и аутентификацию** (централизованное хранение токенов)
- ✅ **Оркестрацию процессов** (бизнес-потоки без переписывания кода)

## 🏗️ Архитектура

```
┌─────────────┐
│   Stripe    │ (REST, JSON)
│  (Webhook)  │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────────┐
│           Go ESB Core                    │
│  ┌─────────────┐  ┌──────────────────┐   │
│  │ Orchestrator│  │ Message Router   │   │
│  └─────────────┘  └──────────────────┘   │
│  ┌─────────────┐  ┌──────────────────┐   │
│  │ Converters  │  │ Protocol Adapters│   │
│  └─────────────┘  └──────────────────┘   │
└─────┬───────────┬───────────────┬────────┘
      │           │               │
      ▼           ▼               ▼
┌─────────┐  ┌───────────┐  ┌─────────────┐
│   SAP   │  │Salesforce │  │  RabbitMQ   │
│ (SOAP,  │  │(REST,JSON)│  │   (AMQP)    │
│   XML)  │  └───────────┘  └─────────────┘
└─────────┘
```

## 🔧 Установка

### Требования

- Go 1.21+
- PostgreSQL 12+
- (Опционально) RabbitMQ для AMQP адаптера

### Шаги установки

1. Клонируйте репозиторий:
```bash
git clone <repository-url>
cd esb
```

2. Установите зависимости:
```bash
go mod download
```

3. Настройте базу данных:
```bash
# Создайте базу данных PostgreSQL
createdb esb

# Настройте переменные окружения (или используйте config.env)
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=esb
export DB_PASSWORD=esb
export DB_NAME=esb
export DB_SSLMODE=disable
```

4. Запустите миграции (они выполняются автоматически при запуске):
```bash
go run cmd/esb-server/main.go
```

## 🚀 Использование

### Запуск сервера

```bash
export HTTP_PORT=8080  # По умолчанию 8080
go run cmd/esb-server/main.go
```

Сервер будет доступен на `http://localhost:8080`

### API Endpoints

#### Health Check
```bash
GET /health
```

#### Обработка сообщения через thread
```bash
POST /api/v1/messages/process/{threadId}?direction=In
Content-Type: application/json

{
  "order_id": "12345",
  "amount": 9999,
  "currency": "USD",
  "status": "paid"
}
```

#### Оркестрация бизнес-процесса
```bash
POST /api/v1/orchestrate/order_payment_flow
Content-Type: application/json

{
  "order_id": "12345",
  "amount": 9999,
  "currency": "USD",
  "status": "succeeded",
  "customer_id": "cus_abc123"
}
```

#### Webhook для Stripe
```bash
POST /api/v1/webhooks/stripe
Content-Type: application/json

{
  "type": "payment_intent.succeeded",
  "data": {
    "object": {
      "id": "pi_123",
      "amount": 9999,
      "currency": "usd",
      "customer": "cus_abc123",
      "status": "succeeded"
    }
  }
}
```

## 📝 Пример использования: Stripe → SAP → Salesforce

### Сценарий

При получении платежа от Stripe:
1. ESB принимает webhook от Stripe (REST, JSON)
2. Преобразует данные в XML и отправляет в SAP через SOAP
3. После подтверждения SAP отправляет обновление в Salesforce (REST, JSON)
4. Весь процесс выполняется за < 5 секунд

### Настройка

#### 1. Создание систем в БД

```sql
-- Создаем системы
INSERT INTO systems (name) VALUES ('Stripe'), ('SAP'), ('Salesforce');
```

#### 2. Создание routes

```sql
-- Route для SAP
INSERT INTO routes (name, path, system, method) 
SELECT 'SAP Order Update', '/sap/api/orders', ref, 'Post'::rest_method
FROM systems WHERE name = 'SAP';

-- Route для Salesforce
INSERT INTO routes (name, path, system, method)
SELECT 'Salesforce Order Sync', '/services/data/v57.0/sobjects/Order', ref, 'Post'::rest_method
FROM systems WHERE name = 'Salesforce';
```

#### 3. Создание connection settings

```sql
-- Настройки подключения для SAP
INSERT INTO connection_settings (name, system, path, port)
SELECT 'SAP SOAP Endpoint', ref, 'https://sap.example.com/sap/bc/soap', 443
FROM systems WHERE name = 'SAP';

-- Настройки подключения для Salesforce
INSERT INTO connection_settings (name, system, path, port)
SELECT 'Salesforce API', ref, 'https://yourinstance.salesforce.com', 443
FROM systems WHERE name = 'Salesforce';
```

#### 4. Создание threads и thread routes

```sql
-- Thread group для SAP (SOAP)
INSERT INTO threads_groups (name, protocol) 
VALUES ('SAP Integration Group', 'SOAP'::protocol_type);

-- Thread для SAP
INSERT INTO threads (name, "group", message_convert_type)
SELECT 'SAP Order Thread', ref, 'None'::message_convert_type
FROM threads_groups WHERE name = 'SAP Integration Group';

-- Thread group для Salesforce (REST)
INSERT INTO threads_groups (name, protocol)
VALUES ('Salesforce Integration Group', 'REST'::protocol_type);

-- Thread для Salesforce
INSERT INTO threads (name, "group", message_convert_type)
SELECT 'Salesforce Sync Thread', ref, 'None'::message_convert_type
FROM threads_groups WHERE name = 'Salesforce Integration Group';

-- Thread routes
-- SAP (Outbound)
INSERT INTO thread_routes (thread, direction, route, file_format)
SELECT 
    t.ref, 
    'Out'::direction,
    r.ref,
    'XML'::file_format
FROM threads t
CROSS JOIN routes r
CROSS JOIN systems s
WHERE t.name = 'SAP Order Thread'
AND r.name = 'SAP Order Update'
AND s.name = 'SAP'
AND r.system = s.ref;

-- Salesforce (Outbound)
INSERT INTO thread_routes (thread, direction, route, file_format)
SELECT
    t.ref,
    'Out'::direction,
    r.ref,
    'JSON'::file_format
FROM threads t
CROSS JOIN routes r
CROSS JOIN systems s
WHERE t.name = 'Salesforce Sync Thread'
AND r.name = 'Salesforce Order Sync'
AND s.name = 'Salesforce'
AND r.system = s.ref;
```

### Тестирование

Отправьте webhook от Stripe:

```bash
curl -X POST http://localhost:8080/api/v1/webhooks/stripe \
  -H "Content-Type: application/json" \
  -d '{
    "type": "payment_intent.succeeded",
    "data": {
      "object": {
        "id": "pi_1234567890",
        "amount": 9999,
        "currency": "usd",
        "customer": "cus_abc123",
        "status": "succeeded",
        "order_id": "ORD-12345"
      }
    }
  }'
```

ESB автоматически:
1. Примет данные от Stripe
2. Преобразует в формат SAP (XML)
3. Отправит в SAP через SOAP
4. После подтверждения отправит в Salesforce через REST

## 🔐 Безопасность

### Аутентификация

ESB поддерживает:
- **Basic Auth** для REST и SOAP
- **Bearer Token** для REST API

Настройка:

```sql
-- Создание аутентификации
INSERT INTO connection_authentications (name, system, type, token)
SELECT 'Salesforce OAuth', ref, 'BearerToken'::authentication_type, 'your-access-token'
FROM systems WHERE name = 'Salesforce';

-- Привязка к connection settings
UPDATE connection_settings 
SET auth = (SELECT ref FROM connection_authentications WHERE name = 'Salesforce OAuth')
WHERE name = 'Salesforce API';
```

## 📊 Мониторинг

Логирование выполняется через стандартный `log` пакет Go. Все операции маршрутизации, трансформации и ошибки логируются.

Примеры логов:
```
📥 Received payment data from Stripe
📤 Sending to SAP via thread: 550e8400-e29b-41d4-a716-446655440000
✅ SAP confirmed order update
📤 Sending to Salesforce via thread: 660e8400-e29b-41d4-a716-446655440001
🎉 Order Payment Flow completed in 2.3s
```

## 🧩 Компоненты

### Конвертеры форматов
- `internal/converter/converter.go` - основной конвертер
- `internal/converter/json_xml.go` - JSON ↔ XML

### Адаптеры протоколов
- `internal/adapter/rest.go` - REST адаптер
- `internal/adapter/soap.go` - SOAP адаптер
- `internal/adapter/amqp.go` - AMQP адаптер (RabbitMQ)

### Сервисы
- `internal/service/message_service.go` - маршрутизация сообщений
- `internal/service/orchestrator.go` - оркестрация процессов

## 🛠️ Разработка

### Структура проекта

```
esb/
├── cmd/
│   └── esb-server/
│       └── main.go          # Точка входа
├── internal/
│   ├── adapter/             # Протокольные адаптеры
│   ├── config/              # Конфигурация
│   ├── converter/            # Конвертеры форматов
│   ├── database/            # БД подключение
│   ├── handler/             # HTTP handlers
│   ├── model/               # Модели данных
│   ├── repository/          # Репозитории
│   └── service/             # Бизнес-логика
├── migrations/              # SQL миграции
├── config.env               # Конфигурация
└── go.mod                   # Зависимости
```

### Добавление нового адаптера

1. Создать файл `internal/adapter/new_protocol.go`
2. Реализовать интерфейс `ProtocolAdapter`
3. Добавить в `AdapterFactory`

### Добавление нового конвертера

Добавить метод в `internal/converter/converter.go` для поддержки нового формата.

## 📚 Документация

Дополнительная документация находится в комментариях к коду. Основные интерфейсы:

- `ProtocolAdapter` - интерфейс для адаптеров протоколов
- `FormatConverter` - интерфейс для конвертеров форматов
- `MessageService` - интерфейс для маршрутизации
- `Orchestrator` - интерфейс для оркестрации

---

**Go ESB** — современное решение для интеграции корпоративных систем 🚀

=======