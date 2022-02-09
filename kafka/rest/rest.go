package rest

const (
	contentType = "application/vnd.kafka.json.v2+json"
)

type postResponse struct {
	ErrorCode int64  `json:"error_code"`
	Message   string `json:"message"`
}
