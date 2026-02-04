# GoClaw2

Go 语言版 OpenClaw - 一个支持工具调用的 AI 助手。

## 功能特性

- **对话能力** - 支持智谱 GLM-4 模型
- **工具执行** - 文件读写、命令执行、目录列表
- **记忆系统** - SQLite 持久化存储会话历史
- **函数调用** - 支持智谱 API 的 Function Calling

## 安装

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/user/goclaw2.git
cd goclaw2

# 构建
go build -o bin/goclaw ./cmd/goclaw/

# 添加到 PATH
export PATH=$PATH:$(pwd)/bin
```

## 配置

### 环境变量

```bash
export ZHIPU_API_KEY="your-api-key"
```

### 配置文件

创建 `~/.goclaw.yaml`:

```yaml
zhipu:
  api_key: "your-api-key"
  base_url: "https://open.bigmodel.cn/api/paas/v4"
  model: "glm-4-flash"
  temperature: 0.7
  max_tokens: 4096

agent:
  max_history: 50

memory:
  type: "sqlite"
  file_path: "./goclaw.db"

gateway:
  enabled: false
  port: 8080
  host: "localhost"
```

## 使用

### 交互式对话

```bash
goclaw chat
```

### 可用命令

- `/help` - 显示帮助和可用工具
- `/clear` - 清空对话历史
- `/quit` - 退出程序

### 其他命令

```bash
# 显示配置
goclaw config

# 显示对话历史
goclaw memory show

# 清空对话历史
goclaw memory clear
```

## 工具使用

### 读取文件

```
You: 帮我读取 README.md 文件
AI: [调用 read_file 工具]
文件内容是: ...
```

### 写入文件

```
You: 创建一个名为 test.txt 的文件，内容是 "Hello World"
AI: [调用 write_file 工具]
文件已创建成功。
```

### 列出目录

```
You: 列出当前目录的文件
AI: [调用 list_dir 工具]
目录内容:
  [FILE] main.go
  [FILE] go.mod
  [DIR]  internal/
```

### 执行命令

```
You: 运行 ls -la 命令
AI: [调用 exec_command 工具]
命令输出: total 24...
```

## 项目结构

```
goclaw2/
├── cmd/
│   └── goclaw/          # CLI 入口
├── internal/
│   ├── config/          # 配置管理
│   ├── provider/        # AI 提供商 (智谱)
│   ├── agent/           # Agent 运行时
│   ├── memory/          # 记忆系统
│   └── tools/           # 工具执行
├── pkg/
│   └── api/             # 公开 API
├── go.mod
└── README.md
```

## 技术栈

- **语言**: Go 1.19+
- **CLI**: cobra
- **配置**: viper
- **数据库**: modernc.org/sqlite (纯 Go SQLite)
- **AI**: 智谱 GLM-4 API

## 开发计划

- [x] Phase 1: 基础对话
- [x] Phase 2: 工具执行
- [ ] Phase 3: 语义搜索
- [ ] Phase 4: WebSocket Gateway

## 安全提示

⚠️ **警告**: 工具执行没有权限控制，请注意：
- 命令注入风险
- 文件系统访问
- 生产环境需添加沙箱机制

## 许可证

MIT License

## 致谢

本项目灵感来源于 OpenClaw (TypeScript 版本)。
