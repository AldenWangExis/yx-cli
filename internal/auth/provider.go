package auth

import (
	"context"
	"strings"
)

type Status struct {
	Profile   string `json:"profile"`
	HasToken  bool   `json:"hasToken"`
	Backend   string `json:"backend"`
	TokenMask string `json:"token,omitempty"`
}

type Provider interface {
	Login(ctx context.Context, profile, token string) (Status, error)
	Status(ctx context.Context, profile string) (Status, error)
	Logout(ctx context.Context, profile string) error
}

type PATProvider struct {
	store TokenStore
}

func NewPATProvider(store TokenStore) *PATProvider {
	return &PATProvider{store: store}
}

func (p *PATProvider) Login(ctx context.Context, profile, token string) (Status, error) {
	if err := p.store.Save(profile, token); err != nil {
		return Status{}, err
	}
	return Status{Profile: profile, HasToken: token != "", Backend: p.store.Backend(), TokenMask: maskToken(token)}, nil
}

func (p *PATProvider) Status(ctx context.Context, profile string) (Status, error) {
	token, ok, err := p.store.Load(profile)
	if err != nil {
		return Status{}, err
	}
	return Status{Profile: profile, HasToken: ok, Backend: p.store.Backend(), TokenMask: maskToken(token)}, nil
}

func (p *PATProvider) Logout(ctx context.Context, profile string) error {
	return p.store.Delete(profile)
}

func maskToken(token string) string {
	if token == "" {
		return ""
	}
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	const prefixLen = 3
	const suffixLen = 4
	if len(token) <= prefixLen+suffixLen {
		return strings.Repeat("*", len(token))
	}
	return token[:prefixLen] + strings.Repeat("*", len(token)-prefixLen-suffixLen) + token[len(token)-suffixLen:]
}
