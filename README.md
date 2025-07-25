# 🍅 Tomato Clock

一个用 **Go** 和 **[Fyne](https://fyne.io/)** 构建的跨平台番茄钟应用程序，助你专注工作与学习。

## 功能特性

- **番茄钟**：支持正计时 *(Count-up)* 与倒计时 *(Count-down)* 两种模式。
- **任务管理**：可为每段专注时间关联任务，并自动持久化到本地 JSON 文件。
- **数据统计**：
    - 顶部实时显示最近 24 小时专注时长（按任务标签聚合）。
    - 双色饼图直观展示各项专注时间占比。
    - 饼图跟踪每日学习目标（默认为 8 小时）的完成进度。
- **自然语言助手与顶栏对话**：在 GUI 顶栏通过输入框即可调用 DeepSeek / OpenAI Chat，快速增删专注记录或提出问题，无需命令行。
- **极简现代 UI**：缩小饼图、图标化按钮、响应式顶栏，整体视觉更轻盈现代。
- **声音提醒**：计时结束时播放 `resources/sounds/alert.mp3`，并支持随机正念提示音，可在应用内静音。
- **离线存储**：所有数据均保存在用户目录 `.tomato_clock.json`，无需数据库。
- **跨平台**：得益于 Fyne，可在 **Windows / macOS / Linux** 运行。
- **纯 Go 实现**：无需额外依赖，`go build` 即可得到单一可执行文件。

## 安装与运行

### 从源代码构建

```bash
# 1. 安装 Go（≥1.22）
# 2. 克隆代码仓库
$ git clone https://github.com/blueberrycongee/tomato_clock.git
$ cd tomato_clock

# 3. 构建可执行文件
$ go build -o tomato_clock.exe ./cmd/tomato_clock

# 4. 运行
$ ./tomato_clock.exe

#
# 如需打包为跨平台安装包，可使用：
#
# ```bash
# fyne package --release  # 在当前平台生成安装包
# # 或指定目标：fyne package -os windows -icon Icon.png
# ```
```

### 二进制发行版

在仓库的 [Releases](https://github.com/blueberrycongee/tomato_clock/releases) 页面将提供预编译好的可执行文件，下载后直接双击即可运行。

## 项目结构

| 路径 | 说明 |
|------|------|
| `cmd/tomato_clock` | 程序入口 |
| `internal/audio`   | 播放提示音逻辑 |
| `internal/logic`   | 计时器实现 |
| `internal/model`   | 本地数据存储逻辑 |
| `internal/ui`      | Fyne 图形界面 |
| `resources/sounds` | 提示音文件目录 |

## 自定义提示音

将您喜欢的 `alert.mp3` 放到 `resources/sounds/` 目录并重启应用即可生效。

## 开发计划

- [ ] 导出 CSV / Markdown 统计报表  
- [ ] 深色 / 浅色主题自适应  
- [ ] 可编辑快捷键  
- [ ] 系统托盘图标

欢迎提交 Issue 与 PR 共同完善！

## 许可证

本项目基于 **MIT License**，详见 [LICENSE](LICENSE)。 

## 🔥 自然语言聊天助手（LangChain）

现在，你可以通过中文对话来记录番茄钟活动：

* **推荐方式（GUI）**：直接在应用顶栏的“向 DeepSeek 提问…” 输入框输入自然语言指令，助手会即时更新专注记录并弹窗反馈。
* **命令行方式（可选）**：启动下文的 Python Agent，可在终端与助手对话。

### 安装依赖

```bash
# 进入仓库根目录
python -m venv .venv && source .venv/bin/activate  # Windows 使用 .venv\Scripts\activate
pip install -r requirements.txt

# 设置 OpenAI API Key（以 PowerShell 为例）
$Env:OPENAI_API_KEY = "sk-..."
```

### 启动助手

```bash
python -m agent.agent
```

### 对话示例

```text
你: 我下午大概两点半的时候看了半个小时的书
助手: 记录成功，计时 ID = 42
```

助手会自动解析时间表达，转换为北京时区，并把记录写入 `.tomato_clock.json`。随后在番茄钟 GUI 中即可看到同步后的统计变化。 

### 使用 DeepSeek 模型

若你拥有 [DeepSeek](https://deepseek.com/) 的 API Key，可按如下方式切换：

```bash
# 设置 DeepSeek 相关环境变量
export OPENAI_API_KEY="sk-..."       # DeepSeek 提供的 Key
export OPENAI_API_BASE="https://api.deepseek.com"
export OPENAI_MODEL="deepseek-chat"  # 可选，自定义模型名

python -m agent.agent
``` 