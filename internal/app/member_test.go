package app

import (
	"context"
	"testing"
)

func TestMemberUseCaseListSearchAndGet(t *testing.T) {
	service := &fakeMemberService{
		members: []Member{
			{UserID: "u1", MemberID: "m1", Name: "王子豪", Email: "wang@example.com", Status: "ENABLED", RoleIDs: []string{"admin"}},
			{UserID: "u2", MemberID: "m2", Name: "王小明", Email: "ming@example.com", Status: "DISABLED"},
		},
	}
	useCase := NewMemberUseCase(service)

	listed, err := useCase.ListMembers(context.Background(), MemberListInput{Status: "ENABLED"})
	if err != nil {
		t.Fatalf("list members: %v", err)
	}
	if service.listInput.Status != "ENABLED" || len(listed) != 2 {
		t.Fatalf("unexpected list result=%+v input=%+v", listed, service.listInput)
	}

	searched, err := useCase.SearchMembers(context.Background(), MemberSearchInput{Name: "王"})
	if err != nil {
		t.Fatalf("search members: %v", err)
	}
	if service.searchInput.Name != "王" || len(searched) != 2 {
		t.Fatalf("unexpected search result=%+v input=%+v", searched, service.searchInput)
	}

	got, err := useCase.GetMember(context.Background(), MemberGetInput{UserID: "u1"})
	if err != nil {
		t.Fatalf("get member: %v", err)
	}
	if got.UserID != "u1" || got.Name != "王子豪" {
		t.Fatalf("unexpected member: %+v", got)
	}
}

func TestMemberUseCaseValidatesInputs(t *testing.T) {
	useCase := NewMemberUseCase(&fakeMemberService{})

	if _, err := useCase.SearchMembers(context.Background(), MemberSearchInput{}); err == nil {
		t.Fatal("expected empty search to fail")
	}
	if _, err := useCase.GetMember(context.Background(), MemberGetInput{}); err == nil {
		t.Fatal("expected get without user id to fail")
	}
}

type fakeMemberService struct {
	members     []Member
	listInput   MemberListInput
	searchInput MemberSearchInput
	getInput    MemberGetInput
}

func (s *fakeMemberService) ListMembers(ctx context.Context, input MemberListInput) ([]Member, error) {
	s.listInput = input
	return s.members, nil
}

func (s *fakeMemberService) SearchMembers(ctx context.Context, input MemberSearchInput) ([]Member, error) {
	s.searchInput = input
	return s.members, nil
}

func (s *fakeMemberService) GetMember(ctx context.Context, input MemberGetInput) (Member, error) {
	s.getInput = input
	for _, member := range s.members {
		if member.UserID == input.UserID {
			return member, nil
		}
	}
	return Member{}, nil
}
