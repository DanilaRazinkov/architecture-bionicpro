-- Создание таблицы (если не существует)
CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    email VARCHAR(100),
    age INTEGER,
    gender VARCHAR(10),
    country VARCHAR(100),
    address VARCHAR(255),
    phone VARCHAR(25)
);

INSERT INTO customers (name, email, age, gender, country, address, phone)
SELECT
    'Customer_' || i,
    'customer.' || i || '@example.com',
    floor(random() * 60 + 18)::int,
    CASE
        WHEN random() < 0.4 THEN 'Male'
        WHEN random() < 0.8 THEN 'Female'
        ELSE 'Non-Binary'
    END,
    (ARRAY[
        'United States', 'Canada', 'United Kingdom', 'Germany', 'France',
        'Australia', 'Japan', 'Brazil', 'Mexico', 'Italy',
        'Falkland Islands (Malvinas)', 'Georgia', 'Russia', 'China', 'India'
    ])[floor(random() * 15 + 1)],
    'Address ' || (i * 100) || ', City ' || (i % 20 + 1) || ', State ' || chr(65 + (i % 26)),
    '(' || floor(random() * 900 + 100)::text || ') ' ||
    floor(random() * 900 + 100)::text || '-' ||
    lpad(floor(random() * 10000)::text, 4, '0')
FROM generate_series(1, 1000) AS i;
