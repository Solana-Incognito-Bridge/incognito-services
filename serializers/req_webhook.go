package serializers

type EventHookResult struct {
	StatusCode int `json:"status_code"`
	IsDone bool `json:"is_done"`
	Error string `json:"error"`
}
