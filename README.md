# DC Payment Gateway

Интеграционный сервис узла **DC** — принимает платежи от платёжной системы
ExpressPay (владелец — DC City Bank) по протоколу `check`/`pay` и зачисляет
их на счета клиентов в своей базе данных.

## Архитектура

Gateway (ExpressPay) → DC (этот сервис) → DB (Postgres)

Собран по Clean Architecture — зависимости текут внутрь, от транспорта к домену:

- `internal/domain` — модели и интерфейсы (`ProcessingCenter`, `AccountRepository`, `PaymentRepository`), не зависят ни от HTTP, ни от БД
- `internal/expresspay` — HTTP-слой: парсинг query-параметров, проверка подписи (`protocol.go`), сами хендлеры `check`/`pay` (`handler.go`)
- `internal/service` — бизнес-логика: идемпотентность по `txn_id`, вызов ProcessingCenter
- `internal/pc` — реализация `ProcessingCenter` (сейчас — читает/пишет балансы прямо в Postgres; в будущем заменится на реального процессингового центра банка, спецификации на который пока нет)
- `internal/repository` — работа с БД (`accounts`, `payments`)
- `internal/transport/http` — роутинг (`router.go`)
- `internal/migrations` — встроенные (`go:embed`) SQL-миграции, применяются автоматически при старте
- `internal/config` — чтение конфигурации из `.env`/переменных окружения
- `pkg/logger` — структурированное логирование (zap)

На текущем этапе `DC` не ходит ни в какой внешний процессинговый центр —
балансы считаются и хранятся локально в собственной таблице `accounts`.
Причина: спецификация внутреннего процессингового центра банка ("PC" на
схеме архитектуры) ещё не предоставлена — `internal/pc` спроектирован через
интерфейс `ProcessingCenter` именно для того, чтобы при появлении такой
спецификации заменить одну реализацию, не трогая остальной код.

## Протокол

Реализованы команды `check` и `pay` из протокола ExpressPay API (v1.0.4).

Запросы приходят методом **GET** на `/test.asp`, ответы — в формате XML.
Подпись считается по-разному для каждой команды:

| Команда | Формула подписи                            |
| ------- | ------------------------------------------ |
| `check` | `md5(login + password)`                    |
| `pay`   | `md5(login + txn_id + account + password)` |

Запрос без подписи или с неверной подписью отклоняется с `result=13`
(`Error sign`) — защита от обхода проверки выключенным/отсутствующим `sign`
закрыта явно.

### Идемпотентность

Повторный `pay` с уже обработанным `txn_id` **не** зачисляет деньги повторно —
возвращается результат того платежа, что был проведён первым. Это защищает
от дублей при сетевых повторах запросов со стороны ExpressPay. Гонка между
параллельными запросами с одним `txn_id` закрыта на уровне БД через
`ON CONFLICT (txn_id) DO NOTHING` + проверку `RowsAffected`.

### Коды результата

| Код  | Значение                                           |
| ---- | -------------------------------------------------- |
| `0`  | OK                                                 |
| `4`  | неверный формат (например, сумма не число или ≤ 0) |
| `5`  | счёт не найден или неактивен                       |
| `13` | ошибка подписи (`sign`)                            |
| `1`  | внутренняя ошибка                                  |

## Запуск

### Через Docker (рекомендуется)

```bash
cp .env.example .env
docker compose up --build
```

Поднимутся два контейнера — `postgres` и `app`. Миграции применятся
автоматически при первом старте. В логах `app` должна появиться цепочка:

database connection established
migrations applied
applied migration: 0001_init.sql
HTTP server running {"port": ":8080"}

Если порт `5432` уже занят локальным Postgres на хосте — поменяй проброс
порта у сервиса `postgres` в `docker-compose.yml` на `"5433:5432"` (на
общение `app` ↔ `postgres` внутри docker-сети это не влияет).

### Локально

```bash
cp .env.example .env
go mod tidy
go run ./cmd/server
```

Требуется локально запущенный Postgres, доступный по данным из `.env`.
Миграции применяются автоматически при каждом старте — повторный запуск не
накатывает уже применённые миграции заново (отслеживается в таблице
`schema_migrations`).

## Переменные окружения

| Переменная                                                              | Описание                              | Обязательна                         |
| ----------------------------------------------------------------------- | ------------------------------------- | ----------------------------------- |
| `ENV`                                                                   | `development` / `production`          | нет                                 |
| `SERVER_PORT`                                                           | порт HTTP-сервера                     | нет                                 |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE` | подключение к Postgres                | нет (есть значения по умолчанию)    |
| `EXPRESSPAY_LOGIN`, `EXPRESSPAY_PASSWORD`                               | креды для проверки подписи ExpressPay | **да** — сервис не стартует без них |

См. `.env.example` для полного шаблона.

## Миграции

Лежат в `internal/migrations/*.sql`, встроены в бинарник через `go:embed` —
отдельно копировать их при деплое не нужно, они физически часть собранного
исполняемого файла. Применяются по порядку имени (`0001_`, `0002_`, ...),
каждая — ровно один раз; прогресс трекается таблицей `schema_migrations`.

Текущая миграция `0001_init.sql` создаёт:

```sql
CREATE TABLE IF NOT EXISTS accounts (
    account    VARCHAR(50) PRIMARY KEY,
    full_name  VARCHAR(255) NOT NULL,
    balance    NUMERIC(18,2) NOT NULL DEFAULT 0,
    is_active  BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS payments (
    txn_id      VARCHAR(50) PRIMARY KEY,
    account     VARCHAR(50) NOT NULL REFERENCES accounts(account),
    amount      NUMERIC(18,2) NOT NULL,
    prv_txn     VARCHAR(50),
    status      VARCHAR(20) NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT now()
);

INSERT INTO accounts (account, full_name, balance, is_active) VALUES
    ('992918400400', 'Usmonalizoda Suhrob Usmonali', 1000.00, true),
    ('992900000001', 'Test Client Two', 500.00, true),
    ('992900000002', 'Blocked Client', 0.00, false)
ON CONFLICT (account) DO NOTHING;
```

## Тестирование через Postman

Юнит-тестов пока нет — протокол проверялся вручную через Postman, напрямую
по HTTP, с ручным расчётом подписи через Pre-request Script.

### Environment `dc-local`

| Variable   | Value                                           |
| ---------- | ----------------------------------------------- |
| `base_url` | `http://localhost:8080`                         |
| `login`    | `21950`                                         |
| `password` | `147521`                                        |
| `account`  | `992918400400`                                  |
| `txn_id`   | `100001`                                        |
| `sum`      | `10.00`                                         |
| `ccy`      | `TJS`                                           |
| `sign`     | _(пусто, вычисляется скриптом перед отправкой)_ |

### Запрос `check`

GET {{base_url}}/test.asp?command=check&login={{login}}&account={{account}}&sum={{sum}}&sign={{sign}}

Pre-request Script:

```javascript
const login = pm.environment.get("login");
const password = pm.environment.get("password");
const sign = CryptoJS.MD5(login + password).toString();
pm.environment.set("sign", sign);
```

### Запрос `pay`

GET {{base_url}}/test.asp?command=pay&login={{login}}&txn_id={{txn_id}}&account={{account}}&sum={{sum}}&ccy={{ccy}}&sign={{sign}}

Pre-request Script (формула подписи отличается от `check`):

```javascript
const login = pm.environment.get("login");
const password = pm.environment.get("password");
const txnID = pm.environment.get("txn_id");
const account = pm.environment.get("account");
const sign = CryptoJS.MD5(login + txnID + account + password).toString();
pm.environment.set("sign", sign);
```

### Проверенные сценарии

| №   | Сценарий                                 | Как воспроизвести                           | Ожидаемый результат               |
| --- | ---------------------------------------- | ------------------------------------------- | --------------------------------- |
| 1   | Happy path: `check` на существующий счёт | `account=992918400400`, корректный `sign`   | `result=0`                        |
| 2   | Happy path: `pay` проводит платёж        | тот же `account`, `txn_id=100001`           | `result=0` + `prv_txn`            |
| 3   | Идемпотентность                          | повторный `pay` с тем же `txn_id=100001`    | тот же самый `prv_txn`, не новый  |
| 4   | Новый платёж                             | `pay` с новым `txn_id` (например `100002`)  | новый `prv_txn`, `result=0`       |
| 5   | Несуществующий счёт                      | `account=000000000000`, `check`             | `result=5`                        |
| 6   | Невалидная сумма                         | `sum=-10.00`, `pay`                         | `result=4`                        |
| 7   | Отсутствующая подпись                    | в URL `pay` убран параметр `&sign={{sign}}` | `result=13`, платёж не проводится |

Все 7 сценариев пройдены успешно как при локальном запуске (`go run`), так и
при запуске через `docker compose up --build` на чистом volume.

## Известные ограничения

- Нет интеграции с реальным процессинговым центром банка — балансы считаются
  в собственной БД сервиса, а не в core-banking системе; как только появится
  спецификация PC, заменяется одна реализация в `internal/pc`
- Не реализованы остальные команды протокола ExpressPay: `getinfo`,
  `getstatus`, `prvid`, `getbalance`, `getcurrency`, `cancel`
- Нет покрытия юнит-тестами (проверено вручную через Postman, см. раздел
  "Тестирование через Postman")
- Нет деплоя на реальный сервер — сервис запускается только локально/в Docker
