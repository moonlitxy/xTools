# xTools - Golang Zap 日志封装库

[![Go Reference](https://pkg.go.dev/badge/github.com/yourusername/xTools/logx.svg)](https://pkg.go.dev/github.com/yourusername/xTools/logx)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/xTools)](https://goreportcard.com/report/github.com/yourusername/xTools)

`xTools/logx` 是一个基于 Uber [zap](https://github.com/uber-go/zap) 的企业级日志封装库，支持：

- JSON 格式日志输出，适合 ELK / Filebeat 采集
- 日志按天分目录存储，按文件名动态生成日志
- 日志保留天数控制及自动清理
- 支持日志级别：Debug / Info / Warn / Error
- 精确定位调用文件和行号 (`caller`)
- 可选强制刷新缓冲区（实时刷盘）

---

## 安装

在你的项目中引入：

```bash
go get github.com/monnlitxy/xTools/logx
