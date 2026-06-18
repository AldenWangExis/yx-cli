package cli

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AldenWangExis/yx-cli/internal/app"
)

func TestMemberCommands(t *testing.T) {
	members := &fakeMemberUseCase{
		members: []app.Member{
			{UserID: "u1", MemberID: "m1", Name: "王子豪", Email: "wang@example.com", Status: "ENABLED", RoleIDs: []string{"admin"}},
		},
		member: app.Member{UserID: "u1", MemberID: "m1", Name: "王子豪", Status: "ENABLED"},
	}
	opts := Options{ConfigPath: filepath.Join(t.TempDir(), "config.yaml"), MemberUseCase: members}

	stdout, stderr, err := executeCommand(t, NewRootCommandWithOptions(opts), "--json", "member", "list", "--status", "ENABLED")
	if err != nil {
		t.Fatalf("member list: %v stderr=%s", err, stderr)
	}
	var listed []app.Member
	if err := json.Unmarshal([]byte(stdout), &listed); err != nil {
		t.Fatalf("decode list: %v output=%s", err, stdout)
	}
	if members.listInput.Status != "ENABLED" || listed[0].UserID != "u1" {
		t.Fatalf("unexpected list input=%+v output=%+v", members.listInput, listed)
	}

	_, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "member", "search", "--name", "王")
	if err != nil {
		t.Fatalf("member search: %v stderr=%s", err, stderr)
	}
	if members.searchInput.Name != "王" {
		t.Fatalf("unexpected search input: %+v", members.searchInput)
	}

	stdout, stderr, err = executeCommand(t, NewRootCommandWithOptions(opts), "member", "get", "--user-id", "u1")
	if err != nil {
		t.Fatalf("member get: %v stderr=%s", err, stderr)
	}
	if members.getInput.UserID != "u1" || !strings.Contains(stdout, "王子豪") {
		t.Fatalf("unexpected get input=%+v output=%s", members.getInput, stdout)
	}
}

type fakeMemberUseCase struct {
	members     []app.Member
	member      app.Member
	listInput   app.MemberListInput
	searchInput app.MemberSearchInput
	getInput    app.MemberGetInput
}

func (u *fakeMemberUseCase) ListMembers(ctx context.Context, input app.MemberListInput) ([]app.Member, error) {
	u.listInput = input
	return u.members, nil
}

func (u *fakeMemberUseCase) SearchMembers(ctx context.Context, input app.MemberSearchInput) ([]app.Member, error) {
	u.searchInput = input
	return u.members, nil
}

func (u *fakeMemberUseCase) GetMember(ctx context.Context, input app.MemberGetInput) (app.Member, error) {
	u.getInput = input
	return u.member, nil
}
