package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	var out bytes.Buffer
	renderer := NewRenderer(&out)

	err := renderer.WriteJSON(map[string]string{"current": "default"})
	if err != nil {
		t.Fatalf("expected JSON write to succeed, got error: %v", err)
	}

	var got map[string]string
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("expected valid JSON, got error: %v output=%s", err, out.String())
	}
	if got["current"] != "default" {
		t.Fatalf("expected current default, got %q", got["current"])
	}
}

func TestWriteTable(t *testing.T) {
	var out bytes.Buffer
	renderer := NewRenderer(&out)

	err := renderer.WriteTable(
		[]string{"NAME", "DOMAIN"},
		[][]string{{"default", "https://devops.aliyun.com"}},
	)
	if err != nil {
		t.Fatalf("expected table write to succeed, got error: %v", err)
	}

	const want = "NAME     DOMAIN\n" +
		"default  https://devops.aliyun.com\n"
	if out.String() != want {
		t.Fatalf("unexpected table output:\nwant:\n%q\ngot:\n%q", want, out.String())
	}
}
