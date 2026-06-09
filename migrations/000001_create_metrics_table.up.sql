-- Создание таблицы для метрик
CREATE TABLE metrics (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    m_type VARCHAR(50) NOT NULL,
    delta BIGINT,
    value DOUBLE PRECISION
);

-- Создание индекса по полю name для ускорения поиска
CREATE INDEX idx_metrics_name ON metrics (name);
