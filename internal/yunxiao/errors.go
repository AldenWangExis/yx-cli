package yunxiao

import "fmt"

type APIError struct {
	StatusCode int
	RequestID  string
	Body       string
}

func (e APIError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("yunxiao api error: status %d request_id %s body %s", e.StatusCode, e.RequestID, e.Body)
	}
	return fmt.Sprintf("yunxiao api error: status %d body %s", e.StatusCode, e.Body)
}
