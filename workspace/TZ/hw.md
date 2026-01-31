# 🏠 Домашнее задание — Неделя 1
**Курс: «Микросервисы, как в BigTech 2.0»**

---

## 📌 Задание

На этой неделе вам предстоит реализовать три сервиса и связать их между собой, следуя принципам контрактного программирования и практикам, используемым в крупных компаниях.

### ✅ Что нужно сделать:

1. Реализовать **HTTP API для `OrderService`**, строго следуя OpenAPI-контракту [`order_service_contracts.md`](order_service_contracts.md).
2. Реализовать **gRPC API для `InventoryService`**, используя контракт [`inventory_service_contracts.md`](inventory_service_contracts.md).
3. Реализовать **gRPC API для `PaymentService`**, используя контракт [`payment_service_contracts.md`](payment_service_contracts.md).
4. В `OrderService` через gRPC-клиенты интегрировать вызовы `InventoryService` и `PaymentService`, реализовав бизнес-логику сервиса.

---

📁 Все контракты находятся в папке [`/shared`](shared). Protobuf-контракты находятся в `shared/proto`, а OpenAPI-контракты — в `shared/api`. Сгенерированные файлы будут находиться в `shared/pkg/proto` и `shared/pkg/openapi`.

📌 Контракты описывают **цель разработки на этой неделе** — к финалу ваша реализация должна соответствовать этим контрактам. В последующих неделях сервисы будут расширяться и усложняться.

---

📂 Для упрощения выполнения домашнего задания, в папке [`/boilerplates`](boilerplates) находятся вспомогательные файлы:
- `Taskfile.yml` — файл с готовыми командами для генерации кода, форматирования, линтинга и других задач. Вы можете использовать этот файл для экономии времени, скопировав его в корень вашего проекта.
  - В файле реализована команда `test-api`, которая запускает автоматические проверки вашего API. Эта команда выполняет ряд тестовых сценариев, включая создание заказа, его оплату и отмену, взаимодействуя с вашими сервисами через HTTP и gRPC. Успешное прохождение этих тестов может служить подтверждением корректности реализации API согласно контрактам.

---

## 🛠 Рекомендации по оформлению проекта

1. **Используйте монорепозиторий с Go Workspaces**:
  - Создайте файл `go.work` в корне и подключите все модули.
  - Это упростит сборку, разработку и генерацию кода.
  - Подробнее: [Go Workspaces](https://go.dev/doc/tutorial/workspaces)

2. **Подключите линтер и придерживайтесь структуры из [boilerplate-репозитория](https://github.com/olezhek28/microservices_course_boilerplate)**:
  - Оформите `Taskfile.yml` с командами, которые были показаны в уроках.
  - Поддерживайте порядок и единообразие между сервисами.

3. **Синхронизируйте версии библиотек и инструментов**:
  - Используйте актуальную версию Go (например, Go 1.24+).
  - Обновляйте зависимости и используйте `go mod tidy`.

---

## 🛠 Актуальная структура проекта

```
.
├── README.md
├── Taskfile.yml
├── buf.work.yaml
├── go.work
├── go.work.sum
├── inventory
│   ├── cmd
│   │   └── main.go
│   ├── go.mod
│   └── go.sum
├── order
│   ├── cmd
│   │   └── main.go
│   ├── go.mod
│   └── go.sum
├── package-lock.json
├── package.json
├── payment
│   ├── cmd
│   │   └── main.go
│   ├── go.mod
│   └── go.sum
└── shared
    ├── api
    │   └── order
    │       └── v1
    │           ├── components
    │           │   ├── create_order_request.yaml
    │           │   ├── create_order_response.yaml
    │           │   ├── enums
    │           │   │   ├── order_status.yaml
    │           │   │   └── payment_method.yaml
    │           │   ├── errors
    │           │   │   ├── bad_gateway_error.yaml
    │           │   │   ├── bad_request_error.yaml
    │           │   │   ├── conflict_error.yaml
    │           │   │   ├── forbidden_error.yaml
    │           │   │   ├── generic_error.yaml
    │           │   │   ├── internal_server_error.yaml
    │           │   │   ├── not_found_error.yaml
    │           │   │   ├── rate_limit_error.yaml
    │           │   │   ├── service_unavailable_error.yaml
    │           │   │   ├── unauthorized_error.yaml
    │           │   │   └── validation_error.yaml
    │           │   ├── get_order_response.yaml
    │           │   ├── order_dto.yaml
    │           │   ├── pay_order_request.yaml
    │           │   └── pay_order_response.yaml
    │           ├── order.openapi.yaml
    │           ├── params
    │           │   └── order_uuid.yaml
    │           └── paths
    │               ├── order_by_uuid.yaml
    │               ├── order_cancel.yaml
    │               ├── order_pay.yaml
    │               └── orders.yaml
    ├── go.mod
    ├── go.sum
    ├── pkg
    │   ├── openapi
    │   │   └── order
    │   │       └── v1
    │   │           ├── oas_cfg_gen.go
    │   │           ├── oas_client_gen.go
    │   │           ├── oas_handlers_gen.go
    │   │           ├── oas_interfaces_gen.go
    │   │           ├── oas_json_gen.go
    │   │           ├── oas_labeler_gen.go
    │   │           ├── oas_middleware_gen.go
    │   │           ├── oas_operations_gen.go
    │   │           ├── oas_parameters_gen.go
    │   │           ├── oas_request_decoders_gen.go
    │   │           ├── oas_request_encoders_gen.go
    │   │           ├── oas_response_decoders_gen.go
    │   │           ├── oas_response_encoders_gen.go
    │   │           ├── oas_router_gen.go
    │   │           ├── oas_schemas_gen.go
    │   │           ├── oas_server_gen.go
    │   │           ├── oas_unimplemented_gen.go
    │   │           └── oas_validators_gen.go
    │   └── proto
    │       ├── inventory
    │       │   └── v1
    │       │       ├── inventory.pb.go
    │       │       └── inventory_grpc.pb.go
    │       └── payment
    │           └── v1
    │               ├── payment.pb.go
    │               └── payment_grpc.pb.go
    └── proto
        ├── buf.gen.yaml
        ├── buf.yaml
        ├── inventory
        │   └── v1
        │       └── inventory.proto
        └── payment
            └── v1
                └── payment.proto
```

---

### 🔧 Комментарии

- **Каждый сервис — в своём модуле с `go.mod`**, подключённом в `go.work`.
- **Логика каждого сервиса пока допускается в одном `main.go`**, позже вы будете её выносить в слои.
- **Контракты и сгенерированные клиенты** (protobuf, ogen) находятся в `shared/`.

---

## 💡 Полезные подсказки

- 📚 Прежде чем писать код, **ознакомьтесь с уроками первой недели**. Они содержат примеры и объяснения, которые помогут быстро стартовать.
- 🧠 Если вы застряли — **обращайтесь в чат или к ревьюеру**, особенно если прошло больше 30 минут и вы не продвинулись.
> **Спросить — не значит сдаться.** Это ускоряет обучение.

---

**Автор курса: Олег Козырев, 2025**
