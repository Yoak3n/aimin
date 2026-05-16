# AI MIN

AI MIN (Artificial Intelligence Mind) 是一个基于仿生学架构设计的智能体系统。项目采用模块化的 Go 工作空间结构，模拟生物体的不同组成部分来构建 AI 的认知、行为和记忆能力。

## 1. 介绍

本项目旨在构建一个具有自主决策、长期记忆和环境交互能力的 AI Agent。系统架构模仿了生物体的机能组织，将功能划分为多个协同工作的模块，包括核心驱动、认知神经、基础支撑、工具执行和交互接口等层次。

## 2. 架构概览

项目包含以下核心模块（Go Workspaces）：

### 🎯 Aimin (Application Layer)
应用入口层，负责启动HTTP服务和WebSocket服务，协调各模块工作。
- **职责**: 程序主入口，初始化各组件，启动Web服务。
- **功能**:
  - **Router**: HTTP路由管理，提供RESTful API接口。
  - **WebSocket**: 实时双向通信，支持客户端状态同步和日志广播。
  - **Component**: 全局组件管理，初始化FSM状态机等核心组件。

### 🧬 DNA (Core Behavior)
核心驱动层，负责智能体的行为调度和状态管理。
- **职责**: 包含有限状态机 (FSM) 和决策树。
- **功能**:
  - **FSM**: 有限状态机实现，支持虚拟状态、工作状态和任务状态三种类型。
  - **Decision**: 决策系统，包含探索、空闲、工作、睡眠、做梦、内省等多种状态。
  - **Action**: 行为执行器，处理具体任务和对话交互。
  - **Persist**: 状态持久化，保存和恢复FSM状态。

### 🧠 Nerve (Cognitive System)
认知神经层，负责处理信息、推理和记忆管理。
- **职责**: 模拟海马体 (Hippocampus) 和大脑皮层功能。
- **功能**:
  - **Memory**: 记忆管理系统，包括短期记忆 (Temporary)、长期记忆 (Enduring) 和实体记忆 (Entity)。
  - **Hippocampus**: 海马体模拟，负责记忆的编码、存储和检索。
  - **Reason**: 逻辑推理和系统提示词生成。
  - **Controller**: 向量处理与元认知控制，管理对话记录和摘要记忆。

### 🩸 Blood (Infrastructure & Data)
基础支撑层，负责数据持久化和基础设施。
- **职责**: 提供类似于血液输送养分的功能，为其他模块提供数据支持。
- **功能**:
  - **Config**: 全局配置管理，支持数据库、LLM、TTS、工作空间等配置。
  - **Schema**: 数据模型定义，包括AI消息、图谱、任务、WebSocket消息等。
  - **Logger**: 日志系统，支持外部处理器集成。
  - **Helper**: 工具函数库，提供数据库、LLM、文件操作等辅助功能。
  - **Service**: 服务层，包含LLM聊天服务。

### 🦴 Bone (MCP Tool System)
工具执行层，基于MCP (Model Context Protocol) 协议提供各种工具能力。
- **职责**: 为AI Agent提供可调用的工具集合。
- **功能**:
  - **FileOperation**: 文件读写、创建、删除等操作。
  - **ShellCommand**: Shell命令执行，支持超时和沙箱隔离。
  - **Glob**: 文件模式匹配搜索。
  - **Grep**: 内容搜索，支持正则表达式。
  - **Skill**: 技能管理系统。
  - **Web**: 网页抓取和内容提取。
  - **TTS**: 文本转语音功能。
  - **Memory**: 记忆读写工具。
  - **Workspace**: 工作空间上下文管理，包括提示词生成和文件状态追踪。

### 🧠 Cerebrum (Agent System)
智能体层，实现ReAct (Reasoning and Acting) 模式的AI Agent。
- **职责**: 核心推理和行动循环。
- **功能**:
  - **ReActAgent**: ReAct模式Agent实现，支持工具调用和多轮推理。
  - **ConversationAgent**: 对话Agent，管理对话上下文和工具调用。
  - **A2A**: Agent-to-Agent通信协议。
  - **Compress**: 上下文压缩，优化长对话性能。
  - **ExecPlan**: 执行计划管理。
  - **Hook**: Agent钩子系统，支持思考、行动、工具结果、最终答案等事件回调。

### 🫁 Lung (LLM Interface)
LLM交互层，负责与大语言模型的通信。
- **职责**: 封装LLM API调用，提供统一的聊天接口。
- **功能**:
  - **Chat**: 聊天接口，支持普通聊天和流式聊天。
  - **Adapter**: LLM适配器系统，支持多种LLM服务提供商。
  - **Delta**: 流式响应增量处理。
  - **Hub**: 适配器管理中心，管理多个LLM适配器实例。

### 🦠 Gut (Data Access Layer)
数据访问层，负责数据库操作和数据持久化。
- **职责**: 提供统一的数据访问接口。
- **功能**:
  - **PostgreSQL**: 关系型数据库，存储对话、摘要、向量等结构化数据。
  - **Neo4j**: 图数据库，存储知识图谱和关联记忆。
  - **Cayley**: 内存图数据库，用于快速图谱查询。
  - **Conversation**: 对话记录CRUD操作。
  - **Summary**: 摘要记忆管理。
  - **Metacognition**: 元认知数据管理。
  - **Graph**: 图谱数据操作，支持节点和边的创建、查询。

### 🫘 Kidney (Retrieval System)
检索系统层，负责信息检索和向量搜索。
- **职责**: 从海量数据中快速检索相关信息。
- **功能**:
  - **VectorSearch**: 向量相似度搜索，基于嵌入向量匹配相关对话。
  - **Embedding**: 文本向量化，将文本转换为高维向量。
  - **ConversationStore**: 对话存储接口，支持相关对话记录检索。

### �️ Hand (Interaction Layer)
交互层，负责与外部环境的通信和工具执行。
- **职责**: 处理网络请求、搜索和沙箱执行。
- **功能**:
  - **Internet/Fetch**: 网页抓取，支持API调用、浏览器渲染、PDF解析等。
  - **Internet/Search**: 搜索引擎集成，包括DuckDuckGo和Bilibili搜索。
  - **Sandbox**: 沙箱管理器，安全执行Shell命令和进程管理。
  - **Interactive**: 交互式功能，包括对话、探索和TTS。
  - **Requests**: HTTP请求工具库，支持GET/POST和代理配置。

### 👅 Tongue (Conversation Manager)
对话管理层，负责对话生命周期管理。
- **职责**: 管理多轮对话的创建、维护和销毁。
- **功能**:
  - **Manager**: 对话管理器，支持多对话并发和超时处理。
  - **Conversation**: 对话实例，维护消息历史和Agent上下文。
  - **Input**: 输入处理，支持直接问答和任务执行。

### 🧬 Skin (Frontend Interface)
前端界面层，基于Vue 3的Web应用。
- **职责**: 提供用户交互界面。
- **功能**:
  - **Chat**: 聊天组件，支持消息展示、思考过程、工具调用展示。
  - **WebSocket**: 实时通信客户端，接收状态更新和日志。
  - **Router**: 前端路由管理。
  - **Store**: 状态管理，基于Pinia。
  - **Layout**: 页面布局组件。

## 3. 技术栈

- **Language**: Go 1.25+ (后端), TypeScript (前端)
- **Database**:
  - PostgreSQL with Vector plugin (结构化数据存储)
  - Neo4j (知识图谱与关联记忆)
  - Cayley (内存图数据库)
- **AI Integration**: OpenAI API (或其他兼容 LLM)
- **Frontend**: Vue 3 + Vite + TypeScript + Pinia
- **WebSocket**: 实时双向通信
- **MCP**: Model Context Protocol 工具协议

## 4. 快速开始

### 前置要求
- Go 1.25 或更高版本
- Node.js 18+ (前端开发)
- PostgreSQL 数据库实例
- Neo4j 数据库实例 (可选)
- 配置好的 `config.json` 文件 (参考 `config.example.json`)

### 构建与运行

1. **进入项目根目录**
   确保 `go.work` 文件存在并包含所有模块。

2. **下载依赖**
   ```bash
   go work sync
   ```

3. **运行后端程序**
   程序入口位于 `aimin` 模块中。
   ```bash
   cd aimin
   go run .
   ```

4. **运行前端开发服务器** (可选)
   ```bash
   cd skin
   pnpm install
   pnpm run dev
   ```

## 5. 目录结构

```
e:\Project\GoProject\codes\aimin
├── aimin/          # 应用入口层 - HTTP/WebSocket服务
├── blood/          # 基础支撑层 - 配置、日志、工具函数
├── bone/           # MCP工具层 - 文件、Shell、搜索等工具
├── cerebrum/       # 智能体层 - ReAct Agent实现
├── dna/            # 核心驱动层 - FSM状态机和决策系统
├── gut/            # 数据访问层 - PostgreSQL/Neo4j/Cayley
├── hand/           # 交互层 - 网络请求、搜索、沙箱
├── kidney/         # 过滤处理中间层 - 对话记录过滤
├── lung/           # LLM接口层 - 大模型通信
├── nerve/          # 认知层 - 记忆管理和推理
├── skin/           # 前端界面 - Vue 3 Web应用
├── tongue/         # 对话层 - 对话生命周期管理
├── docs/           # 文档目录
├── config.example.json  # 配置示例文件
├── go.work         # Go Workspace 配置
└── README.md       # 项目说明文档
```

## 6. 模块依赖关系

```
                    ┌─────────────┐
                    │   Aimin     │ (应用入口)
                    └──────┬──────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
        ▼                  ▼                  ▼
   ┌─────────┐       ┌─────────┐       ┌─────────┐
   │  Skin   │       │ Tongue  │       │   DNA   │
   │ (前端)  │       │ (对话)  │       │ (驱动)  │
   └─────────┘       └────┬────┘       └────┬────┘
                          │                  │
                          ▼                  ▼
                    ┌─────────────┐     ┌─────────┐
                    │  Cerebrum   │     │  Nerve  │
                    │ (智能体)   │     │ (认知)  │
                    └──────┬──────┘     └────┬────┘
                           │                 │
        ┌──────────────────┼─────────────────┤
        │                  │                 │
        ▼                  ▼                 ▼
   ┌─────────┐       ┌─────────┐       ┌─────────┐
   │  Bone   │       │  Lung   │       │  Kidney │
   │ (工具)  │       │ (LLM)   │       │ (过滤处理)  │
   └────┬────┘       └────┬────┘       └────┬────┘
        │                  │                 │
        └──────────────────┼─────────────────┘
                           │
                    ┌──────┴──────┐
                    │    Blood    │
                    │ (基础支撑)  │
                    └──────┬──────┘
                           │
                    ┌──────┴──────┐
                    │     Gut     │
                    │ (数据访问)  │
                    └─────────────┘
```

## 7. 核心工作流程

1. **用户输入**: 通过Skin前端或API发起请求
2. **对话管理**: Tongue模块创建和管理对话实例
3. **决策驱动**: DNA模块的FSM状态机决定当前行为模式
4. **智能推理**: Cerebrum模块的ReAct Agent进行推理和工具调用
5. **工具执行**: Bone模块提供文件操作、Shell命令等工具能力
6. **记忆存储**: Nerve模块管理短期和长期记忆
7. **数据持久化**: Gut模块将数据存储到PostgreSQL或Neo4j
8. **响应输出**: 通过WebSocket或HTTP返回处理结果
