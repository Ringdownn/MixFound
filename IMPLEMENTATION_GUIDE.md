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

## 1. 核心架构设计

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