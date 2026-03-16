# 电商智能搜索引擎实施指南 (E-commerce AI Search Engine Implementation Guide)

本文档旨在指导如何基于 `GoFound` 的核心理念，从零复现并增强一个适用于电商场景的智能搜索引擎。该引擎具备中文分词、倒排索引、AI 图片打标及中英文混合搜索能力。

## 0. 项目概览 (Project Overview)

本项目旨在构建一个**高性能、智能化、自研核心**的垂直电商搜索引擎。它不仅具备传统搜索引擎的文本检索能力，还融合了最新的 AI 视觉技术，能够理解商品图片内容，实现“搜你所想，见你未见”的搜索体验。

### 核心目标
1.  **自研搜索引擎内核**: 摒弃 ES 等重型框架，使用 Go + LevelDB 手写倒排索引，掌握核心搜索技术。
2.  **AI 视觉增强**: 利用 OpenAI CLIP 模型自动分析商品图片，生成“红色”、“复古”、“连衣裙”等语义标签，解决商家漏填属性的问题。
3.  **中英文混合搜索**: 实现用户搜中文（“红色”），系统自动匹配英文标签（“red”），打破语言障碍。
4.  **实时数据同步**: 接入 MySQL Binlog，实现商品上架秒级可搜，无需人工同步。

### 主要功能
*   **商品搜索**: 支持全文检索、关键词高亮、分页排序。
*   **智能联想**: 输入“红”，自动联想“红色连衣裙”。
*   **以图搜文**: 图片里的隐性特征（风格、材质、场景）均可被检索。
*   **自动化运营**: 自动打标、自动翻译、自动同步，极大降低运营成本。

### 技术方案摘要
*   **语言**: Go (核心引擎), Python (AI Agent)
*   **存储**: LevelDB (索引), MySQL (源数据), Redis (缓存)
*   **AI 模型**: OpenAI CLIP (视觉理解)
*   **中间件**: Canal (Binlog 同步)
*   **架构**: 微服务 + 插件化，各模块松耦合。

## 1. 核心架构设计 (Core Architecture Design)

本项目采用**多语言、事件驱动、多级缓存**的混合架构，确保搜索的实时性、准确性与高性能。

### 1.1 系统架构图
```text
[ 用户端 ]
    | (HTTP REST)
    v
[ 搜索服务层 (Go / Gin) ] <------------> [ 翻译服务 (Go / Redis / MySQL) ]
    | (本地调用)                                  | (外部调用)
    |                                             v
    +---> [ 搜索引擎内核 (LevelDB) ]         [ 外部翻译 API ]
    |           (倒排/正排索引)
    |
[ 数据同步层 (Go / Canal) ] <------------> [ AI 打标服务 (Python / CLIP) ]
    | (监听 Binlog)                               | (图像识别)
    v                                             v
[ 权威数据库 (MySQL) ]                       [ 图像存储 (OSS/Local) ]
```
```text
[ MySQL ] --(Binlog)--> [ Canal Server ]
|
v (TCP/Kafka)
[ Go Sync Consumer ] <------------------ [ AI Tagging Agent (Python) ]
|        (调用打标 & 翻译)                   (CLIP Model)
v
[ Go Search Engine ] <----(本地调用)----> [ LevelDB Storage ]
|                                    (倒排索引/正排数据)
v
[ Client API (REST) ]
```
### 1.2 核心组件说明
*   **Search Engine (Go)**: 采用高性能 Go 编写，负责接收用户请求、查询扩展、执行倒排索引检索、聚合结果及高亮。
*   **AI Tagging Agent (Python)**: 运行 OpenAI CLIP 模型，将图像转换为高维特征向量并与标签库匹配，生成语义标签。
*   **Canal Client**: 监听 MySQL 的二进制日志（Binlog），捕获 `INSERT/UPDATE/DELETE` 事件，实现搜索数据的近实时同步。
*   **Translator Service**: 采用三层缓存策略（Map/Redis/MySQL），将中文分词结果映射为对应的英文语义词。
*   **LevelDB**: 高性能嵌入式 KV 存储，用于持久化存储索引数据。

### 1.3 核心数据流转 (Data Flow)

#### 1.3.1 索引构建流 (Indexing Flow)
1.  **数据变更**: 商品在 MySQL 中被创建或修改。
2.  **事件捕获**: Canal 捕获 Binlog，Go Sync Consumer 接收到 `RowEvent`。
3.  **AI 处理**: 如果有图片 URL，Sync Consumer 调用 AI 打标服务获取英文标签。
4.  **翻译回填**: 系统利用字典将英文标签翻译为中文，作为搜索词补充。
5.  **索引写入**: 将 `标题 + 描述 + 标签` 写入 LevelDB 倒排索引。

#### 1.3.2 搜索查询流 (Search Flow)
1.  **用户输入**: 用户搜索中文关键词。
2.  **查询扩展**: 调用翻译服务，将关键词（如“红色”）扩展为中英混合词（“红色 red”）。
3.  **并发检索**: 并发在 LevelDB 多个分片中查找命中的文档 ID。
4.  **打分排序**: 根据命中词频（TF）、匹配精度、业务权重（销量等）进行综合打分。
5.  **详情提取**: 根据 ID 从正排索引中提取 JSON 详情并返回给用户。

## 2. 模块详细设计

### A. 基础设施层 (Infrastructure)
*   **MySQL**: 开启 `binlog-format=ROW`，作为权威数据源。
*   **Canal Server**: 伪装成从库，实时提取 Binlog 变更。

### B. 数据同步模块 (Go Sync Consumer)
**职责**: 消费 Canal 消息，驱动索引实时更新。
*   **核心函数**:
    *   `Listen()`: 建立 TCP 连接，循环接收 Canal Entry。
    *   `ProcessEvent(entry)`: 解析行数据，区分 `INSERT/UPDATE/DELETE`。
    *   `handleInsert(row)`: 提取图片 URL -> 调用 AI Agent -> 调用 `SearchEngine.AddIndex`。
    *   `handleDelete(row)`: 根据 ID 调用 `SearchEngine.RemoveIndex`。

### C. 搜索核心模块 (Go Search Engine)
**职责**: 管理持久化索引（LevelDB）。
*   **技术**: `syndtr/goleveldb` + `jieba` 分词。
*   **核心函数**:
    *   `AddIndex(id, text, data)`: 分词后更新 LevelDB 中的倒排和正排数据。
    *   `Search(query)`: 执行分词、倒排查找、结果合并与评分。
    *   `RemoveIndex(id)`: 从所有相关词条的倒排列表中移除该 ID。

### D. AI 图像打标模块 (Python Agent)
**职责**: 视觉理解，提供英文语义标签。
*   **技术**: `FastAPI` + `OpenAI CLIP`。
*   **核心函数**:
    *   `tag_image(url)`: 返回 Top 3 英文标签（如 `["red", "dress", "vintage"]`）。

### E. 中英文翻译模块 (Go Translator)
**职责**: 搜索时的中英互通。
*   **核心架构**: 三层缓存 (L1 内存 -> L2 Redis -> L3 MySQL/API)。
*   **核心函数**:
    *   `Translate(zhWord)`: 获取词汇翻译。
    *   `ExpandQuery(query)`: 将搜索词扩展为中英混合形式。

---

## 3. 通信协议与数据流

| 路径 | 协议 | 描述 |
| :--- | :--- | :--- |
| MySQL -> Canal | MySQL Slave | Binlog 字节流同步 |
| Canal -> Go Sync | TCP/Protobuf | 变更行数据分发 |
| Go Sync -> AI Agent | HTTP/JSON | 图像打标请求 |
| Go Sync -> LevelDB | Native Call | 索引持久化写入 |
| User -> API | HTTP/REST | 搜索请求 |

## 4. 实施路线图 (Implementation Roadmap)

本项目建议分 **4 个阶段** 实施，从基础搜索引擎到全功能 AI 增强版。

### Phase 1: 核心搜索引擎 (Search Core)
**目标**: 实现一个基于 LevelDB 的持久化倒排索引，并提供 HTTP 写入接口。

*   **技术栈**: Go, Gin, LevelDB (`goleveldb`), Jieba (`jiebago`)
*   **模块划分**:
    1.  **Storage (存储层)**
        *   `NewLevelDB(path)`: 初始化数据库。
        *   `Set(key, val)`, `Get(key)`, `Delete(key)`: 基础 KV 操作。
        *   `BatchWrite(batch)`: 批量写入，提高吞吐。
    2.  **Analysis (分词层)**
        *   `NewTokenizer(dictPath)`: 加载词典。
        *   `Cut(text)`: 分词并去除停用词。
    3.  **Engine (索引逻辑)**
        *   `AddIndex(id, text, doc)`:
            *   调用 `Cut(text)`。
            *   更新倒排索引：`Key=word, Val=IDs` (需处理 ID 列表合并)。
            *   更新正排索引：`Key=doc_id, Val=JSON`。
        *   `RemoveIndex(id)`: 标记删除或从倒排列表中移除 ID。
        *   `Search(query)`:
            *   分词 Query。
            *   多路并发读取倒排列表。
            *   **List Merging**: ID 列表取交集/并集。
            *   **Scoring**: 简单词频统计排序。
    4.  **Web API (Gin Handler)**
        *   `POST /api/index`: 单条新增。
            *   接收 JSON: `{"id": 1, "text": "...", "document": {...}}`
            *   调用 `Engine.AddIndex`。
        *   `POST /api/index/batch`: 批量新增。
            *   接收 JSON 数组: `[{"id": 1...}, {"id": 2...}]`
            *   循环调用 `Engine.AddIndex` (或优化为 Batch 操作)。

### Phase 2: 数据同步系统 (Data Sync)
**目标**: 接入 MySQL Binlog，实现数据自动流入搜索引擎。

*   **技术栈**: MySQL, Canal (`canal-go`), Protobuf
*   **模块划分**:
    1.  **Canal Client (消费者)**
        *   `Connect(host, port)`: 连接 Canal Server。
        *   `Subscribe(filter)`: 订阅 `shop.products` 表变更。
    2.  **Event Processor (事件处理)**
        *   `HandleInsert(row)`: 提取字段 -> 转换格式 -> 调用 Engine.AddIndex。
        *   `HandleUpdate(oldRow, newRow)`: 差异对比 -> 增量更新索引。
        *   `HandleDelete(row)`: 调用 Engine.RemoveIndex。

### Phase 3: AI 增强与多语言 (AI & Translation)
**目标**: 集成 CLIP 图片打标和中英文自动翻译。

*   **技术栈**: Python (FastAPI, PyTorch, CLIP), Redis, MySQL
*   **模块划分**:
    1.  **AI Agent (Python Service)**
        *   `load_model()`: 加载 CLIP 模型与标签库。
        *   `api_tag_image(url)`: 下载图片 -> 推理 -> 返回 Top Tags。
        *   **部署**: 独立 Docker 容器，建议配置 GPU。
    2.  **Translator (Go Service)**
        *   `NewTranslator(redis, db)`: 初始化三层缓存。
        *   `Translate(zh)`: L1(Mem) -> L2(Redis) -> L3(DB) -> API。
    3.  **Integration (集成)**
        *   在 `Event Processor` 中：收到 INSERT/UPDATE 图片时，异步调用 AI Agent，将标签追加到 `Text`。
        *   在 `Engine.Search` 中：调用 Translator 扩展 Query。

### Phase 4: 接口与部署 (API & Deployment)
**目标**: 暴露 HTTP 接口，容器化部署。

*   **技术栈**: Gin, Docker, Docker Compose
*   **模块划分**:
    1.  **Web API**
        *   `GET /search?q=...`: 搜索接口。
        *   `GET /debug/doc/:id`: 查看文档详情（调试用）。
    2.  **Deployment**
        *   编写 `docker-compose.yml`: 编排 MySQL, Redis, Canal, AI Service, Search Engine。
        *   挂载 LevelDB 数据目录到宿主机，确保持久化。

## 5. 部署架构建议

```yaml
version: '3'
services:
  mysql:
    image: mysql:8.0
    command: --binlog-format=ROW
  canal:
    image: canal/canal-server
    environment:
      - canal.instance.master.address=mysql:3306
  redis:
    image: redis:alpine
  ai-agent:
    build: ./ai-agent
    ports: ["8000:8000"]
  search-engine:
    build: ./search-engine
    ports: ["5678:5678"]
    volumes:
      - ./data:/app/data # LevelDB 数据持久化
    depends_on:
      - canal
      - redis
      - ai-agent
```