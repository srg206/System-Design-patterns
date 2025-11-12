.PHONY: help swagger docker-build docker-up docker-down docker-restart docker-logs run migrate-up migrate-down migrate-status migrate-create

# Цвета для вывода
GREEN  := \033[0;32m
YELLOW := \033[0;33m
NC     := \033[0m # No Color

help: ## Показать справку
	@echo "$(GREEN)Доступные команды:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

swag: ## Генерация Swagger документации
	@echo "$(GREEN)Генерация Swagger документации...$(NC)"
	swag init -g cmd/api/main.go -o docs
	@echo "$(GREEN)Swagger документация успешно сгенерирована!$(NC)"

build: ## Сборка Docker контейнеров
	@echo "$(GREEN)Сборка Docker контейнеров...$(NC)"
	docker compose build
	@echo "$(GREEN)Контейнеры успешно собраны!$(NC)"

up: ## Запуск Docker контейнеров
	@echo "$(GREEN)Запуск Docker контейнеров...$(NC)"
	docker compose up -d
	@echo "$(GREEN)Контейнеры запущены!$(NC)"
	@echo "API доступен по адресу: http://localhost:3000"
	@echo "Swagger UI: http://localhost:3000/swagger/index.html"
	@echo "Producer: http://localhost:3001"

down: ## Остановка Docker контейнеров
	@echo "$(YELLOW)Остановка Docker контейнеров...$(NC)"
	docker compose down
	@echo "$(GREEN)Контейнеры остановлены!$(NC)"

restart: docker-down docker-up ## Перезапуск Docker контейнеров

up-build: down build up ## Перезапуск Docker контейнеров

docker-logs: ## Просмотр логов контейнеров
	docker compose logs -f

docker-logs-api: ## Просмотр логов API контейнера
	docker compose logs -f api

docker-logs-producer: ## Просмотр логов Producer контейнера
	docker compose logs -f producer

docker-rebuild: docker-down docker-build docker-up ## Пересборка и перезапуск контейнеров

run: ## Запуск API локально (без Docker)
	@echo "$(GREEN)Запуск API...$(NC)"
	go run cmd/api/main.go

run-producer: ## Запуск Producer локально (без Docker)
	@echo "$(GREEN)Запуск Producer...$(NC)"
	go run cmd/producer/main.go

tidy: ## Обновление зависимостей Go
	@echo "$(GREEN)Обновление зависимостей...$(NC)"
	go mod tidy
	@echo "$(GREEN)Зависимости обновлены!$(NC)"

install-swag: ## Установка swag CLI
	@echo "$(GREEN)Установка swag...$(NC)"
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)swag установлен!$(NC)"

dev: swagger run ## Генерация Swagger и запуск API локально

clean: ## Очистка сгенерированных файлов и Docker ресурсов
	@echo "$(YELLOW)Очистка...$(NC)"
	docker compose down -v --remove-orphans
	rm -rf docs/docs.go docs/swagger.json docs/swagger.yaml
	@echo "$(GREEN)Очистка завершена!$(NC)"

# Миграции базы данных
# Загружаем переменные окружения из app.env
define load_env
	set -a && source app.env && set +a &&
endef

migrate-up: ## Применение миграций (up)
	@echo "$(GREEN)Применение миграций...$(NC)"
	@$(load_env) goose -dir migrations postgres "postgres://$${DB_USER}:$${DB_PASSWORD}@$${DB_HOST}:$${DB_PORT}/$${DB_NAME}?sslmode=disable" up
	@echo "$(GREEN)Миграции применены!$(NC)"

migrate-down: ## Откат миграций (down)
	@echo "$(YELLOW)Откат миграций...$(NC)"
	@$(load_env) goose -dir migrations postgres "postgres://$${DB_USER}:$${DB_PASSWORD}@$${DB_HOST}:$${DB_PORT}/$${DB_NAME}?sslmode=disable" down
	@echo "$(GREEN)Миграции откатаны!$(NC)"

migrate-status: ## Статус миграций
	@echo "$(GREEN)Статус миграций:$(NC)"
	@$(load_env) goose -dir migrations postgres "postgres://$${DB_USER}:$${DB_PASSWORD}@$${DB_HOST}:$${DB_PORT}/$${DB_NAME}?sslmode=disable" status

migrate-create: ## Создание новой миграции (используйте: make migrate-create NAME=имя_миграции)
	@echo "$(GREEN)Создание новой миграции: $(NAME)$(NC)"
	goose -dir migrations postgres "$(NAME)" create
	@echo "$(GREEN)Миграция создана!$(NC)"

