# Inference gRPC Server

Простой gRPC сервер для детекции объектов на изображениях.

## Структура проекта

```
System-Design-patterns/
├── proto/
│   └── inference/
│       └── v1/
│           └── inference.proto  # Общие proto определения
└── RTSP_PROCESSING/
    └── inference/
        ├── server.py            # gRPC сервер
        ├── client.py            # Тестовый клиент
        ├── generate_proto.sh    # Скрипт генерации
        ├── requirements.txt     # Python зависимости
        └── proto/               # Сгенерированные файлы (gitignored)
```

## Установка

1. Установите зависимости:
```bash
pip install -r requirements.txt
```

2. Сгенерируйте protobuf файлы:
```bash
chmod +x generate_proto.sh
./generate_proto.sh
```

Или вручную:
```bash
mkdir -p proto/inference/v1
python3 -m grpc_tools.protoc \
    -I../../proto \
    --python_out=./proto \
    --grpc_python_out=./proto \
    proto/v1/inference.proto
```

## Запуск

### Запуск сервера

```bash
python3 server.py
```

Сервер запустится на порту `50051`.

### Тестирование клиентом

Клиент по умолчанию отправляет `frame_000000.jpg`:

```bash
python3 client.py
```

Или укажите адрес сервера:

```bash
python3 client.py localhost:50051
```

Клиент:
- Читает изображение `frame_000000.jpg` из текущей директории
- Отправляет его на gRPC сервер
- Выводит результаты детекции в удобном формате

## API

### DetectRequest

Запрос содержит изображение в байтах:
```protobuf
message DetectRequest {
  bytes image = 1;
}
```

### DetectResponse

Ответ содержит список обнаруженных объектов:
```protobuf
message DetectResponse {
  repeated Detection detections = 1;
}

message Detection {
  string class_name = 1;
  Rectangle rectangle = 2;
}

message Rectangle {
  float x0 = 1;  // Левая верхняя точка X
  float y0 = 2;  // Левая верхняя точка Y
  float x1 = 3;  // Правая нижняя точка X
  float y1 = 4;  // Правая нижняя точка Y
}
```

## Использование proto в других микросервисах

Proto файлы находятся в общей директории `System-Design-patterns/proto/` и могут быть использованы в любых микросервисах проекта.

### Python

```bash
# Из директории вашего микросервиса
python3 -m grpc_tools.protoc \
    -I<path_to_proto_root> \
    --python_out=./proto \
    --grpc_python_out=./proto \
    <path_to_proto_root>/inference/v1/inference.proto
```

Затем импортируйте:
```python
from proto.inference.v1 import inference_pb2
from proto.inference.v1 import inference_pb2_grpc
```

### Go

```bash
protoc \
    -I<path_to_proto_root> \
    --go_out=. \
    --go-grpc_out=. \
    <path_to_proto_root>/inference/v1/inference.proto
```

Импорт в Go:
```go
import inferencev1 "github.com/patterns/proto/inference/v1"
```

## Интеграция с реальной моделью

В текущей реализации используются mock данные. Для интеграции с реальной моделью:

1. Загрузите модель в `__init__` метод `InferenceServicer`:
```python
def __init__(self):
    from ultralytics import YOLO
    self.model = YOLO('yolov8n.pt')
```

2. Замените `_mock_inference` на реальную inference:
```python
def _real_inference(self, image):
    results = self.model(image)
    detections = []
    
    for r in results:
        boxes = r.boxes
        for box in boxes:
            x1, y1, x2, y2 = box.xyxy[0].tolist()
            cls = int(box.cls[0])
            
            detections.append({
                'class_name': self.model.names[cls],
                'rectangle': {
                    'x0': x1, 'y0': y1,
                    'x1': x2, 'y1': y2
                }
            })
    
    return detections
```

## Примеры использования

### Python клиент
```python
import grpc
from proto.inference.v1 import inference_pb2
from proto.inference.v1 import inference_pb2_grpc

with grpc.insecure_channel('localhost:50051') as channel:
    stub = inference_pb2_grpc.InferenceServiceStub(channel)
    
    with open('image.jpg', 'rb') as f:
        image_bytes = f.read()
    
    request = inference_pb2.DetectRequest(image=image_bytes)
    response = stub.Detect(request)
    
    for detection in response.detections:
        print(f"Class: {detection.class_name}")
        print(f"Box: ({detection.rectangle.x0}, {detection.rectangle.y0}) - "
              f"({detection.rectangle.x1}, {detection.rectangle.y1})")
```

### Go клиент (пример)

```go
package main

import (
    "context"
    "os"
    
    "google.golang.org/grpc"
    inferencev1 "github.com/patterns/proto/inference/v1"
)

func main() {
    conn, _ := grpc.Dial("localhost:50051", grpc.WithInsecure())
    defer conn.Close()
    
    client := inferencev1.NewInferenceServiceClient(conn)
    
    imageBytes, _ := os.ReadFile("image.jpg")
    
    resp, _ := client.Detect(context.Background(), &inferencev1.DetectRequest{
        Image: imageBytes,
    })
    
    for _, det := range resp.Detections {
        println("Class:", det.ClassName)
        println("Box:", det.Rectangle.X0, det.Rectangle.Y0, 
                det.Rectangle.X1, det.Rectangle.Y1)
    }
}
```

## Производительность

- Сервер использует ThreadPoolExecutor с 10 воркерами
- Поддерживает одновременную обработку нескольких запросов
- Для production окружения рекомендуется настроить количество воркеров под вашу нагрузку

## Версионирование

Используется версионирование proto файлов (v1, v2, и т.д.) для обеспечения обратной совместимости.
При необходимости breaking changes создавайте новую версию в `proto/inference/v2/`.
