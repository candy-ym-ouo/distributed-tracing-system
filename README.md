# Distributed Tracing System

标准库实现的轻量分布式追踪服务，包含 Span 生成与采样、HTTP Trace ID 透传、异步批量收集、内存/JSONL 存储、按 Trace 聚合查询、统计 API 和零构建前端。

## 运行

```bash
go run ./cmd/server -addr :8080 -file data/spans.jsonl -demo
```

打开 <http://localhost:8080/>，或调用 `POST /api/v1/spans` 上报 `{"spans":[...]}`。查询接口包括 `/api/v1/traces/{traceId}`、`/api/v1/traces`、`/api/v1/stats/services`、`/api/v1/stats/operations`、`/api/v1/stats/topology` 和 `/healthz`。
