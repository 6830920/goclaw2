# GoClaw2 使用指南

## 快速开始

### 1. 获取智谱 API Key

访问 https://open.bigmodel.cn/ 注册并获取 API Key。

### 2. 配置

#### 方式一：环境变量（推荐）

```bash
export ZHIPU_API_KEY="your-api-key-here"
```

#### 方式二：配置文件

复制示例配置：

```bash
cp .goclaw.example.yaml ~/.goclaw.yaml
```

编辑 `~/.goclaw.yaml`，填入你的 API Key：

```yaml
zhipu:
  api_key: "your-api-key-here"
```

### 3. 运行

```bash
# 使用 Makefile
make run

# 或直接运行
./bin/goclaw chat
```

## 命令参考

### 交互式对话

```bash
goclaw chat
```

交互式命令：
- `/help` - 显示帮助和可用工具
- `/clear` - 清空对话历史
- `/quit` 或 `/exit` - 退出程序

### 查看配置

```bash
goclaw config
```

输出示例：
```
Current Configuration:
  Zhipu API Key: abc1...xyz9
  Zhipu Model: glm-4-flash
  Temperature: 0.70
  Max Tokens: 4096
  Memory Path: ./goclaw.db
  Max History: 50
  Message Count: 12
```

### 管理记忆

```bash
# 显示对话历史
goclaw memory show

# 清空对话历史
goclaw memory clear
```

## 工具使用示例

### 1. 读取文件

```
You: 帮我读取 README.md 文件
AI: [自动调用 read_file 工具]
文件内容已读取，README.md 包含以下内容...
```

### 2. 写入文件

```
You: 创建一个名为 hello.txt 的文件，内容是 "Hello, GoClaw2!"
AI: [自动调用 write_file 工具]
文件已创建成功。
```

### 3. 列出目录

```
You: 显示当前目录的文件
AI: [自动调用 list_dir 工具]
当前目录包含：
  [FILE] main.go
  [FILE] go.mod
  [DIR]  internal/
```

### 4. 执行命令

```
You: 运行 ls -la 命令
AI: [自动调用 exec_command 工具]
命令执行结果：
total 24
drwxr-xr-x  5 user  staff   160 Feb  4 08:00 .
...
```

## 开发

### 构建

```bash
# 使用 Makefile
make build

# 或手动构建
go build -o bin/goclaw ./cmd/goclaw/
```

### 安装到系统

```bash
make install
```

这会将二进制文件复制到 `~/go/bin/`。确保 `~/go/bin` 在你的 PATH 中：

```bash
export PATH=$PATH:~/go/bin
```

### 清理

```bash
make clean
```

## 高级配置

### 修改模型

在 `~/.goclaw.yaml` 中修改：

```yaml
zhipu:
  model: "glm-4"  # 或 glm-4-plus, glm-4-flash
```

可用模型：
- `glm-4-flash` - 快速响应（默认）
- `glm-4` - 标准性能
- `glm-4-plus` - 高性能

### 调整温度

```yaml
zhipu:
  temperature: 0.9  # 更有创造性
  # 或
  temperature: 0.2  # 更确定
```

温度范围：0.0 - 2.0
- 低值：更确定、更一致的输出
- 高值：更有创造性、更多样化的输出

### 修改历史记录数量

```yaml
agent:
  max_history: 100  # 保留更多历史
```

## 故障排除

### 问题：API Key 无效

错误信息：`zhipu api_key is required`

**解决方法**：
1. 检查环境变量：`echo $ZHIPU_API_KEY`
2. 或检查配置文件：`cat ~/.goclaw.yaml`

### 问题：数据库锁定

错误信息：`database is locked`

**解决方法**：
```bash
# 删除数据库文件
rm goclaw.db

# 或使用不同的数据库路径
goclaw chat --config custom-config.yaml
```

### 问题：命令执行失败

错误信息：`command failed`

**解决方法**：
- 确保命令在当前环境中可用
- 使用完整路径：`/usr/bin/ls` 而不是 `ls`
- 检查文件权限

## 安全建议

1. **不要在公共仓库中提交 API Key**
   - 使用 `.gitignore` 排除配置文件
   - 使用环境变量而非配置文件

2. **限制命令执行权限**
   - 生产环境考虑添加命令白名单
   - 使用 chroot 或容器隔离

3. **文件访问控制**
   - 限制可访问的目录
   - 添加文件大小限制

## 下一步

- [ ] 尝试多轮对话
- [ ] 创建自定义工具
- [ ] 集成到自动化流程
- [ ] 探索语义搜索（Phase 3）

## 更多信息

- 项目主页：https://github.com/user/goclaw2
- 智谱 AI 文档：https://open.bigmodel.cn/dev/api
