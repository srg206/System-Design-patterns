# Proto Definitions

Общие Protocol Buffers определения для всех микросервисов проекта.

## Структура

```
proto/
├── inference/
│   └── v1/
│       └── inference.proto
└── README.md
```

## Версионирование

Все proto файлы организованы по версиям (v1, v2, и т.д.) для обеспечения обратной совместимости.

## Генерация кода

### Python

Для Python микросервисов используйте:

```bash
# Из корня проекта
python3 -m grpc_tools.protoc \
    -I./System-Design-patterns/proto \
    --python_out=<target_directory> \
    --grpc_python_out=<target_directory> \
    System-Design-patterns/proto/inference/v1/inference.proto
```

Пример для inference сервиса:

```bash
cd System-Design-patterns/RTSP_PROCESSING/inference
python3 -m grpc_tools.protoc \
    -I../../proto \
    --python_out=./proto \
    --grpc_python_out=./proto \
    ../../proto/inference/v1/inference.proto
```

### Go

Для Go микросервисов:

```bash
# Установите необходимые инструменты
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Генерация
protoc \
    -I./System-Design-patterns/proto \
    --go_out=. \
    --go-grpc_out=. \
    System-Design-patterns/proto/inference/v1/inference.proto
```

## Использование в микросервисах

### Python

После генерации импортируйте:

```python
from proto.inference.v1 import inference_pb2
from proto.inference.v1 import inference_pb2_grpc
```

### Go

```go
import inferencev1 "github.com/patterns/proto/inference/v1"
```

## Добавление новых proto файлов

1. Создайте новую директорию для сервиса: `proto/<service_name>/v1/`
2. Добавьте `.proto` файл с правильным package: `<service_name>.v1`
3. Укажите `go_package` для Go: `github.com/patterns/proto/<service_name>/v1;<service_name>v1`
4. Обновите документацию

## Миграция версий

При breaking changes создайте новую версию:

```
proto/
├── inference/
│   ├── v1/
│   │   └── inference.proto
│   └── v2/
│       └── inference.proto
```

## Доступные сервисы

### inference/v1

Сервис для детекции объектов на изображениях.

**Методы:**
- `Detect(DetectRequest) returns (DetectResponse)` - детекция объектов на изображении

**Сообщения:**
- `DetectRequest` - запрос с изображением в байтах
- `DetectResponse` - ответ со списком обнаруженных объектов
- `Detection` - один обнаруженный объект (класс + координаты)
- `Rectangle` - координаты bounding box (x0, y0, x1, y1)

