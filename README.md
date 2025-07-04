# 🍅 Tomato Clock

一个用 **Go** 和 **[Fyne](https://fyne.io/)** 构建的跨平台番茄钟应用程序，助你专注工作与学习。

## 功能特性

- **番茄钟**：支持正计时 *(Count-up)* 与倒计时 *(Count-down)* 两种模式。
- **任务管理**：可为每段专注时间关联任务，并自动持久化到本地 JSON 文件。
- **数据统计**：统计最近 24 小时的专注时长，并按任务标签聚合。
- **声音提醒**：计时结束时播放 `resources/sounds/alert.mp3`，可在应用内静音。
- **离线存储**：所有数据均保存到用户目录下的 `.tomato_clock.json` 文件，无需数据库。
- **跨平台**：得益于 Fyne 框架，可在 **Windows / macOS / Linux** 运行。
- **纯 Go 实现**：无需额外依赖，`go build` 即可得到单一可执行文件。

## 预览

> 截图占位。如有需要请自行补充应用运行界面。

## 安装与运行

### 从源代码构建

```bash
# 1. 安装 Go（≥1.20）
# 2. 克隆代码仓库
$ git clone https://github.com/yourusername/tomato_clock.git
$ cd tomato_clock

# 3. 构建可执行文件
$ go build -o tomato_clock.exe ./cmd/tomato_clock

# 4. 运行
$ ./tomato_clock.exe
```

### 二进制发行版

在仓库的 [Releases](https://github.com/yourusername/tomato_clock/releases) 页面将提供预编译好的可执行文件，下载后直接双击即可运行。

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