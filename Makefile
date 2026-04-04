.PHONY: all build run stop logs clean build-search build-ai

# 默认目标
all: build

# 构建所有服务
build:
	docker-compose -f deploy/docker/docker-compose.yml build

# 运行所有服务
run:
	docker-compose -f deploy/docker/docker-compose.yml up -d

# 停止所有服务
stop:
	docker-compose -f deploy/docker/docker-compose.yml down

# 查看日志
logs:
	docker-compose -f deploy/docker/docker-compose.yml logs -f

# 清理
clean:
	docker-compose -f deploy/docker/docker-compose.yml down -v
	docker system prune -f

# 构建搜索引擎服务
build-search:
	cd services/search-engine && docker build -t mixfound-search-engine .

# 构建AI打标服务
build-ai:
	cd services/ai-tagging && docker build -t mixfound-ai-tagging .

# 运行搜索引擎服务（本地开发）
run-search-local:
	cd services/search-engine && go run cmd/main.go

# 运行AI打标服务（本地开发）
run-ai-local:
	cd services/ai-tagging && python cmd/main.py

# 查看服务状态
status:
	docker-compose -f deploy/docker/docker-compose.yml ps

# 重启所有服务
restart: stop run

# 进入搜索引擎容器
shell-search:
	docker exec -it mixfound-search-engine sh

# 进入AI打标容器
shell-ai:
	docker exec -it mixfound-ai-tagging bash

# 进入Redis容器
shell-redis:
	docker exec -it mixfound-redis redis-cli

# 进入RabbitMQ容器
shell-rabbitmq:
	docker exec -it mixfound-rabbitmq rabbitmqctl
