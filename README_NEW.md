# MixFound 智能搜索与 AI 标注系统

## 项目结构

```
MixFound/
├── services/                    # 所有服务
│   ├── search-engine/           # 搜索引擎服务（Go）
│   │   ├── cmd/                 # 入口文件
│   │   │   └── main.go
│   │   ├── internal/            # 内部代码
│   │   │   ├── searcher/        # 搜索引擎核心
│   │   │   ├── web/             # Web服务
│   │   │   ├── queue/           # 消息队列（消费者）
│   │   │   ├── core/            # 核心初始化
│   │   │   ├── global/          # 全局配置
│   │   │   └── redis/           # Redis客户端
│   │   ├── pkg/                 # 可导出的包
│   │   ├── go.mod               # Go依赖
│   │   ├── go.sum
│   │   ├── Dockerfile           # Docker构建文件
│   │   └── README.md
│   │
│   └── ai-tagging/              # AI打标服务（Python）
│       ├── cmd/                 # 入口文件
│       │   └── main.py
│       ├── internal/            # 内部代码
│       │   ├── app/             # 应用代码
│       │   │   ├── api/
│       │   │   ├── models/
│       │   │   ├── schemas/
│       │   │   ├── services/
│       │   │   └── config.py
│       │   └── queue/           # 消息队列
│       │       ├── consumer/
│       │       └── producer/
│       ├── pkg/                 # 可导出的包
│       ├── requirements.txt     # Python依赖
│       ├── Dockerfile           # Docker构建文件
│       └── README.md
│
├── shared/                      # 共享代码
│   ├── proto/                   # Protobuf定义（如果使用gRPC）
│   ├── docs/                    # 文档
│   └── scripts/                 # 共享脚本
│
├── deploy/                      # 部署配置
│   ├── docker/                  # Docker配置
│   │   └── docker-compose.yml
│   ├── k8s/                     # Kubernetes配置
│   └── scripts/                 # 部署脚本
│
├── .gitignore
├── README.md
└── Makefile                     # 构建脚本
```

## 快速开始

### 1. 使用 Docker Compose 启动所有服务

```bash
make run
```

### 2. 查看服务状态

```bash
make status
```

### 3. 查看日志

```bash
make logs
```

### 4. 停止所有服务

```bash
make stop
```

## 本地开发

### 搜索引擎服务

```bash
# 进入服务目录
cd services/search-engine

# 安装依赖
go mod download

# 运行服务
go run cmd/main.go
```

### AI打标服务

```bash
# 进入服务目录
cd services/ai-tagging

# 创建虚拟环境
python -m venv venv
source venv/bin/activate  # Linux/Mac
# 或
venv\Scripts\activate  # Windows

# 安装依赖
pip install -r requirements.txt

# 运行服务
python cmd/main.py
```

## 服务端口

- **搜索引擎服务**: http://localhost:8080
- **AI打标服务**: http://localhost:8081
- **Redis**: localhost:6379
- **RabbitMQ**: localhost:5672
- **RabbitMQ管理界面**: http://localhost:15672

## 技术栈

### 搜索引擎服务
- **语言**: Go 1.21
- **Web框架**: Gin
- **存储**: LevelDB
- **缓存**: Redis
- **消息队列**: RabbitMQ

### AI打标服务
- **语言**: Python 3.13
- **Web框架**: FastAPI
- **AI模型**: OpenAI CLIP
- **消息队列**: RabbitMQ

## 架构设计

```
[用户] → [搜索引擎服务] → [RabbitMQ] → [AI打标服务]
                                    ↓
                              [RabbitMQ] → [搜索引擎服务]
```

## 开发指南

### 添加新服务

1. 在 `services/` 目录下创建新的服务目录
2. 创建 `cmd/` 和 `internal/` 目录
3. 创建 `Dockerfile`
4. 更新 `deploy/docker/docker-compose.yml`
5. 更新 `Makefile`

### 添加共享代码

1. 在 `shared/` 目录下添加共享代码
2. 更新文档

## 部署

### Docker部署

```bash
# 构建所有服务
make build

# 运行所有服务
make run
```

### Kubernetes部署

```bash
# 应用Kubernetes配置
kubectl apply -f deploy/k8s/
```

## 监控

### 查看服务状态

```bash
make status
```

### 查看日志

```bash
make logs
```

### 进入容器

```bash
# 进入搜索引擎容器
make shell-search

# 进入AI打标容器
make shell-ai

# 进入Redis容器
make shell-redis

# 进入RabbitMQ容器
make shell-rabbitmq
```

## 故障排查

### 服务无法启动

1. 检查端口是否被占用
2. 检查依赖服务是否启动
3. 查看日志：`make logs`

### 消息队列问题

1. 检查RabbitMQ是否启动：`make status`
2. 访问RabbitMQ管理界面：http://localhost:15672
3. 查看队列状态

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

MIT License
