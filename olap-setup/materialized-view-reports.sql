-- 1. Сначала создайте пустую таблицу с нужной структурой
CREATE TABLE user_reports_realtime
(
    user_id UInt64,
    session_count UInt64,
    total_signals UInt64,
    avg_duration Float64,
    muscle_groups Array(String),
    avg_accuracy Float64,
    last_activity DateTime,
    updated DateTime
)
ENGINE = ReplacingMergeTree(updated)
ORDER BY user_id
PRIMARY KEY user_id;

CREATE MATERIALIZED VIEW user_reports_mv
TO user_reports_realtime
AS
SELECT
    c.id as user_id,
    count() as session_count,
    sum(s.signal_frequency) as total_signals,
    avg(s.signal_duration) as avg_duration,
    groupUniqArray(s.muscle_group) as muscle_groups,
    avg(s.signal_amplitude) * 100 as avg_accuracy,
    max(s.signal_time) as last_activity,
    now() as updated
FROM default.dim_customers c
LEFT JOIN default.emg_sensor_data s ON c.id = s.user_id
WHERE c.is_deleted = 0
GROUP BY c.id;
