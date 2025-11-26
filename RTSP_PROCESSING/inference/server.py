import grpc
from concurrent import futures
import time
import io
from PIL import Image
import numpy as np
import sys
import os

# Добавляем proto директорию в путь Python
sys.path.append(os.path.join(os.path.dirname(__file__), 'proto'))

# Импорты сгенерированных protobuf файлов
try:
    from inference.v1 import inference_pb2
    from inference.v1 import inference_pb2_grpc
except ImportError:
    # Альтернативный способ импорта
    from proto.inference.v1 import inference_pb2
    from proto.inference.v1 import inference_pb2_grpc

    
class InferenceServicer(inference_pb2_grpc.InferenceServiceServicer):
    """Реализация gRPC сервиса для детекции объектов"""
    
    def __init__(self):
        print("Инициализация InferenceService...")
        # Здесь можно загрузить модель (YOLO, TensorFlow, PyTorch и т.д.)
        # self.model = load_model()
        
    def Detect(self, request, context):
        """
        Обрабатывает запрос на детекцию объектов на изображении.
        
        Args:
            request: DetectRequest с изображением в байтах
            context: gRPC context
            
        Returns:
            DetectResponse со списком обнаруженных объектов
        """
        try:
            # Декодируем изображение из байтов
            image_bytes = request.image
            image = Image.open(io.BytesIO(image_bytes))
            
            # Конвертируем в numpy array если нужно
            img_array = np.array(image)
            
            print(f"Получено изображение: размер {image.size}, формат {image.format}")
            
            # Здесь должна быть реальная inference модели
            # Пример: detections = self.model.predict(img_array)
            
            # Для демонстрации возвращаем mock данные
            detections = self._mock_inference(image)
            
            # Формируем ответ
            response = inference_pb2.DetectResponse()
            
            for det in detections:
                detection = response.detections.add()
                detection.class_name = det['class_name']
                
                # Заполняем координаты прямоугольника
                detection.rectangle.x0 = det['rectangle']['x0']
                detection.rectangle.y0 = det['rectangle']['y0']
                detection.rectangle.x1 = det['rectangle']['x1']
                detection.rectangle.y1 = det['rectangle']['y1']
            
            print(f"Обнаружено объектов: {len(detections)}")
            return response
            
        except Exception as e:
            print(f"Ошибка при обработке изображения: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Ошибка обработки: {str(e)}")
            return inference_pb2.DetectResponse()
    
    def _mock_inference(self, image):
        """
        Mock функция для имитации работы модели детекции.
        В реальном приложении здесь должна быть реальная модель (YOLO, etc.)
        
        Args:
            image: PIL Image объект
            
        Returns:
            Список обнаруженных объектов
        """
        width, height = image.size
        
        # Возвращаем несколько тестовых детекций
        mock_detections = [
            {
                'class_name': 'person',
                'rectangle': {
                    'x0': width * 0.1,
                    'y0': height * 0.2,
                    'x1': width * 0.4,
                    'y1': height * 0.8
                }
            },
            {
                'class_name': 'car',
                'rectangle': {
                    'x0': width * 0.5,
                    'y0': height * 0.3,
                    'x1': width * 0.9,
                    'y1': height * 0.7
                }
            }
        ]
        
        return mock_detections


def serve(port=50051):
    """
    Запускает gRPC сервер
    
    Args:
        port: Порт для прослушивания
    """
    # Создаем gRPC сервер с пулом потоков
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    
    # Регистрируем наш сервис
    inference_pb2_grpc.add_InferenceServiceServicer_to_server(
        InferenceServicer(), server
    )
    
    # Привязываем к порту
    server.add_insecure_port(f'[::]:{port}')
    
    # Запускаем сервер
    server.start()
    print(f"✓ gRPC сервер запущен на порту {port}")
    print("Ожидание запросов...")
    
    try:
        # Держим сервер запущенным
        while True:
            time.sleep(86400)  # 24 часа
    except KeyboardInterrupt:
        print("\nОстановка сервера...")
        server.stop(0)


if __name__ == '__main__':
    serve()
