# Setup

```shell
docker compose up -d
```

Потом необходимо подождать несколько минут, пока выполнятся миграции и поднимутся сервисы.

# Task 1

1. Диаграмма архитектуры - [link](./diagrams/tags1.drawio.png)
2. Frontend - [link](./frontend)
3. Backend - [link](./backend/auth)
4. Экспортированный realm - [link](./keycloak/keycloak-results-export.json)
5. Для добавления Yandex ID, реализован прокси, который адаптирует ответы Yandex под
   Keycloack - [link](./backend/yandex-proxy)

**Креды тестовых пользователей:**

user1 / password123

user2 / password123

# Task 2

1. Диаграмма архитектуры - [link](./diagrams/task2.drawio.png)
2. Код DAG-а - [link](./airflow/dags)
3. Backend Report - [link](./backend/report)

# Task 3

1. Обновленный сервис - [link](./backend/report)
2. Конфиг Nginx для хостинга S3 - [link](./nginx/nginx.conf)
3. Docker compose с настройкой nginx - [link](./docker-compose.yaml)

# Task 4

1. Docker compose с настройкой Kafka и kafka-connect - [link](./docker-compose.yaml)
2. Файл конфигурации debezium-connector - [link](./debezium/connectors/register-crm-connector.json)
3. Скрипты для Clickhouse
    1. Настройка - [link](./olap-db)
    2. Cоздания MaterializedView - [link](./olap-setup/materialized-view-reports.sql)
