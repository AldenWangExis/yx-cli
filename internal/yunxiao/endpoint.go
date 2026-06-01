package yunxiao

import "strings"

func NormalizeBaseURL(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return value
	}
	if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
		value = "https://" + value
	}
	if strings.HasPrefix(value, "https://devops.aliyun.com") || strings.HasPrefix(value, "http://devops.aliyun.com") {
		return "https://openapi-rdc.aliyuncs.com"
	}
	return strings.TrimRight(value, "/")
}
