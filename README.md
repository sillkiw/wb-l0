# wb-l0

Демонстрационный сервис с простейшим интерфейсом, отображающий данные о заказе. Эмулятор заказов → Kafka → консюмер → REST API


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
```

Основной поток: эмулятор публикует валидные/ошибочные заказы; консюмер читает, валидирует и идемпотентно пишет в БД (UPSERT). Непригодные сообщения при включенном DLQ отправляются в отдельный топик. Веб‑сервис отдаёт заказ из кэша или БД, умеет быстрый `/healthz`. Фронтенд - простая страница поиска


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

Значения по умолчанию уже зашиты в compose:

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
### 2) Поднимите весь стек 
```bash
make up-all
```
### 3) Откройте:

* Веб‑страница: [http://127.0.0.1:4000](http://127.0.0.1:4000)
---
## Команды Makefile (шпаргалка)

> Профиль **dev** (Kafka-UI, pgAdmin) включается флагом `DEV=1`. `DEV=1` нужен только если хотите Kafka-UI и pgAdmin.

### Базовые операции

```bash
make up            # поднять текущий стек (без миграций)
make down          # остановить все контейнеры
make down-v        # остановить и УДАЛИТЬ volumes (данные БД/Кафки пропадут!)
make ps            # список сервисов и их порты
make logs S=web    # логи конкретного сервиса (например, web/consumer/emulator)
make restart S=consumer  # перезапуск одного сервиса
make config        # показать итоговую docker-compose конфигурацию
```

### Полный запуск стека

```bash
make up-all            # infra -> wait-db -> миграции -> apps
make up-all DEV=1      # то же + dev-сервисы (Kafka-UI, pgAdmin)
```

После `make up-all DEV=1` доступны:

* Web UI: [http://127.0.0.1:4000](http://127.0.0.1:4000)
* Kafka UI: [http://127.0.0.1:8080](http://127.0.0.1:8080)
* pgAdmin: [http://127.0.0.1:15050](http://127.0.0.1:15050)

### Масштабирование

```bash
make up consumer=3 emulator=2        # поднять с масштабированием
make apps consumer=3 emulator=2      # перезапустить только приложения с масштабом
```

> Параметры `consumer` и `emulator` управляют числом реплик.

### Миграции (profile: tooling)

```bash
make mig-up                    # применить все миграции
make mig-down                  # откатить 1 миграцию
make mig-down-all              # откатить все миграции
make mig-version               # текущая версия
make mig-new name=create_orders_table  # создать новый файл миграции
```

### Полезные примеры

```bash
make up DEV=1                       # поднять всё (без миграций) + dev-утилиты
make up-all DEV=1 consumer=2        # полный запуск + 2 консьюмера
make logs S=consumer                # логи консьюмера 
make restart S=web                  # перезапустить только web
make down-v                         # полностью снести окружение и данные
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

### Здоровье:

* `GET /healthz` → `200 OK`

---
