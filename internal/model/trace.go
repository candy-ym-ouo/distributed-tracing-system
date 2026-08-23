package model

type SpanNode struct {
	Span     *Span       `json:"span,omitempty"`
	Children []*SpanNode `json:"children,omitempty"`
	Depth    int         `json:"depth"`
	Orphan   bool        `json:"orphan,omitempty"`
}

type TraceTree struct {
	TraceID     string    `json:"trace_id"`
	Root        *SpanNode `json:"root"`
	DurationUs  int64     `json:"duration_us"`
	SpanCount   int       `json:"span_count"`
	ServiceHops int       `json:"service_hops"`
}

type TraceSummary struct {
	TraceID       string `json:"trace_id"`
	RootService   string `json:"root_service"`
	RootOperation string `json:"root_operation"`
	StartTime     int64  `json:"start_time"`
	DurationUs    int64  `json:"duration_us"`
	SpanCount     int    `json:"span_count"`
	Status        Status `json:"status"`
}

type Query struct {
	StartTime int64
	EndTime   int64
	Service   string
	Operation string
	Status    Status
	Page      int
	PageSize  int
}

type TracePage struct {
	Items    []TraceSummary `json:"items"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

type ServiceStat struct {
	ServiceName string  `json:"service_name"`
	SpanCount   int64   `json:"span_count"`
	ErrorCount  int64   `json:"error_count"`
	ErrorRate   float64 `json:"error_rate"`
	AvgDuration int64   `json:"avg_duration_us"`
	P50         int64   `json:"p50_us"`
	P95         int64   `json:"p95_us"`
	P99         int64   `json:"p99_us"`
}

type OperationStat struct {
	ServiceName   string `json:"service_name"`
	OperationName string `json:"operation_name"`
	Count         int64  `json:"count"`
	AvgDuration   int64  `json:"avg_duration_us"`
	P99           int64  `json:"p99_us"`
}

type EdgeStat struct {
	From        string `json:"from"`
	To          string `json:"to"`
	CallCount   int64  `json:"call_count"`
	AvgDuration int64  `json:"avg_duration_us"`
}
