-- ============================================
-- Пример настройки интеграции Stripe → SAP → Salesforce
-- ============================================

-- 1. Создание систем
INSERT INTO systems (name) VALUES 
    ('Stripe'),
    ('SAP'),
    ('Salesforce')
ON CONFLICT DO NOTHING;

-- 2. Создание routes для SAP
INSERT INTO routes (name, path, system, method)
SELECT 
    'SAP Order Update',
    '/sap/bc/soap/order_update',
    s.ref,
    'Post'::rest_method
FROM systems s
WHERE s.name = 'SAP'
ON CONFLICT DO NOTHING;

-- 3. Создание routes для Salesforce
INSERT INTO routes (name, path, system, method)
SELECT
    'Salesforce Order Sync',
    '/services/data/v57.0/sobjects/Order__c',
    s.ref,
    'Post'::rest_method
FROM systems s
WHERE s.name = 'Salesforce'
ON CONFLICT DO NOTHING;

-- 4. Создание connection settings для SAP
INSERT INTO connection_settings (name, system, path, port)
SELECT
    'SAP SOAP Endpoint',
    s.ref,
    'https://sap.example.com',
    443
FROM systems s
WHERE s.name = 'SAP'
ON CONFLICT DO NOTHING;

-- 5. Создание connection settings для Salesforce
INSERT INTO connection_settings (name, system, path, port)
SELECT
    'Salesforce API Endpoint',
    s.ref,
    'https://yourinstance.salesforce.com',
    443
FROM systems s
WHERE s.name = 'Salesforce'
ON CONFLICT DO NOTHING;

-- 6. Создание аутентификации для SAP (если требуется)
INSERT INTO connection_authentications (name, system, type, username, password)
SELECT
    'SAP Basic Auth',
    s.ref,
    'Basic'::authentication_type,
    'sap_user',
    'sap_password'
FROM systems s
WHERE s.name = 'SAP'
ON CONFLICT DO NOTHING;

-- 7. Создание аутентификации для Salesforce
INSERT INTO connection_authentications (name, system, type, token)
SELECT
    'Salesforce Bearer Token',
    s.ref,
    'BearerToken'::authentication_type,
    'your-salesforce-access-token'
FROM systems s
WHERE s.name = 'Salesforce'
ON CONFLICT DO NOTHING;

-- 8. Привязка аутентификации к connection settings
UPDATE connection_settings cs
SET auth = ca.ref
FROM connection_authentications ca
WHERE cs.system = ca.system
AND cs.name = 'SAP SOAP Endpoint'
AND ca.name = 'SAP Basic Auth';

UPDATE connection_settings cs
SET auth = ca.ref
FROM connection_authentications ca
WHERE cs.system = ca.system
AND cs.name = 'Salesforce API Endpoint'
AND ca.name = 'Salesforce Bearer Token';

-- 9. Создание thread group для SAP (SOAP протокол)
INSERT INTO threads_groups (name, protocol)
VALUES ('SAP Integration Group', 'SOAP'::protocol_type)
ON CONFLICT DO NOTHING;

-- 10. Создание thread для SAP
INSERT INTO threads (name, "group", message_convert_type)
SELECT
    'SAP Order Processing Thread',
    tg.ref,
    'None'::message_convert_type
FROM threads_groups tg
WHERE tg.name = 'SAP Integration Group'
ON CONFLICT DO NOTHING;

-- 11. Создание thread group для Salesforce (REST протокол)
INSERT INTO threads_groups (name, protocol)
VALUES ('Salesforce Integration Group', 'REST'::protocol_type)
ON CONFLICT DO NOTHING;

-- 12. Создание thread для Salesforce
INSERT INTO threads (name, "group", message_convert_type)
SELECT
    'Salesforce Order Sync Thread',
    tg.ref,
    'None'::message_convert_type
FROM threads_groups tg
WHERE tg.name = 'Salesforce Integration Group'
ON CONFLICT DO NOTHING;

-- 13. Создание thread route для SAP (Outbound, JSON -> XML)
INSERT INTO thread_routes (thread, direction, route, file_format)
SELECT
    t.ref,
    'Out'::direction,
    r.ref,
    'XML'::file_format
FROM threads t
CROSS JOIN routes r
CROSS JOIN systems s
WHERE t.name = 'SAP Order Processing Thread'
AND r.name = 'SAP Order Update'
AND s.name = 'SAP'
AND r.system = s.ref
ON CONFLICT (thread, direction, route) DO NOTHING;

-- 14. Создание thread route для Salesforce (Outbound, JSON)
INSERT INTO thread_routes (thread, direction, route, file_format)
SELECT
    t.ref,
    'Out'::direction,
    r.ref,
    'JSON'::file_format
FROM threads t
CROSS JOIN routes r
CROSS JOIN systems s
WHERE t.name = 'Salesforce Order Sync Thread'
AND r.name = 'Salesforce Order Sync'
AND s.name = 'Salesforce'
AND r.system = s.ref
ON CONFLICT (thread, direction, route) DO NOTHING;

-- Проверка созданных записей
SELECT 'Systems' as type, COUNT(*) as count FROM systems
UNION ALL
SELECT 'Routes', COUNT(*) FROM routes
UNION ALL
SELECT 'Thread Groups', COUNT(*) FROM threads_groups
UNION ALL
SELECT 'Threads', COUNT(*) FROM threads
UNION ALL
SELECT 'Thread Routes', COUNT(*) FROM thread_routes;

