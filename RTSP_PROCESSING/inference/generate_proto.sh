#!/bin/bash

# Скрипт для генерации Python кода из .proto файла

# Путь к proto файлам (относительно корня проекта)
PROTO_DIR="../../proto"
PROTO_FILE="inference/v1/inference.proto"
OUTPUT_DIR="./proto"

# Создаем директорию для сгенерированных файлов
mkdir -p "$OUTPUT_DIR/inference/v1"

# Создаем __init__.py файлы для Python модулей
touch "$OUTPUT_DIR/__init__.py"
touch "$OUTPUT_DIR/inference/__init__.py"
touch "$OUTPUT_DIR/inference/v1/__init__.py"

# Генерируем Python код из protobuf
python3 -m grpc_tools.protoc \
    -I"$PROTO_DIR" \
    --python_out="$OUTPUT_DIR" \
    --grpc_python_out="$OUTPUT_DIR" \
    "$PROTO_DIR/$PROTO_FILE"



echo "✓ Protobuf файлы сгенерированы успешно"
echo "  - $OUTPUT_DIR/inference/v1/inference_pb2.py"
echo "  - $OUTPUT_DIR/inference/v1/inference_pb2_grpc.py"

