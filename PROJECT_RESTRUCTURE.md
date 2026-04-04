# 项目重组说明

## 重组完成时间
2026-04-05

## 重组目标
按照混合方案重组项目，将搜索引擎服务和AI打标服务分离到独立的目录中。

## 重组后的项目结构

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
│   │   └── Dockerfile           # Docker构建文件
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
│       └── Dockerfile           # Docker构建文件
│
├── shared/                      # 共享代码
│   ├── proto/                   # Protobuf定义
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
├── README_NEW.md
├── Makefile                     # 构建脚本
├── update_imports.sh            # 更新导入路径脚本
└── fix_imports.sh               # 修复导入路径脚本
```

## 主要变更

### 1. 目录结构重组
- 将搜索引擎服务移动到 `services/search-engine/`
- 将AI打标服务移动到 `services/ai-tagging/`
- 创建 `shared/` 目录用于共享代码
- 创建 `deploy/` 目录用于部署配置

### 2. Go模块更新
- 模块名称从 `MixFound` 更新为 `MixFound/services/search-engine`
- 所有导入路径已更新为新的模块路径
- 导入路径格式：`MixFound/services/search-engine/internal/{package}`

### 3. Docker配置
- 为每个服务创建了独立的 `Dockerfile`
- 创建了 `docker-compose.yml` 用于统一部署
- 配置了服务依赖和网络

### 4. 构建脚本
- 创建了 `Makefile` 用于简化构建和部署
- 提供了常用命令：build, run, stop, logs等

## 使用指南

### 本地开发

#### 搜索引擎服务
```bash
cd services/search-engine
go run cmd/main.go
```

#### AI打标服务
```bash
cd services/ai-tagging
python cmd/main.py
```

### Docker部署

#### 启动所有服务
```bash
make run
```

#### 查看服务状态
```bash
make status
```

#### 查看日志
```bash
make logs
```

#### 停止所有服务
```bash
make stop
```

## 验证结果

### 编译验证
- ✅ 搜索引擎服务编译成功
- ✅ 导入路径更新正确
- ✅ 模块名称更新正确

### 文件移动验证
- ✅ 搜索引擎服务文件移动完成
- ✅ AI打标服务文件移动完成
- ✅ 配置文件创建完成

## 后续工作

### 1. 消息队列集成
- 在 `services/search-engine/internal/queue/` 中实现消费者
- 在 `services/ai-tagging/internal/queue/` 中实现生产者和消费者

### 2. 共享代码
- 将共享的消息模型移动到 `shared/proto/`
- 创建共享的配置文件

### 3. 部署配置
- 创建 Kubernetes 配置文件
- 创建 CI/CD 流程

### 4. 文档完善
- 更新 README.md
- 添加 API 文档
- 添加部署文档

## 注意事项

1. **导入路径**：所有Go文件的导入路径已更新，请确保使用新的导入路径
2. **模块名称**：Go模块名称已更新为 `MixFound/services/search-engine`
3. **配置文件**：配置文件路径可能需要调整
4. **数据目录**：数据目录已配置为 `../../data/`，请确保路径正确

## 脚本说明

### update_imports.sh
更新Go模块名称和导入路径

### fix_imports.sh
修复导入路径，将包路径从根目录改为 `internal/` 目录

## 联系方式

如有问题，请联系项目维护者。
