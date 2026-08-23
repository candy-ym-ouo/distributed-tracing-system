# Distributed Tracing System 打包说明

这是一个仅依赖 Go 标准库的分布式追踪服务，提供 Span 采样、HTTP Trace ID 透传、批量收集、文件持久化、Trace 查询和统计 API，并附带零构建 Web 界面。

## 本地验证

```bash
go test ./...
go vet ./...
go build ./...
go run ./cmd/server -addr :8080 -file data/spans.jsonl -demo
```

## Docker 打包

```bash
./build_benzhi_docker.sh distributed-tracing-system linux/amd64
./build_benzhi_docker.sh distributed-tracing-system linux/arm64
docker run --rm -it distributed-tracing-system:latest
```

评测 Dockerfile 使用官方多架构 `golang:1.23` 镜像，保留完整 Go 工具链，并在镜像构建阶段执行 `go mod download` 与 `go build ./...`。
