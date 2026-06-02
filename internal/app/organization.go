package app

import "context"

type Organization struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
}

type OrganizationService interface {
	ListOrganizations(ctx context.Context) ([]Organization, error)
}

type OrganizationUseCase struct {
	organizations OrganizationService
}

func NewOrganizationUseCase(organizations OrganizationService) *OrganizationUseCase {
	return &OrganizationUseCase{organizations: organizations}
}

func (u *OrganizationUseCase) ListOrganizations(ctx context.Context) ([]Organization, error) {
	return u.organizations.ListOrganizations(ctx)
}
