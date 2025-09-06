# wb-l0

Демонстрационный сервис с простейшим интерфейсом, отображающий данные о заказе. Эмулятор заказов → Kafka → консюмер → REST API

## Содержание

* [Архитектура](#архитектура)
* [Компоненты](#компоненты)
* [Быстрый старт](#быстрый-старт)
* [Makefile](#makefile)
* [Переменные окружения](#переменные-окружения)
* [API](#api)
* [Фронтенд](#фронтенд)
* [Хранилище и миграции](#хранилище-и-миграции)
* [Валидация](#валидация)
* [DLQ](#dlq)
* [Тесты](#тесты)
* [Траблшутинг](#траблшутинг)
* [Структура проекта](#структура-проекта)
* [Лицензия](#лицензия)

---

## Архитектура

```mermaid
%%{init: {'flowchart': {'nodeSpacing': 20, 'rankSpacing': 25, 'padding': 6}, 'themeVariables': {'fontSize': '12px'}} }%%
flowchart LR
  subgraph Prod["Производители"]
    E["Эмуляторы заказов ×N"]
  end
  subgraph K["Kafka"]
    T[(topic: orders)]
    TD[(topic: orders-dlq)]
  end
  subgraph Cgrp["Консьюмеры ×N"]
    C["Consumer (validate→save→commit)"]
  end
  subgraph Web["Веб-сервис"]
    FE["UI /static"]
    API["HTTP API /order/{id}"]
    Cache["In-proc LRU cache"]
  end
  DB[("Postgres")]
  E --> T
  T --> C
  C -->|UPSERT| DB
  C -. DLQ(reason, attempt) .-> TD
  FE --> API
  API -->|get| Cache
  Cache -- hit --> API
  API -->|miss| DB
  DB -->|result| API


Основной поток: эмулятор публикует валидные/ошибочные заказы; консюмер читает, валидирует и идемпотентно пишет в БД (UPSERT). Непригодные сообщения при включенном DLQ отправляются в отдельный топик. Веб‑сервис отдаёт заказ из кэша или БД, умеет быстрый `/healthz`. Фронтенд - простая страница поиска

---

## Компоненты

* **cmd/emulator** — генератор фейковых заказов (ULID‑идентификаторы, согласованные суммы).
* **cmd/consumer** — Kafka консюмер: десериализация, валидация, сохранение, опц. DLQ.
* **cmd/web** — HTTP API (`/order/{id}`) + статика (страница поиска).
* **internal/kafka** — тонкая обёртка над `segmentio/kafka-go`.
* **internal/storage** — Postgres (чтение/запись, UPSERT).
* **internal/consumer** — handler с бизнес‑логикой (DLQ/валидация/логгирование).
* **internal/dlq** — publisher в DLQ (Kafka).
* **internal/generator** — генерация фейковых заказов (ULID/track/chksum).
* **internal/validation** — пакет для декларативной проверки заказов.
* **db/migrations** — миграции для Postgres (`migrate`).

---

## Быстрый старт

### 1) Подготовьте `.env`

Минимум для локалки (значения по умолчанию уже зашиты в compose):

```env
COMPOSE_PROJECT_NAME=wb-l0
TZ=Europe/Amsterdam

# Kafka
KAFKA_PORT=9092
KAFKA_PORT_EXT=19092
KAFKA_BOOTSTRAP_INTERNAL=kafka:9092
KAFKA_BOOTSTRAP_EXTERNAL=localhost:19092
KAFKA_TOPIC=orders
KAFKA_DLQ_TOPIC=orders-dlq
KAFKA_GROUP_ID=orders-consumer

# Postgres
POSTGRES_USER=demo_user
POSTGRES_PASSWORD=demo
POSTGRES_DB=demo_db
POSTGRES_PORT=5432
POSTGRES_PORT_EXT=15432
DATABASE_URL=postgres://demo_user:demo@postgres:5432/demo_db?sslmode=disable
DATABASE_URL_HOST=postgres://demo_user:demo@localhost:15432/demo_db?sslmode=disable

# Apps
HTTP_PORT_EXT=4000
LOG_FORMAT=json      # json|text
LOG_LEVEL=INFO       # DEBUG|INFO|WARN|ERROR

# Эмулятор
PRODUCER_COUNT=0
PRODUCER_INTERVAL=2s
PRODUCER_BAD_RATE=0.05
PRODUCER_BAD_KINDS=malformed,validation,unknown_field,type_mismatch,sums_mismatch,future_date
```

### 2) Поднимите всё

* **Без dev‑инструментов**:

```bash
make up-all
```

* **С dev‑инструментами (Kafka UI, pgAdmin):**

```bash
make up-all DEV=1
```

Откройте:

* Веб‑страница: [http://127.0.0.1:4000](http://127.0.0.1:4000)
* Kafka UI: [http://127.0.0.1:8080](http://127.0.0.1:8080) (только с `DEV=1`)
* pgAdmin: [http://127.0.0.1:15050](http://127.0.0.1:15050) (логин/пароль — из `.env`, по умолчанию [admin@example.com](mailto:admin@example.com)/admin)

> Масштабирование: `make up consumer=3 emulator=2`

---

## Makefile


```bash
make up            # поднять текущее окружение
make down          # остановить
make down-v        # остановить и удалить volumes (осторожно!)
make ps            # состояние контейнеров
make logs S=web    # логи конкретного сервиса
make restart S=consumer

# миграции (используется профиль tooling)
make mig-up
make mig-down
make mig-down-all
make mig-version
make mig-new name=create_orders_table
```

Dev‑профиль включается через `DEV=1`:

```bash
make up DEV=1
make up-all DEV=1
```

---

## Переменные окружения

Ключевые:

* **Kafka**

  * `KAFKA_BOOTSTRAP_INTERNAL` — адрес брокера внутри докера (`kafka:9092`).
  * `KAFKA_BOOTSTRAP_EXTERNAL` — адрес брокера с хоста (`localhost:19092`).
  * `KAFKA_TOPIC` — основной топик заказов.
  * `KAFKA_DLQ_TOPIC` — топик DLQ (если пусто — DLQ выключен).
  * `KAFKA_GROUP_ID` — группа консюмера.
* **Postgres**

  * `DATABASE_URL` — для контейнеров.
  * `DATABASE_URL_HOST` — для хоста (локальная отладка).
* **Логи**

  * `LOG_FORMAT` — `json|text`.
  * `LOG_LEVEL` — `DEBUG|INFO|WARN|ERROR`.
* **Эмулятор**

  * `PRODUCER_COUNT` — сколько сообщений послать (0 = бесконечно).
  * `PRODUCER_INTERVAL` — интервал генерации.
  * `PRODUCER_BAD_RATE` и `PRODUCER_BAD_KINDS` — доля и виды «плохих» сообщений.
* **Прочее**

  * `TZ` — часовой пояс контейнеров (по умолчанию `Europe/Amsterdam`).

---

## API

### GET `/order/{id}`

Возвращает JSON заказа. Источник данных помечается заголовком:

* `X-Source: cache` — найдено в кэше процесса;
* `X-Source: db` — прочитано из БД;
* `X-Source: miss` — не найдено в БД (404).

Коды:

* `200 OK` — найден
* `404 Not Found` — нет такого `order_uid`
* `400 Bad Request` — некорректный `id`
* `504 Gateway Timeout` — таймаут БД
* `500 Internal Server Error` — иные ошибки

Здоровье:

* `GET /healthz` → `200 OK`

---
