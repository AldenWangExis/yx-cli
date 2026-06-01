package auth

import "context"

type Status struct {
	Profile  string `json:"profile"`
	HasToken bool   `json:"hasToken"`
	Backend  string `json:"backend"`
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
	return Status{Profile: profile, HasToken: token != "", Backend: p.store.Backend()}, nil
}

func (p *PATProvider) Status(ctx context.Context, profile string) (Status, error) {
	_, ok, err := p.store.Load(profile)
	if err != nil {
		return Status{}, err
	}
	return Status{Profile: profile, HasToken: ok, Backend: p.store.Backend()}, nil
}

func (p *PATProvider) Logout(ctx context.Context, profile string) error {
	return p.store.Delete(profile)
}
