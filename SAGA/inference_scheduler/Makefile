.PHONY: help migrate-up migrate-down migrate-status migrate-create sqlc-generate

# Цвета для вывода
GREEN  := \033[0;32m
YELLOW := \033[0;33m
NC     := \033[0m # No Color

# Захардкоженные переменные для БД
DB_USER := scheduler_user
DB_PASSWORD := scheduler_password
DB_HOST := localhost
DB_PORT := 5433
DB_NAME := inference_scheduler
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

help: ## Показать справку
	@echo "$(GREEN)Доступные команды:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

migrate-up: ## Применение миграций (up)
	@echo "$(GREEN)Применение миграций...$(NC)"
	goose -dir migrations postgres "$(DB_URL)" up
	@echo "$(GREEN)Миграции применены!$(NC)"

migrate-down: ## Откат последней миграции (down)
	@echo "$(YELLOW)Откат миграции...$(NC)"
	goose -dir migrations postgres "$(DB_URL)" down
	@echo "$(GREEN)Миграция откачена!$(NC)"

migrate-status: ## Статус миграций
	@echo "$(GREEN)Статус миграций:$(NC)"
	goose -dir migrations postgres "$(DB_URL)" status

migrate-create: ## Создание новой миграции (используйте: make migrate-create NAME=имя_миграции)
	@echo "$(GREEN)Создание новой миграции: $(NAME)$(NC)"
	goose -dir migrations create $(NAME) sql
	@echo "$(GREEN)Миграция создана!$(NC)"

migrate-reset: ## Откат всех миграций и повторное применение
	@echo "$(YELLOW)Откат всех миграций...$(NC)"
	goose -dir migrations postgres "$(DB_URL)" reset
	@echo "$(GREEN)Применение миграций...$(NC)"
	goose -dir migrations postgres "$(DB_URL)" up
	@echo "$(GREEN)База данных пересоздана!$(NC)"

sqlc-generate: ## Генерация Go кода из SQL запросов с помощью sqlc
	@echo "$(GREEN)Генерация Go кода из SQL...$(NC)"
	sqlc generate
	@echo "$(GREEN)Код успешно сгенерирован!$(NC)"

