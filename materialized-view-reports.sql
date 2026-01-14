CREATE DATABASE IF NOT EXISTS bionicpro;

USE bionicpro;

-- Создание таблицы отчетов в реальном времени
CREATE TABLE IF NOT EXISTS bionicpro.user_reports_realtime (
    user_id UInt32,
    report_date Date,
    total_signals UInt32,
    avg_duration Float64,
    muscle_groups String,
    session_count UInt32,
    avg_accuracy Float64,
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (user_id, report_date);
