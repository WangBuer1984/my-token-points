# Cobra、Viper、go-homedir 完全使用教程

> 构建专业命令行应用的三剑客

## 📚 目录

1. [概述](#概述)
2. [Cobra - 强大的 CLI 框架](#cobra---强大的-cli-框架)
3. [Viper - 配置管理利器](#viper---配置管理利器)
4. [go-homedir - 跨平台主目录获取](#go-homedir---跨平台主目录获取)
5. [三者协同实战](#三者协同实战)
6. [完整项目示例](#完整项目示例)
7. [最佳实践](#最佳实践)
8. [常见问题](#常见问题)

---

## 概述

### 这三个库是什么？

| 库 | 作用 | 使用场景 | 著名项目 |
|---|---|---|---|
| **Cobra** | CLI 框架 | 构建命令行应用 | Kubernetes, Hugo, Docker |
| **Viper** | 配置管理 | 读取配置文件、环境变量 | 几乎所有 Cobra 项目 |
| **go-homedir** | 获取用户主目录 | 跨平台路径处理 | 许多需要访问 $HOME 的工具 |

### 为什么一起使用？

```
┌─────────────────────────────────────────┐
│  用户输入命令                            │
│  $ myapp server --config ~/.myapp.yaml │
└──────────────┬──────────────────────────┘
               ↓
    ┌──────────────────────┐
    │   Cobra 解析命令     │
    │   server + flags     │
    └──────────┬───────────┘
               ↓
    ┌──────────────────────┐
    │  go-homedir 解析路径 │
    │  ~/.myapp.yaml       │
    │  → /Users/rick/...   │
    └──────────┬───────────┘
               ↓
    ┌──────────────────────┐
    │  Viper 读取配置      │
    │  - 配置文件          │
    │  - 环境变量          │
    │  - 命令行参数        │
    └──────────────────────┘
```

---

## Cobra - 强大的 CLI 框架

### 1. 简介

**Cobra** 是 Go 语言最流行的 CLI（命令行界面）框架，提供：
- ✅ 简单的子命令结构
- ✅ 全局和局部参数（flags）
- ✅ 智能的帮助信息生成
- ✅ 自动完成脚本生成（bash/zsh）
- ✅ 丰富的文档生成（Markdown/ReStructuredText）

### 2. 安装

```bash
go get -u github.com/spf13/cobra@latest
```

### 3. 核心概念

#### 3.1 命令（Command）

命令是 CLI 应用的基本单元：

```go
&cobra.Command{
    Use:   "serve",           // 命令名称
    Short: "启动服务器",      // 简短描述
    Long:  "启动 HTTP 服务器，监听指定端口", // 详细描述
    Run: func(cmd *cobra.Command, args []string) {
        // 执行逻辑
        fmt.Println("服务器启动中...")
    },
}
```

#### 3.2 命令树结构

```
myapp (rootCmd)
├── serve (子命令)
│   ├── --port (参数)
│   └── --host (参数)
├── version (子命令)
└── config (子命令)
    ├── set (子命令的子命令)
    └── get (子命令的子命令)
```

**对应的命令行：**
```bash
myapp serve --port 8080 --host 0.0.0.0
myapp version
myapp config set key value
myapp config get key
```

#### 3.3 参数类型（Flags）

| 类型 | 说明 | 作用域 | 示例 |
|---|---|---|---|
| **Persistent Flags** | 持久参数 | 当前命令及所有子命令 | `--verbose` |
| **Local Flags** | 本地参数 | 仅当前命令 | `serve --port` |

### 4. 基础使用示例

#### 示例 1：最简单的 CLI

```go
package main

import (
    "fmt"
    "github.com/spf13/cobra"
    "os"
)

func main() {
    var rootCmd = &cobra.Command{
        Use:   "hello",
        Short: "一个简单的问候程序",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("Hello, Cobra!")
        },
    }

    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
```

**运行：**
```bash
$ go run main.go
Hello, Cobra!

$ go run main.go --help
一个简单的问候程序

Usage:
  hello [flags]

Flags:
  -h, --help   help for hello
```

#### 示例 2：带参数的命令

```go
package main

import (
    "fmt"
    "github.com/spf13/cobra"
)

func main() {
    var name string
    var age int

    var rootCmd = &cobra.Command{
        Use:   "greet",
        Short: "问候用户",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Printf("你好 %s，你今年 %d 岁了！\n", name, age)
        },
    }

    // 添加参数
    rootCmd.Flags().StringVarP(&name, "name", "n", "朋友", "你的名字")
    rootCmd.Flags().IntVarP(&age, "age", "a", 18, "你的年龄")

    rootCmd.Execute()
}
```

**运行：**
```bash
$ go run main.go -n Rick -a 25
你好 Rick，你今年 25 岁了！

$ go run main.go --name Alice
你好 Alice，你今年 18 岁了！
```

#### 示例 3：多级子命令

```go
package main

import (
    "fmt"
    "github.com/spf13/cobra"
)

func main() {
    // 根命令
    var rootCmd = &cobra.Command{
        Use:   "app",
        Short: "应用管理工具",
    }

    // 子命令：server
    var serverCmd = &cobra.Command{
        Use:   "server",
        Short: "服务器管理",
    }

    // server 的子命令：start
    var startCmd = &cobra.Command{
        Use:   "start",
        Short: "启动服务器",
        Run: func(cmd *cobra.Command, args []string) {
            port, _ := cmd.Flags().GetInt("port")
            fmt.Printf("服务器启动在端口 %d\n", port)
        },
    }
    startCmd.Flags().Int("port", 8080, "监听端口")

    // server 的子命令：stop
    var stopCmd = &cobra.Command{
        Use:   "stop",
        Short: "停止服务器",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("服务器已停止")
        },
    }

    // 构建命令树
    serverCmd.AddCommand(startCmd)
    serverCmd.AddCommand(stopCmd)
    rootCmd.AddCommand(serverCmd)

    rootCmd.Execute()
}
```

**运行：**
```bash
$ go run main.go server start --port 9000
服务器启动在端口 9000

$ go run main.go server stop
服务器已停止

$ go run main.go server --help
服务器管理

Usage:
  app server [command]

Available Commands:
  start       启动服务器
  stop        停止服务器
```

### 5. Flags 深入详解

#### 5.1 不同类型的 Flag

```go
// String 类型
cmd.Flags().String("name", "default", "描述")
cmd.Flags().StringP("name", "n", "default", "描述")  // 带短名称
cmd.Flags().StringVar(&variable, "name", "default", "描述")
cmd.Flags().StringVarP(&variable, "name", "n", "default", "描述")

// Int 类型
cmd.Flags().Int("count", 0, "描述")
cmd.Flags().IntVarP(&count, "count", "c", 0, "描述")

// Bool 类型
cmd.Flags().Bool("verbose", false, "描述")
cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "描述")

// StringSlice 类型（多个值）
cmd.Flags().StringSlice("tags", []string{}, "标签列表")
```

#### 5.2 必需参数

```go
var name string
cmd.Flags().StringVarP(&name, "name", "n", "", "用户名（必需）")
cmd.MarkFlagRequired("name")  // 标记为必需
```

#### 5.3 Persistent vs Local Flags

```go
// Persistent Flag（对所有子命令有效）
rootCmd.PersistentFlags().BoolP("verbose", "v", false, "详细输出")

// Local Flag（仅对当前命令有效）
serveCmd.Flags().Int("port", 8080, "端口号")
```

**效果：**
```bash
# verbose 对所有命令有效
$ app --verbose server start
$ app config --verbose set key value

# port 只对 serve 有效
$ app serve --port 9000  ✅
$ app config --port 9000  ❌ (错误：未知参数)
```

#### 5.4 Flag 分组

```go
// 创建一组互斥的 flags
cmd.Flags().String("json", "", "JSON 格式输出")
cmd.Flags().String("yaml", "", "YAML 格式输出")
cmd.MarkFlagsMutuallyExclusive("json", "yaml")  // 只能选一个
```

### 6. 命令生命周期钩子

Cobra 提供了多个钩子函数，按执行顺序：

```go
var cmd = &cobra.Command{
    Use: "example",
    
    // 1. 最早执行（包括子命令）
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        fmt.Println("PersistentPreRun")
    },
    
    // 2. 在 Run 之前执行
    PreRun: func(cmd *cobra.Command, args []string) {
        fmt.Println("PreRun")
    },
    
    // 3. 主要逻辑
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Run")
    },
    
    // 4. 在 Run 之后执行
    PostRun: func(cmd *cobra.Command, args []string) {
        fmt.Println("PostRun")
    },
    
    // 5. 最后执行（包括子命令）
    PersistentPostRun: func(cmd *cobra.Command, args []string) {
        fmt.Println("PersistentPostRun")
    },
}
```

**执行顺序：**
```
PersistentPreRun → PreRun → Run → PostRun → PersistentPostRun
```

**使用场景：**
- `PersistentPreRun`: 初始化日志、数据库连接
- `PreRun`: 验证参数、加载配置
- `Run`: 核心业务逻辑
- `PostRun`: 清理资源
- `PersistentPostRun`: 全局清理、统计

### 7. 参数验证

```go
var cmd = &cobra.Command{
    Use:  "print [string]",
    Args: cobra.ExactArgs(1),  // 精确 1 个参数
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("参数:", args[0])
    },
}
```

**常用验证器：**

| 验证器 | 说明 | 示例 |
|---|---|---|
| `NoArgs` | 不接受任何参数 | `app version` |
| `ArbitraryArgs` | 任意数量参数 | `app print a b c` |
| `OnlyValidArgs` | 只接受预定义的参数 | 需配合 `ValidArgs` |
| `MinimumNArgs(n)` | 至少 n 个参数 | `app copy file1 file2 ...` |
| `MaximumNArgs(n)` | 最多 n 个参数 | - |
| `ExactArgs(n)` | 精确 n 个参数 | `app rename old new` |
| `RangeArgs(min, max)` | 参数数量范围 | - |

**自定义验证：**
```go
var cmd = &cobra.Command{
    Use: "add [numbers...]",
    Args: func(cmd *cobra.Command, args []string) error {
        if len(args) < 2 {
            return fmt.Errorf("至少需要 2 个数字")
        }
        for _, arg := range args {
            if _, err := strconv.Atoi(arg); err != nil {
                return fmt.Errorf("'%s' 不是有效数字", arg)
            }
        }
        return nil
    },
    Run: func(cmd *cobra.Command, args []string) {
        sum := 0
        for _, arg := range args {
            num, _ := strconv.Atoi(arg)
            sum += num
        }
        fmt.Printf("总和: %d\n", sum)
    },
}
```

### 8. 自动生成文档

#### 8.1 生成 Markdown 文档

```go
import "github.com/spf13/cobra/doc"

func main() {
    rootCmd := &cobra.Command{Use: "myapp"}
    // ... 添加子命令 ...
    
    // 生成 Markdown 文档
    err := doc.GenMarkdownTree(rootCmd, "./docs")
    if err != nil {
        log.Fatal(err)
    }
}
```

#### 8.2 生成 Man Pages

```go
err := doc.GenManTree(rootCmd, &doc.GenManHeader{
    Title:   "MYAPP",
    Section: "1",
}, "/usr/local/share/man/man1/")
```

### 9. Shell 自动完成

```go
// 生成 bash 自动完成脚本
rootCmd.GenBashCompletionFile("myapp_completion.bash")

// 生成 zsh 自动完成脚本
rootCmd.GenZshCompletionFile("myapp_completion.zsh")
```

**安装自动完成（bash）：**
```bash
# 生成脚本
./myapp completion bash > myapp_completion.bash

# 安装
sudo mv myapp_completion.bash /etc/bash_completion.d/

# 或者临时使用
source myapp_completion.bash
```

---

## Viper - 配置管理利器

### 1. 简介

**Viper** 是 Go 语言功能最全的配置解决方案，支持：
- ✅ 读取多种格式：JSON, TOML, YAML, HCL, envfile, Java properties
- ✅ 实时监控配置文件变化
- ✅ 从多种来源读取配置：
  - 配置文件
  - 环境变量
  - 命令行参数（与 Cobra 集成）
  - 远程配置系统（etcd, Consul）
- ✅ 支持默认值
- ✅ 配置优先级管理

### 2. 安装

```bash
go get github.com/spf13/viper
```

### 3. 核心概念

#### 3.1 配置优先级（从高到低）

```
1. 显式调用 viper.Set()
2. 命令行参数（flags）
3. 环境变量
4. 配置文件
5. 远程配置（如 etcd）
6. 默认值
```

#### 3.2 配置键的访问

Viper 使用 `.` 作为键的分隔符：

```yaml
# config.yaml
database:
  host: localhost
  port: 5432
  credentials:
    username: admin
    password: secret
```

```go
// 访问方式
host := viper.GetString("database.host")              // "localhost"
port := viper.GetInt("database.port")                 // 5432
user := viper.GetString("database.credentials.username") // "admin"
```

### 4. 基础使用示例

#### 示例 1：读取配置文件

**config.yaml:**
```yaml
app:
  name: MyApp
  version: 1.0.0
server:
  host: 0.0.0.0
  port: 8080
database:
  host: localhost
  port: 5432
  username: admin
  password: secret123
```

**main.go:**
```go
package main

import (
    "fmt"
    "github.com/spf13/viper"
)

func main() {
    // 设置配置文件名（不带扩展名）
    viper.SetConfigName("config")
    // 设置配置文件类型
    viper.SetConfigType("yaml")
    // 添加配置文件搜索路径
    viper.AddConfigPath(".")
    viper.AddConfigPath("./config")
    viper.AddConfigPath("/etc/myapp/")

    // 读取配置文件
    if err := viper.ReadInConfig(); err != nil {
        panic(fmt.Errorf("配置文件读取失败: %s", err))
    }

    // 读取配置
    appName := viper.GetString("app.name")
    serverPort := viper.GetInt("server.port")
    dbHost := viper.GetString("database.host")

    fmt.Printf("应用: %s\n", appName)
    fmt.Printf("端口: %d\n", serverPort)
    fmt.Printf("数据库: %s\n", dbHost)
}
```

**输出：**
```
应用: MyApp
端口: 8080
数据库: localhost
```

#### 示例 2：设置默认值

```go
package main

import (
    "fmt"
    "github.com/spf13/viper"
)

func main() {
    // 设置默认值
    viper.SetDefault("server.host", "0.0.0.0")
    viper.SetDefault("server.port", 8080)
    viper.SetDefault("log.level", "info")
    viper.SetDefault("log.format", "json")

    // 即使没有配置文件，也能获取值
    fmt.Println("Host:", viper.GetString("server.host"))
    fmt.Println("Port:", viper.GetInt("server.port"))
    fmt.Println("Log Level:", viper.GetString("log.level"))
}
```

#### 示例 3：环境变量支持

```go
package main

import (
    "fmt"
    "github.com/spf13/viper"
)

func main() {
    // 自动读取环境变量
    viper.AutomaticEnv()
    
    // 设置环境变量前缀（只读取 MYAPP_ 开头的）
    viper.SetEnvPrefix("MYAPP")
    
    // 绑定特定环境变量
    viper.BindEnv("database.password", "DB_PASSWORD")
    
    // 读取值（优先从环境变量）
    dbPass := viper.GetString("database.password")
    fmt.Println("密码:", dbPass)
}
```

**运行：**
```bash
# 设置环境变量
export MYAPP_DATABASE_PASSWORD="secret123"
# 或
export DB_PASSWORD="secret123"

# 运行程序
go run main.go
# 输出: 密码: secret123
```

#### 示例 4：指定配置文件路径

```go
package main

import (
    "fmt"
    "github.com/spf13/viper"
)

func main() {
    // 方式1：直接指定完整路径
    viper.SetConfigFile("./config/prod.yaml")
    
    // 方式2：指定名称和路径
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("/etc/myapp/")
    viper.AddConfigPath("$HOME/.myapp")
    viper.AddConfigPath(".")
    
    if err := viper.ReadInConfig(); err != nil {
        panic(err)
    }
    
    fmt.Println("使用配置文件:", viper.ConfigFileUsed())
}
```

### 5. 高级功能

#### 5.1 监控配置文件变化

```go
package main

import (
    "fmt"
    "github.com/fsnotify/fsnotify"
    "github.com/spf13/viper"
    "time"
)

func main() {
    viper.SetConfigFile("./config.yaml")
    viper.ReadInConfig()

    // 监控配置文件变化
    viper.WatchConfig()
    viper.OnConfigChange(func(e fsnotify.Event) {
        fmt.Println("配置文件已修改:", e.Name)
        // 重新读取配置
        newPort := viper.GetInt("server.port")
        fmt.Println("新端口:", newPort)
    })

    // 保持程序运行
    for {
        fmt.Println("当前端口:", viper.GetInt("server.port"))
        time.Sleep(5 * time.Second)
    }
}
```

**测试：**
1. 运行程序
2. 修改 `config.yaml` 中的 `server.port`
3. 保存文件
4. 程序自动检测并输出新值

#### 5.2 将配置映射到结构体

```go
package main

import (
    "fmt"
    "github.com/spf13/viper"
)

// 定义配置结构体
type Config struct {
    App struct {
        Name    string `mapstructure:"name"`
        Version string `mapstructure:"version"`
    } `mapstructure:"app"`
    
    Server struct {
        Host string `mapstructure:"host"`
        Port int    `mapstructure:"port"`
    } `mapstructure:"server"`
    
    Database struct {
        Host     string `mapstructure:"host"`
        Port     int    `mapstructure:"port"`
        Username string `mapstructure:"username"`
        Password string `mapstructure:"password"`
    } `mapstructure:"database"`
}

func main() {
    viper.SetConfigFile("./config.yaml")
    viper.ReadInConfig()

    var config Config
    // 将配置解析到结构体
    if err := viper.Unmarshal(&config); err != nil {
        panic(err)
    }

    fmt.Printf("应用: %s v%s\n", config.App.Name, config.App.Version)
    fmt.Printf("服务器: %s:%d\n", config.Server.Host, config.Server.Port)
    fmt.Printf("数据库: %s@%s:%d\n", 
        config.Database.Username, 
        config.Database.Host, 
        config.Database.Port)
}
```

**输出：**
```
应用: MyApp v1.0.0
服务器: 0.0.0.0:8080
数据库: admin@localhost:5432
```

#### 5.3 读取嵌套配置

```yaml
# config.yaml
features:
  authentication:
    enabled: true
    providers:
      - oauth
      - ldap
      - local
  notifications:
    email:
      enabled: true
      smtp_host: smtp.gmail.com
    sms:
      enabled: false
```

```go
// 读取嵌套的布尔值
authEnabled := viper.GetBool("features.authentication.enabled")

// 读取数组
providers := viper.GetStringSlice("features.authentication.providers")
fmt.Println(providers) // [oauth ldap local]

// 读取子配置
emailConfig := viper.Sub("features.notifications.email")
if emailConfig != nil {
    host := emailConfig.GetString("smtp_host")
    fmt.Println("SMTP:", host)
}
```

#### 5.4 设置和保存配置

```go
package main

import (
    "fmt"
    "github.com/spf13/viper"
)

func main() {
    viper.SetConfigFile("./config.yaml")
    viper.ReadInConfig()

    // 修改配置值
    viper.Set("server.port", 9000)
    viper.Set("app.version", "2.0.0")

    // 保存到文件
    if err := viper.WriteConfig(); err != nil {
        panic(err)
    }
    
    // 或者另存为
    if err := viper.WriteConfigAs("./config_new.yaml"); err != nil {
        panic(err)
    }

    fmt.Println("配置已保存")
}
```

#### 5.5 安全读取配置（防止 panic）

```go
// 不安全的方式（如果键不存在会返回零值）
port := viper.GetInt("server.port")  // 不存在返回 0

// 安全的方式
if viper.IsSet("server.port") {
    port := viper.GetInt("server.port")
    fmt.Println("端口:", port)
} else {
    fmt.Println("未设置端口")
}

// 使用默认值
port := viper.GetInt("server.port")
if port == 0 {
    port = 8080  // 默认值
}
```

### 6. 数据类型支持

| 方法 | 返回类型 | 示例 |
|---|---|---|
| `Get(key)` | `interface{}` | 任意类型 |
| `GetBool(key)` | `bool` | `true`/`false` |
| `GetFloat64(key)` | `float64` | `3.14` |
| `GetInt(key)` | `int` | `42` |
| `GetInt32(key)` | `int32` | - |
| `GetInt64(key)` | `int64` | - |
| `GetUint(key)` | `uint` | - |
| `GetString(key)` | `string` | `"hello"` |
| `GetStringSlice(key)` | `[]string` | `["a", "b"]` |
| `GetStringMap(key)` | `map[string]interface{}` | 嵌套对象 |
| `GetStringMapString(key)` | `map[string]string` | 字符串映射 |
| `GetTime(key)` | `time.Time` | 时间类型 |
| `GetDuration(key)` | `time.Duration` | `5s`, `2h` |

**示例：**
```yaml
# config.yaml
timeout: 30s
retry_count: 3
enabled: true
tags: [backend, api, production]
metadata:
  author: Rick
  date: 2025-11-09
```

```go
timeout := viper.GetDuration("timeout")      // 30 * time.Second
retries := viper.GetInt("retry_count")       // 3
enabled := viper.GetBool("enabled")          // true
tags := viper.GetStringSlice("tags")         // []string{"backend", "api", "production"}
metadata := viper.GetStringMapString("metadata") // map[string]string{"author":"Rick",...}
```

### 7. 环境变量高级用法

#### 7.1 自动环境变量映射

```go
viper.SetEnvPrefix("MYAPP")  // 前缀
viper.AutomaticEnv()         // 自动绑定

// 键名转换规则：
// database.host → MYAPP_DATABASE_HOST
// server.port   → MYAPP_SERVER_PORT
```

#### 7.2 自定义键名转换

```go
import "strings"

viper.SetEnvPrefix("MYAPP")
viper.AutomaticEnv()

// 将 . 替换为 _
replacer := strings.NewReplacer(".", "_")
viper.SetEnvKeyReplacer(replacer)

// database.host → MYAPP_DATABASE_HOST
host := viper.GetString("database.host")
```

#### 7.3 绑定特定环境变量

```go
// 绑定单个环境变量
viper.BindEnv("db.password", "DATABASE_PASSWORD")

// 绑定多个可能的环境变量（按顺序查找）
viper.BindEnv("db.host", "DATABASE_HOST", "DB_HOST")
```

**示例：**
```bash
export DATABASE_PASSWORD="secret123"
export DB_HOST="localhost"
```

```go
password := viper.GetString("db.password")  // 从 DATABASE_PASSWORD
host := viper.GetString("db.host")          // 从 DATABASE_HOST 或 DB_HOST
```

### 8. 远程配置支持

Viper 支持从远程配置中心读取配置（如 etcd, Consul）：

```go
import _ "github.com/spf13/viper/remote"

viper.AddRemoteProvider("etcd", "http://127.0.0.1:4001", "/config/myapp.json")
viper.SetConfigType("json")

if err := viper.ReadRemoteConfig(); err != nil {
    panic(err)
}

// 监控远程配置变化
go func(){
    for {
        time.Sleep(time.Second * 5)
        viper.WatchRemoteConfig()
    }
}()
```

---

## go-homedir - 跨平台主目录获取

### 1. 简介

**go-homedir** 是一个轻量级库，用于跨平台地获取用户主目录路径。

**为什么需要它？**

| 平台 | 主目录环境变量 | 示例路径 |
|---|---|---|
| Linux/Mac | `$HOME` | `/home/rick` |
| Windows | `%USERPROFILE%` | `C:\Users\Rick` |
| 特殊情况 | 多种变量 | 需要兼容性处理 |

**go-homedir** 统一处理了这些差异。

### 2. 安装

```bash
go get github.com/mitchellh/go-homedir
```

### 3. 基础使用

#### 示例 1：获取主目录

```go
package main

import (
    "fmt"
    "github.com/mitchellh/go-homedir"
)

func main() {
    // 获取用户主目录
    home, err := homedir.Dir()
    if err != nil {
        panic(err)
    }
    
    fmt.Println("主目录:", home)
    // Linux/Mac: /home/rick
    // Windows: C:\Users\Rick
}
```

#### 示例 2：展开路径中的 ~

```go
package main

import (
    "fmt"
    "github.com/mitchellh/go-homedir"
)

func main() {
    // 展开 ~ 为实际路径
    path := "~/.myapp/config.yaml"
    expandedPath, err := homedir.Expand(path)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("原始路径:", path)
    fmt.Println("展开路径:", expandedPath)
    // 原始路径: ~/.myapp/config.yaml
    // 展开路径: /Users/rick/.myapp/config.yaml
}
```

#### 示例 3：实际应用场景

```go
package main

import (
    "fmt"
    "os"
    "path/filepath"
    "github.com/mitchellh/go-homedir"
)

func main() {
    // 获取主目录
    home, _ := homedir.Dir()
    
    // 构建配置文件路径
    configDir := filepath.Join(home, ".myapp")
    configFile := filepath.Join(configDir, "config.yaml")
    
    // 创建配置目录（如果不存在）
    if err := os.MkdirAll(configDir, 0755); err != nil {
        panic(err)
    }
    
    fmt.Println("配置目录:", configDir)
    fmt.Println("配置文件:", configFile)
    // 配置目录: /Users/rick/.myapp
    // 配置文件: /Users/rick/.myapp/config.yaml
}
```

### 4. 与 Viper 集成

```go
package main

import (
    "github.com/mitchellh/go-homedir"
    "github.com/spf13/viper"
)

func main() {
    // 获取主目录
    home, err := homedir.Dir()
    if err != nil {
        panic(err)
    }
    
    // 添加主目录下的配置路径
    viper.AddConfigPath(home)
    viper.AddConfigPath(home + "/.myapp")
    viper.SetConfigName("config")
    
    viper.ReadInConfig()
}
```

### 5. 缓存机制

go-homedir 会缓存主目录路径以提高性能：

```go
// 第一次调用会读取环境变量
home1, _ := homedir.Dir()

// 后续调用直接返回缓存值
home2, _ := homedir.Dir()

// 如果需要强制重新读取
homedir.Reset()
home3, _ := homedir.Dir()
```

### 6. 错误处理

```go
home, err := homedir.Dir()
if err != nil {
    // 处理错误的情况：
    // 1. 环境变量未设置
    // 2. 无法确定主目录
    fmt.Println("无法获取主目录，使用当前目录")
    home = "."
}
```

---

## 三者协同实战

现在我们将 Cobra、Viper 和 go-homedir 结合起来，构建一个完整的应用。

### 完整示例：文件管理工具

#### 目录结构

```
myapp/
├── cmd/
│   ├── root.go       # 根命令
│   ├── serve.go      # serve 子命令
│   └── config.go     # config 子命令
├── config/
│   └── config.go     # 配置管理
├── main.go
└── config.yaml       # 配置文件
```

#### 1. config/config.go（配置管理）

```go
package config

import (
    "fmt"
    "path/filepath"
    "strings"

    "github.com/mitchellh/go-homedir"
    "github.com/spf13/viper"
)

type Config struct {
    Server struct {
        Host string `mapstructure:"host"`
        Port int    `mapstructure:"port"`
    } `mapstructure:"server"`
    
    Database struct {
        Host     string `mapstructure:"host"`
        Port     int    `mapstructure:"port"`
        Username string `mapstructure:"username"`
        Password string `mapstructure:"password"`
    } `mapstructure:"database"`
    
    Log struct {
        Level  string `mapstructure:"level"`
        Format string `mapstructure:"format"`
    } `mapstructure:"log"`
}

var AppConfig *Config

// 初始化配置
func InitConfig(cfgFile string) error {
    if cfgFile != "" {
        // 使用指定的配置文件
        viper.SetConfigFile(cfgFile)
    } else {
        // 获取主目录
        home, err := homedir.Dir()
        if err != nil {
            return err
        }

        // 搜索配置文件的位置
        viper.AddConfigPath(".")
        viper.AddConfigPath(filepath.Join(home, ".myapp"))
        viper.AddConfigPath("/etc/myapp/")
        viper.SetConfigName("config")
        viper.SetConfigType("yaml")
    }

    // 环境变量支持
    viper.SetEnvPrefix("MYAPP")
    viper.AutomaticEnv()
    replacer := strings.NewReplacer(".", "_")
    viper.SetEnvKeyReplacer(replacer)

    // 设置默认值
    viper.SetDefault("server.host", "0.0.0.0")
    viper.SetDefault("server.port", 8080)
    viper.SetDefault("log.level", "info")
    viper.SetDefault("log.format", "json")

    // 读取配置文件
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ok {
            // 配置文件不存在，使用默认值
            fmt.Println("未找到配置文件，使用默认配置")
        } else {
            return err
        }
    } else {
        fmt.Println("使用配置文件:", viper.ConfigFileUsed())
    }

    // 解析到结构体
    AppConfig = &Config{}
    if err := viper.Unmarshal(AppConfig); err != nil {
        return err
    }

    return nil
}

// 获取配置值
func GetString(key string) string {
    return viper.GetString(key)
}

func GetInt(key string) int {
    return viper.GetInt(key)
}

func GetBool(key string) bool {
    return viper.GetBool(key)
}
```

#### 2. cmd/root.go（根命令）

```go
package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "myapp/config"
)

var cfgFile string
var verbose bool

var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "一个功能完整的应用示例",
    Long: `这是一个集成了 Cobra、Viper 和 go-homedir 的示例应用，
展示了如何构建专业的命令行工具。`,
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

func init() {
    // 在执行命令前初始化配置
    cobra.OnInitialize(initConfig)

    // Persistent flags（对所有子命令有效）
    rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", 
        "配置文件路径 (默认搜索 ./config.yaml 或 ~/.myapp/config.yaml)")
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, 
        "详细输出")
}

func initConfig() {
    if err := config.InitConfig(cfgFile); err != nil {
        fmt.Println("配置初始化失败:", err)
        os.Exit(1)
    }

    if verbose {
        fmt.Println("详细模式已开启")
        fmt.Printf("配置: %+v\n", config.AppConfig)
    }
}
```

#### 3. cmd/serve.go（serve 子命令）

```go
package cmd

import (
    "fmt"
    "net/http"

    "github.com/spf13/cobra"
    "myapp/config"
)

var port int

var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "启动 HTTP 服务器",
    Long:  "启动 HTTP 服务器并监听指定端口",
    Run: func(cmd *cobra.Command, args []string) {
        // 优先使用命令行参数，其次使用配置文件
        if port == 0 {
            port = config.AppConfig.Server.Port
        }
        host := config.AppConfig.Server.Host

        addr := fmt.Sprintf("%s:%d", host, port)
        fmt.Printf("服务器启动在 http://%s\n", addr)

        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
            fmt.Fprintf(w, "Hello from MyApp!\n")
            fmt.Fprintf(w, "Log Level: %s\n", config.AppConfig.Log.Level)
        })

        if err := http.ListenAndServe(addr, nil); err != nil {
            fmt.Println("服务器启动失败:", err)
        }
    },
}

func init() {
    rootCmd.AddCommand(serveCmd)

    // Local flags（仅对 serve 命令有效）
    serveCmd.Flags().IntVarP(&port, "port", "p", 0, 
        "监听端口（覆盖配置文件）")
}
```

#### 4. cmd/config.go（config 子命令）

```go
package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var configCmd = &cobra.Command{
    Use:   "config",
    Short: "配置管理",
}

var configGetCmd = &cobra.Command{
    Use:   "get [key]",
    Short: "获取配置值",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        key := args[0]
        value := viper.Get(key)
        
        if value != nil {
            fmt.Printf("%s = %v\n", key, value)
        } else {
            fmt.Printf("配置项 '%s' 不存在\n", key)
        }
    },
}

var configSetCmd = &cobra.Command{
    Use:   "set [key] [value]",
    Short: "设置配置值",
    Args:  cobra.ExactArgs(2),
    Run: func(cmd *cobra.Command, args []string) {
        key := args[0]
        value := args[1]
        
        viper.Set(key, value)
        
        if err := viper.WriteConfig(); err != nil {
            fmt.Println("保存配置失败:", err)
            return
        }
        
        fmt.Printf("已设置 %s = %s\n", key, value)
    },
}

var configListCmd = &cobra.Command{
    Use:   "list",
    Short: "列出所有配置",
    Run: func(cmd *cobra.Command, args []string) {
        settings := viper.AllSettings()
        fmt.Println("当前配置:")
        for key, value := range settings {
            fmt.Printf("  %s: %v\n", key, value)
        }
    },
}

func init() {
    rootCmd.AddCommand(configCmd)
    configCmd.AddCommand(configGetCmd)
    configCmd.AddCommand(configSetCmd)
    configCmd.AddCommand(configListCmd)
}
```

#### 5. main.go（程序入口）

```go
package main

import "myapp/cmd"

func main() {
    cmd.Execute()
}
```

#### 6. config.yaml（配置文件）

```yaml
server:
  host: 0.0.0.0
  port: 8080

database:
  host: localhost
  port: 5432
  username: admin
  password: secret123

log:
  level: info
  format: json
```

### 使用示例

```bash
# 1. 构建应用
go build -o myapp

# 2. 查看帮助
./myapp --help

# 3. 启动服务器（使用配置文件）
./myapp serve

# 4. 启动服务器（覆盖端口）
./myapp serve --port 9000

# 5. 使用自定义配置文件
./myapp serve -c /path/to/config.yaml

# 6. 详细模式
./myapp serve --verbose

# 7. 环境变量覆盖配置
export MYAPP_SERVER_PORT=7000
./myapp serve

# 8. 配置管理
./myapp config list
./myapp config get server.port
./myapp config set server.port 9000
```

---

## 最佳实践

### 1. 配置文件组织

```
项目根目录/
├── config/
│   ├── default.yaml      # 默认配置
│   ├── development.yaml  # 开发环境
│   ├── production.yaml   # 生产环境
│   └── test.yaml         # 测试环境
```

**动态加载：**
```go
env := os.Getenv("APP_ENV")
if env == "" {
    env = "development"
}

viper.SetConfigName(env)
viper.AddConfigPath("./config")
viper.ReadInConfig()
```

### 2. 敏感信息处理

**不要在配置文件中存储敏感信息！**

```yaml
# ❌ 错误：明文密码
database:
  password: secret123

# ✅ 正确：使用环境变量占位符
database:
  password: ${DB_PASSWORD}
```

```go
// 从环境变量读取
viper.AutomaticEnv()
dbPassword := viper.GetString("database.password")
```

### 3. 配置验证

```go
func ValidateConfig() error {
    required := []string{
        "server.host",
        "server.port",
        "database.host",
    }
    
    for _, key := range required {
        if !viper.IsSet(key) {
            return fmt.Errorf("缺少必需配置: %s", key)
        }
    }
    
    // 值验证
    port := viper.GetInt("server.port")
    if port < 1 || port > 65535 {
        return fmt.Errorf("无效的端口号: %d", port)
    }
    
    return nil
}
```

### 4. 优雅的错误处理

```go
func InitConfig(cfgFile string) error {
    viper.SetConfigFile(cfgFile)
    
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ok {
            // 配置文件不存在，创建默认配置
            return createDefaultConfig(cfgFile)
        }
        return fmt.Errorf("读取配置文件失败: %w", err)
    }
    
    // 验证配置
    if err := ValidateConfig(); err != nil {
        return fmt.Errorf("配置验证失败: %w", err)
    }
    
    return nil
}
```

### 5. 命令组织

对于大型项目，将命令拆分到不同文件：

```
cmd/
├── root.go           # 根命令
├── server/
│   ├── start.go
│   ├── stop.go
│   └── restart.go
├── database/
│   ├── migrate.go
│   └── seed.go
└── user/
    ├── create.go
    ├── delete.go
    └── list.go
```

### 6. 日志集成

```go
var rootCmd = &cobra.Command{
    Use: "myapp",
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        // 根据配置初始化日志
        logLevel := viper.GetString("log.level")
        initLogger(logLevel)
    },
}
```

### 7. 进度显示

```go
import "github.com/schollz/progressbar/v3"

var longRunningCmd = &cobra.Command{
    Use: "process",
    Run: func(cmd *cobra.Command, args []string) {
        bar := progressbar.Default(100)
        for i := 0; i < 100; i++ {
            time.Sleep(10 * time.Millisecond)
            bar.Add(1)
        }
    },
}
```

---

## 常见问题

### Q1：Cobra 和标准库的 flag 包有什么区别？

| 特性 | flag 包 | Cobra |
|---|---|---|
| 子命令支持 | ❌ | ✅ |
| 自动帮助生成 | 基础 | 丰富 |
| 命令别名 | ❌ | ✅ |
| Shell 自动完成 | ❌ | ✅ |
| 学习曲线 | 低 | 中 |

**建议：** 简单脚本用 `flag`，复杂 CLI 用 `Cobra`。

### Q2：Viper 和直接读取配置文件有什么区别？

**直接读取（如 `encoding/json`）：**
```go
file, _ := os.ReadFile("config.json")
var config Config
json.Unmarshal(file, &config)
```

**使用 Viper：**
- ✅ 支持多种格式（JSON, YAML, TOML...）
- ✅ 环境变量自动覆盖
- ✅ 配置热更新
- ✅ 默认值支持
- ✅ 配置优先级管理

### Q3：配置文件应该放在哪里？

**推荐路径（按优先级）：**

1. 命令行指定：`--config /path/to/config.yaml`
2. 当前目录：`./config.yaml`
3. 用户目录：`~/.myapp/config.yaml`
4. 系统目录：`/etc/myapp/config.yaml`

```go
viper.AddConfigPath(".")
viper.AddConfigPath("$HOME/.myapp")
viper.AddConfigPath("/etc/myapp/")
```

### Q4：如何处理配置文件不存在的情况？

```go
if err := viper.ReadInConfig(); err != nil {
    if _, ok := err.(viper.ConfigFileNotFoundError); ok {
        // 方案1：使用默认值
        fmt.Println("使用默认配置")
        
        // 方案2：创建默认配置文件
        if err := createDefaultConfig(); err != nil {
            return err
        }
        
        // 方案3：交互式配置向导
        if err := runConfigWizard(); err != nil {
            return err
        }
    } else {
        return err
    }
}
```

### Q5：如何测试使用了 Cobra 的应用？

```go
func TestRootCommand(t *testing.T) {
    // 重置 rootCmd（避免测试间干扰）
    rootCmd.SetArgs([]string{"serve", "--port", "9000"})
    
    // 捕获输出
    output := new(bytes.Buffer)
    rootCmd.SetOut(output)
    rootCmd.SetErr(output)
    
    // 执行命令
    if err := rootCmd.Execute(); err != nil {
        t.Fatal(err)
    }
    
    // 验证输出
    if !strings.Contains(output.String(), "9000") {
        t.Error("输出中未包含端口号")
    }
}
```

### Q6：Viper 的配置优先级具体是怎样的？

```go
viper.SetDefault("key", "default")           // 优先级: 6（最低）
viper.ReadInConfig()                          // 优先级: 5
viper.ReadRemoteConfig()                      // 优先级: 4
os.Setenv("MYAPP_KEY", "env_value")          // 优先级: 3
viper.BindPFlag("key", cmd.Flags().Lookup("key")) // 优先级: 2
viper.Set("key", "explicit")                  // 优先级: 1（最高）
```

**实际效果：**
```go
viper.SetDefault("port", 8080)    // 默认 8080
// config.yaml: port: 9000        // 配置文件 9000
// 环境变量: MYAPP_PORT=7000      // 环境变量 7000
viper.Set("port", 6000)           // 显式设置 6000

fmt.Println(viper.GetInt("port")) // 输出: 6000
```

### Q7：如何在不同环境使用不同配置？

```go
// 方式1：通过环境变量选择配置文件
env := os.Getenv("APP_ENV")
if env == "" {
    env = "development"
}
viper.SetConfigName(fmt.Sprintf("config.%s", env))

// 方式2：使用配置继承
viper.SetConfigName("config")        // 基础配置
viper.ReadInConfig()

if env := os.Getenv("APP_ENV"); env != "" {
    viper.SetConfigName(fmt.Sprintf("config.%s", env))
    viper.MergeInConfig()  // 合并环境特定配置
}
```

---

## 进阶主题

### 1. 自定义 Cobra 模板

```go
rootCmd.SetUsageTemplate(`自定义使用说明:
命令: {{.Name}}
描述: {{.Short}}

用法:
  {{.UseLine}}

可用命令:{{range .Commands}}{{if .IsAvailableCommand}}
  {{.Name}}: {{.Short}}{{end}}{{end}}
`)
```

### 2. Viper 插件开发

```go
// 实现自定义配置源
type CustomConfigProvider struct{}

func (c *CustomConfigProvider) Get(key string) interface{} {
    // 从自定义源读取配置
    return nil
}

// 注册到 Viper
viper.AddRemoteProvider("custom", "endpoint", "path")
```

### 3. 命令别名和隐藏命令

```go
var serveCmd = &cobra.Command{
    Use:     "serve",
    Aliases: []string{"server", "start", "run"}, // 别名
    Hidden:  false,  // 设为 true 在帮助中隐藏
}
```

### 4. 动态命令注册

```go
func RegisterPlugins() {
    plugins := []string{"plugin1", "plugin2"}
    
    for _, plugin := range plugins {
        cmd := &cobra.Command{
            Use: plugin,
            Run: func(cmd *cobra.Command, args []string) {
                fmt.Printf("执行插件: %s\n", plugin)
            },
        }
        rootCmd.AddCommand(cmd)
    }
}
```

---

## 总结

### 学习路线图

```
1. 基础阶段
   ├─ Cobra: 创建基本命令
   ├─ Viper: 读取配置文件
   └─ go-homedir: 获取主目录

2. 进阶阶段
   ├─ Cobra: 子命令、参数、验证
   ├─ Viper: 环境变量、结构体映射
   └─ 集成: 三者协同使用

3. 高级阶段
   ├─ Cobra: 自定义模板、插件系统
   ├─ Viper: 远程配置、热更新
   └─ 生产: 日志、监控、测试
```

### 核心要点

| 库 | 核心功能 | 关键方法 |
|---|---|---|
| **Cobra** | CLI 框架 | `Command`, `Flags`, `Execute()` |
| **Viper** | 配置管理 | `ReadInConfig()`, `Get()`, `Set()` |
| **go-homedir** | 主目录 | `Dir()`, `Expand()` |

### 推荐资源

1. **官方文档：**
   - Cobra: https://github.com/spf13/cobra
   - Viper: https://github.com/spf13/viper
   - go-homedir: https://github.com/mitchellh/go-homedir

2. **示例项目：**
   - Kubernetes CLI: https://github.com/kubernetes/kubectl
   - Hugo: https://github.com/gohugoio/hugo

3. **相关工具：**
   - cobra-cli: 生成 Cobra 项目脚手架
   - viper-gen: 生成 Viper 配置代码

---

## 附录：完整代码模板

### 项目初始化脚本

```bash
#!/bin/bash

# 创建项目结构
mkdir -p myapp/{cmd,config,internal}
cd myapp

# 初始化 Go 模块
go mod init github.com/yourusername/myapp

# 安装依赖
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest
go get github.com/mitchellh/go-homedir@latest

# 创建基础文件
touch main.go
touch cmd/root.go
touch config/config.go
touch config.yaml

echo "项目初始化完成！"
```

### Makefile

```makefile
.PHONY: build run test clean install

# 构建
build:
	go build -o bin/myapp main.go

# 运行
run:
	go run main.go

# 测试
test:
	go test -v ./...

# 清理
clean:
	rm -rf bin/

# 安装
install:
	go install

# 生成文档
docs:
	go run main.go docs --dir ./docs

# 构建多平台
build-all:
	GOOS=linux GOARCH=amd64 go build -o bin/myapp-linux-amd64
	GOOS=darwin GOARCH=amd64 go build -o bin/myapp-darwin-amd64
	GOOS=windows GOARCH=amd64 go build -o bin/myapp-windows-amd64.exe
```

---

希望这份教程能帮助你掌握 Cobra、Viper 和 go-homedir！如有疑问，欢迎随时提问。🚀

