package safety

import (
	"bytes"
	"testing"
)

func TestDecisionTable(t *testing.T) {
	tests := []struct {
		name    string
		req     Request
		env     Environment
		want    Decision
		wantErr bool
	}{
		{
			name: "dry run skips mutation",
			req:  Request{DryRun: true},
			env:  Environment{ConfirmWrites: true, IsTerminal: false},
			want: Decision{Allowed: false, DryRun: true},
		},
		{
			name: "confirm disabled allows explicit write",
			req:  Request{},
			env:  Environment{ConfirmWrites: false, IsTerminal: false},
			want: Decision{Allowed: true},
		},
		{
			name: "yes bypasses required confirmation",
			req:  Request{Yes: true},
			env:  Environment{ConfirmWrites: true, IsTerminal: false},
			want: Decision{Allowed: true},
		},
		{
			name:    "non terminal confirmation fails without yes",
			req:     Request{},
			env:     Environment{ConfirmWrites: true, IsTerminal: false},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decide(tt.req, tt.env)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if got != tt.want {
				t.Fatalf("unexpected decision: want %+v got %+v", tt.want, got)
			}
		})
	}
}

func TestInteractiveConfirmation(t *testing.T) {
	var out bytes.Buffer
	env := Environment{
		ConfirmWrites: true,
		IsTerminal:    true,
		Input:         bytes.NewBufferString("y\n"),
		Output:        &out,
	}

	got, err := Decide(Request{Summary: "merge MR 1"}, env)
	if err != nil {
		t.Fatalf("expected confirmation to succeed, got: %v", err)
	}
	if !got.Allowed {
		t.Fatal("expected write to be allowed")
	}
	if out.String() == "" {
		t.Fatal("expected confirmation prompt")
	}
}
