-- Удаление индекса по полю name
DROP INDEX IF EXISTS idx_metrics_name;

-- Удаление таблицы метрик
DROP TABLE IF EXISTS metrics;
