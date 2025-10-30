CREATE EXTENSION IF NOT EXISTS "pgcrypto";
-- ===========================
-- ENUMS
-- ===========================

CREATE TYPE direction AS ENUM ('In', 'Out');
CREATE TYPE value_type AS ENUM ('String', 'Date', 'Integer', 'Boolean', 'Null', 'Structure', 'Array');
CREATE TYPE file_format AS ENUM ('XML', 'JSON', 'DBF', 'CSV', 'TXT');
CREATE TYPE routine_type AS ENUM ('Before', 'After');
CREATE TYPE message_convert_type AS ENUM ('Multiplex', 'Split', 'None');
CREATE TYPE protocol_type AS ENUM ('TCP', 'REST', 'SOAP', 'AMQP');
CREATE TYPE message_broker_type AS ENUM ('Kafka', 'Rabbit');
CREATE TYPE authentication_type AS ENUM ('Basic', 'BearerToken');
CREATE TYPE rest_method AS ENUM ('Get', 'Post', 'Patch', 'Put', 'Delete');

-- ===========================
-- USERS
-- ===========================

CREATE TABLE users (
    ref UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(50) NOT NULL
);

-- ===========================
-- GLOBAL SETTINGS
-- ===========================

CREATE TABLE global (
    name VARCHAR(100) PRIMARY KEY,
    value JSONB
);

-- ===========================
-- SYSTEMS
-- ===========================

CREATE TABLE systems (
    ref UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE
);

-- ===========================
-- ROUTES
-- ===========================

CREATE TABLE routes (
    ref UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    path VARCHAR(300) NOT NULL,
    system UUID REFERENCES systems(ref) ON DELETE CASCADE,
    method rest_method NOT NULL
);

-- ===========================
-- CONNECTIONS
-- ===========================

CREATE TABLE connection_authentications (
    ref UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    system UUID REFERENCES systems(ref) ON DELETE CASCADE,
    type authentication_type NOT NULL,
    username VARCHAR(50),
    password VARCHAR(50),
    token VARCHAR(100)
);

CREATE TABLE connection_settings (
    ref UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    system UUID REFERENCES systems(ref) ON DELETE CASCADE,
    path VARCHAR(300),
    port INTEGER,
    auth UUID REFERENCES connection_authentications(ref) ON DELETE SET NULL
);

-- ===========================
-- THREADS & GROUPS
-- ===========================

CREATE TABLE threads_groups (
    ref UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    protocol protocol_type NOT NULL,
    parent UUID REFERENCES threads_groups(ref) ON DELETE CASCADE,
    message_broker message_broker_type
);

CREATE TABLE threads (
    ref UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    "group" UUID REFERENCES threads_groups(ref) ON DELETE CASCADE,
    message_convert_type message_convert_type NOT NULL
);

-- ===========================
-- THREAD OBJECTS
-- ===========================

CREATE TABLE thread_objects (
    ref UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    name_object VARCHAR(100),
    type value_type NOT NULL,
    parent UUID REFERENCES thread_objects(ref) ON DELETE CASCADE
);

-- ===========================
-- ROUTINES
-- ===========================

CREATE TABLE routines (
    ref UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    type routine_type NOT NULL,
    code TEXT
);

-- ===========================
-- THREAD ROUTES
-- ===========================

CREATE TABLE thread_routes (
    thread UUID REFERENCES threads(ref) ON DELETE CASCADE,
    direction direction NOT NULL,
    route UUID REFERENCES routes(ref) ON DELETE CASCADE,
    file_format file_format NOT NULL,
    object UUID REFERENCES thread_objects(ref) ON DELETE SET NULL,
    routine UUID REFERENCES routines(ref) ON DELETE SET NULL,
    PRIMARY KEY (thread, direction, route)
);
