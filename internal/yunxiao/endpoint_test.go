package yunxiao

import "testing"

func TestNormalizeBaseURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "web workbench domain maps to openapi", in: "https://devops.aliyun.com", want: "https://openapi-rdc.aliyuncs.com"},
		{name: "web workbench url maps to openapi", in: "https://devops.aliyun.com/workbench?orgId=abc", want: "https://openapi-rdc.aliyuncs.com"},
		{name: "missing scheme gets https", in: "openapi-rdc.aliyuncs.com", want: "https://openapi-rdc.aliyuncs.com"},
		{name: "openapi stays openapi", in: "https://openapi-rdc.aliyuncs.com", want: "https://openapi-rdc.aliyuncs.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeBaseURL(tt.in); got != tt.want {
				t.Fatalf("want %q got %q", tt.want, got)
			}
		})
	}
}
