package app

import (
	"context"
	"fmt"
)

type Member struct {
	UserID   string   `json:"userId"`
	MemberID string   `json:"memberId,omitempty"`
	Name     string   `json:"name"`
	Email    string   `json:"email,omitempty"`
	Status   string   `json:"status,omitempty"`
	RoleIDs  []string `json:"roleIds,omitempty"`
}

type MemberListInput struct {
	Status string
}

type MemberSearchInput struct {
	Name   string
	Email  string
	Status string
}

type MemberGetInput struct {
	UserID string
}

type MemberService interface {
	ListMembers(ctx context.Context, input MemberListInput) ([]Member, error)
	SearchMembers(ctx context.Context, input MemberSearchInput) ([]Member, error)
	GetMember(ctx context.Context, input MemberGetInput) (Member, error)
}

type MemberUseCase struct {
	service MemberService
}

func NewMemberUseCase(service MemberService) *MemberUseCase {
	return &MemberUseCase{service: service}
}

func (u *MemberUseCase) ListMembers(ctx context.Context, input MemberListInput) ([]Member, error) {
	return u.service.ListMembers(ctx, input)
}

func (u *MemberUseCase) SearchMembers(ctx context.Context, input MemberSearchInput) ([]Member, error) {
	if input.Name == "" && input.Email == "" && input.Status == "" {
		return nil, fmt.Errorf("name, email, or status is required")
	}
	return u.service.SearchMembers(ctx, input)
}

func (u *MemberUseCase) GetMember(ctx context.Context, input MemberGetInput) (Member, error) {
	if input.UserID == "" {
		return Member{}, fmt.Errorf("user id is required")
	}
	return u.service.GetMember(ctx, input)
}
