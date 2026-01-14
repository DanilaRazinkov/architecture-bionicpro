-- Создание таблицы (оставляем как есть)
CREATE TABLE IF NOT EXISTS emg_sensor_data (
    user_id UInt32,
    prosthesis_type String,
    muscle_group String,
    signal_frequency UInt32,
    signal_duration UInt32,
    signal_amplitude Decimal(5,2),
    signal_time DateTime
) ENGINE = MergeTree()
ORDER BY (user_id, prosthesis_type, signal_time);

INSERT INTO emg_sensor_data
SELECT
    toUInt32(rand() % 1000) as user_id,
    ['arm', 'hand', 'leg', 'foot'][rand() % 4 + 1] as prosthesis_type,
    ['Biceps', 'Triceps', 'Hamstrings', 'Quadriceps', 'Gastrocnemius', 'Deltoids'][rand() % 6 + 1] as muscle_group,
    toUInt32(rand() % 500 + 50) as signal_frequency,
    toUInt32(rand() % 5000 + 1000) as signal_duration,
    toDecimal32(rand() % 500 / 100.0, 2) as signal_amplitude,
    now() - rand() % 2592000 as signal_time
FROM numbers(100000);
