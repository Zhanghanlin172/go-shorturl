# Go ShortURL - 高并发短链接服务

[![Go](https://img.shields.io/badge/Go-1.22-blue)](https://go.dev)

用 Go 语言实现的高并发 URL 短链接服务，内存存储 + RESTful API + 浏览器管理面板。

## 特性

- **高并发安全**：`sync.RWMutex` 读写锁，读操作无互斥
- **短码生成**：MD5 + base62 + 原子计数器，碰撞概率极低
- **301 重定向**：永久重定向，SEO 友好
- **访问统计**：原子计数记录每次跳转
- **RESTful API**：`POST /api/shorten` `GET /api/stats/{code}` `GET /api/health`
- **内置管理面板**：浏览器直接使用的短链接生成器页面

## 快速开始

```bash
git clone https://github.com/Zhanghanlin172/go-shorturl.git
cd go-shorturl
go build -o shorturl ./cmd/
./shorturl
```

打开 `http://localhost:8080`，粘贴长网址即可生成短链接。

## API

```bash
# 生成短链接
curl -X POST http://localhost:8080/api/shorten \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://github.com/Zhanghanlin172/mini-redis"}'

# 查看统计
curl http://localhost:8080/api/stats/3TUaElb

# 健康检查
curl http://localhost:8080/api/health
```

## 项目结构

```
go-shorturl/
├── cmd/main.go                    # 入口：HTTP路由 + 优雅关闭
├── internal/
│   ├── handler/handler.go         # 请求处理：Shorten/Redirect/Stats
│   ├── store/store.go             # 内存存储：RWMutex + map
│   └── generator/generator.go     # 短码生成：MD5 + base62
└── go.mod
```

## 技术栈

- Go 1.22
- `net/http` 标准库（零外部依赖）
- `sync.RWMutex` 并发控制
- `sync/atomic` 原子计数器
