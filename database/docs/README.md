# Cервис library

Сервис **library** реализован с использованием принципов чистой архитектуры. 

Структура данных, которые необходимо сериализовать/десериализовать для их передачи 
при помощи Protobuf для взаимодействия с сервисом с использованием gRPC, определена 
в файле [api/library/library.proto](../api/library/library.proto).

В нём же реализована поддержка путей для [gRPC gateway](https://github.com/grpc-ecosystem/grpc-gateway), а 
также валидации.

Слой web реализован в файле [internal/app/app.go](../internal/app/app.go).

Он корректно обрабатывает SIGINT и SIGTERM, а также реализует graceful shutdown.

Для валидации используется [protoc-gen-validate](https://github.com/bufbuild/protoc-gen-validate). В качестве
драйвера для Postgres используется [pgx](https://github.com/jackc/pgx). 

Слой контроллеров реализован в директории [internal/controller](../internal/controller).

Реализация слоя бизнес-логики содержится в директории
[internal/usecase/library](../internal/usecase/library).

Реализация репозитория сервиса, как и интерфейсы к нему, содержатся в 
директории [internal/usecase/repository](../internal/usecase/repository).

Конфиг, содержащийся в файле [config/config.go](../config/config.go), 
работает со следующими переменными окружения:

* GRPC_PORT - порт для gRPC сервера
* GRPC_GATEWAY_PORT - порт для gRPC gateway (REST -> gRPC API)
* POSTGRES_HOST, POSTGRES_PORT, 
POSTGRES_DB, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_MAX_CONN - параметры для подключения к Postgres

В директории [db/migrations](../db/migrations) реализованы миграции с использованием
[goose](https://github.com/pressly/goose), а в файле [db/migrations/migrate.go](../db/migrations/migrate.go]) - 
логика их применения. 

В рамках проекта реализованы интеграционные тесты в директории [integration-test](../integration-test), а также
unit-тесты в тех пакетах, функции которых они тестируют. Моки для unit-тестов генерируются ари помощи
[mockgen](https://github.com/uber-go/mock).

Для генерации кода используются [Makefile](../Makefile) и [easyp.yaml](../easyp.yaml). Также они генерируют 
OpenAPI спецификацию.

Для логирования используется [zap](https://github.com/uber-go/zap) и [logrus](https://github.com/sirupsen/logrus).