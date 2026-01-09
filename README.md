# AI MIN

AI MIN (Artificial Intelligence Mind) 是一个基于仿生学架构设计的智能体系统。项目采用模块化的 Go 工作空间结构，模拟生物体的不同组成部分来构建 AI 的认知、行为和记忆能力。

## 1. 介绍

本项目旨在构建一个具有自主决策、长期记忆和环境交互能力的 AI Agent。系统架构模仿了生物体的机能组织，将功能划分为 DNA（核心驱动）、Nerve（认知神经）、Blood（基础支撑）和 Face（交互接口）四个主要部分。

## 2. 架构概览

项目包含以下核心模块（Go Workspaces）：

### 🧬 DNA (Core Behavior)
核心驱动层，负责智能体的行为调度和状态管理。
- **职责**: 包含有限状态机 (FSM) 和决策树。
- **功能**: 决定智能体当前应该处于什么状态（探索、空闲、工作、睡眠等），是程序的入口 (`main` package)。

### 🧠 Nerve (Cognitive System)
认知神经层，负责处理信息、推理和记忆管理。
- **职责**: 模拟海马体 (Hippocampus) 和大脑皮层功能。
- **功能**:
  - **Memory**: 管理短期记忆 (Temporary) 和长期记忆 (Enduring)。
  - **Reason**: 逻辑推理和系统提示词生成。
  - **Controller**: 向量处理与元认知控制。

### 🩸 Blood (Infrastructure & Data)
基础支撑层，负责数据持久化和基础设施。
- **职责**: 提供类似于血液输送养分的功能，为其他模块提供数据支持。
- **功能**:
  - **DAO**: 数据库访问对象，支持 PostgreSQL (关系型数据) 和 Neo4j (图谱数据)。
  - **Config**: 全局配置管理。
  - **Adapter**: 外部服务适配器。

### 😶 Face (Interaction Layer)
交互接口层，负责与外部环境的沟通。
- **职责**: 管理对话和对外接口。
- **功能**:
  - **Conversation**: 对话上下文管理、会话生命周期维护。

## 3. 技术栈

- **Language**: Go 1.25+
- **Database**:
  - PostgreSQL (结构化数据存储)
  - Neo4j (知识图谱与关联记忆)
- **AI Integration**: OpenAI API (或其他兼容 LLM)

## 4. 快速开始

### 前置要求
- Go 1.25 或更高版本
- PostgreSQL 数据库实例
- Neo4j 数据库实例
- 配置好的 `config` 文件 (需根据 `blood/config` 设置数据库连接串)

### 构建与运行

1. **进入项目根目录**
   确保 `go.work` 文件存在并包含所有模块。

2. **下载依赖**
   ```bash
   go work sync
   ```

3. **运行程序**
   程序入口位于 `dna` 模块中。
   ```bash
   cd dna
   go run .
   ```

## 5. 目录结构

```
e:\Project\GoProject\codes\aimin
├── blood/      # 基础设施与数据层
├── dna/        # 核心状态机与入口
├── face/       # 交互与对话管理
├── nerve/      # 记忆与推理引擎
└── go.work     # Go Workspace 配置
```
