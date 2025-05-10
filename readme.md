
```
sub-pub/
├── internal/
│   ├── api/                # gRPC-имплементация (handlers)
│   │   └── subpub/
│   │        ├── publish.go
│   │        ├── service.go
│   │        └── subscribe.go
│   ├── app/                # Инициализация приложения и зависимостей (DI контейнер)
│   │   ├── app.go
│   │   ├── service-provider.go
│   │   └── ...
│   ├── closer/             # Graceful shutdown (освобождение ресурсов)
│   │   └── closer.go
│   ├── config/             # Загрузка конфигов (grpc, .env)
│   └── logger/             # Инициализация zap-логгера (наверно зря вынес)
│
├── pkg/
│   └── subpub_v1/          # gRPC protobuf сгенерированные файлы
│
├── subpub/                 # Пакет с реализацией шины Pub/Sub (без gRPC)
│   ├── subpub.go
│   └── subpub_test.go
│
├── internal/cmd/
│   └── grpc_server/        # Точка входа для gRPC сервера
│        └── main.go
│
├── api/                    # Описание контракта grpc (protobuf)
│   └── subpub_v1/
│        └── subpub.proto
│
├── Dockerfile              # Dockerfile для сборки сервера
├── docker-compose.yml      # Docker Compose для быстрого запуска
├── go.mod / go.sum         # Go-модули
├── .env                    # Переменные окружения (опционально)
└── README.md               # Описание проекта
```

## Возможности
- Подписка и публикация сообщений по темам (Pub/Sub)
- gRPC API (streaming + unary)
- Логирование API через zap в ```/logger```
- DI контейнер в ```/app```
- Абстракция через интерфейсы
- Изменение параметров сервиса в конфиге (.env)
- Graceful shutdown (корректное завершение по Ctrl+C) в ```/Closer```
- Покрытие тестами основных сценариев 
- Docker-окружение для быстрого старта

## Запуск
1. Клонируйте репозиторий
 ```git clone https://github.com/yourname/sub-pub.git```

 ``` cd sub-pub```
2. Соберите и запустите через Docker-Compose
```docker-compose up --build```
3. Локальный запуск (без Docker)
```go run internal/cmd/grpc_server/main.go```

##### !Сервис работает по порту: 50051

## Тестирование

Я тестировал через POSTMAN, поэтому:
1. Откройте Postman.
2. Выберите вкладку "New" → "gRPC Request".
3. Введите адрес сервера:
``localhost:50051``
4. Тык на  "Connect".

#### На серврере включена рефлексия, поэтому все ручки апишки должны подгрузиться сами

### Вызов `Publish`
1. Выберите метод Publish.

2. Пример запроса:
```
{
"key": "news",
"data": "Hello from Postman!"
}
```
3. Нажмите "Invoke".

### Вызов `Subscribe`
1. Выберите метод Publish.

2. Пример запроса:
```
{
  "key": "news"
}
```
3. Нажмите "Invoke".
4. Ткперь в responce приходят публикации
