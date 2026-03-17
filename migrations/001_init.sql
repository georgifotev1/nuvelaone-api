-- +goose Up
CREATE TABLE addresses (
    id CHAR(27) PRIMARY KEY,
    street TEXT NOT NULL,
    house_number VARCHAR(10),
    city VARCHAR(50) NOT NULL,
    postal_code VARCHAR(10),
    country VARCHAR(50) NOT NULL
);  

CREATE TABLE tenants (
    id CHAR(27) PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    slug VARCHAR(50) NOT NULL UNIQUE,
    phone VARCHAR(20),
    email VARCHAR(255),
    address_id CHAR(27) REFERENCES addresses(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE working_hours (
    id CHAR(27) PRIMARY KEY,
    tenant_id CHAR(27) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,  
    day_of_week INTEGER NOT NULL CHECK(day_of_week BETWEEN 0 AND 6),
    opens_at TEXT NOT NULL,
    closes_at TEXT NOT NULL,
    is_closed BOOLEAN DEFAULT FALSE, 
    UNIQUE(tenant_id, day_of_week) 
);

CREATE INDEX idx_working_hours_org ON working_hours(tenant_id);

CREATE TABLE users (
    id CHAR(27) PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password TEXT NOT NULL,
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    tenant_id CHAR(27) REFERENCES tenants(id) ON DELETE CASCADE,
    avatar TEXT,
    role VARCHAR(20) NOT NULL CHECK(role IN ('admin', 'owner', 'member')),
    verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_tenant_id ON users(tenant_id);

CREATE TABLE customers (
    id CHAR(27) PRIMARY KEY,
    tenant_id CHAR(27) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE,
    password TEXT,
    phone VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_customers_tenant_id ON customers(tenant_id);

CREATE TABLE user_invitations (
    id CHAR(27) PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    token TEXT NOT NULL UNIQUE,
    role VARCHAR(20) NOT NULL,
    invited_by CHAR(27) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    tenant_id CHAR(27) REFERENCES tenants(id) ON DELETE CASCADE,
    accepted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invitations_token ON user_invitations(token);
CREATE INDEX idx_invitations_email ON user_invitations(email);

CREATE TABLE services (
    id CHAR(27) PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    duration INTEGER NOT NULL,
    buffer INTEGER,
    cost INTEGER NOT NULL,
    visible BOOLEAN DEFAULT FALSE,
    tenant_id CHAR(27) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_services_tenant_id ON services(tenant_id);

CREATE TABLE user_services (
    id CHAR(27) PRIMARY KEY,
    user_id CHAR(27) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    service_id CHAR(27) NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    tenant_id CHAR(27) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_services_tenant_id ON user_services(tenant_id);

CREATE TABLE events (
    id CHAR(27) PRIMARY KEY,
    customer_id CHAR(27) NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    service_id CHAR(27) NOT NULL REFERENCES services(id) ON DELETE RESTRICT,
    user_id CHAR(27) REFERENCES users(id) ON DELETE SET NULL,
    tenant_id CHAR(27) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'confirmed', 'cancelled', 'completed')),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (end_time > start_time)
);

CREATE INDEX idx_events_customer ON events(customer_id);
CREATE INDEX idx_events_service ON events(service_id);
CREATE INDEX idx_events_tenant_id ON events(tenant_id);
CREATE INDEX idx_events_user_id ON events(user_id);
CREATE INDEX idx_events_start_time ON events(start_time);

-- +goose Down
DROP INDEX IF EXISTS idx_events_start_time;
DROP INDEX IF EXISTS idx_events_user_id;
DROP INDEX IF EXISTS idx_events_tenant_id;
DROP INDEX IF EXISTS idx_events_service;
DROP INDEX IF EXISTS idx_events_customer;
DROP TABLE events;
DROP INDEX IF EXISTS idx_user_services_tenant_id;
DROP TABLE user_services;
DROP INDEX IF EXISTS idx_services_tenant_id;
DROP TABLE services;
DROP INDEX IF EXISTS idx_invitations_email;
DROP INDEX IF EXISTS idx_invitations_token;
DROP TABLE user_invitations;
DROP INDEX IF EXISTS idx_customers_tenant_id;
DROP TABLE customers;
DROP INDEX IF EXISTS idx_users_tenant_id;
DROP TABLE users;
DROP INDEX idx_working_hours_org;
DROP TABLE working_hours;
DROP TABLE tenants;
DROP TABLE addresses;
