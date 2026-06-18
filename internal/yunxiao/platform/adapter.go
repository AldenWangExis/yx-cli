package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao"
)

type Adapter struct {
	client *yunxiao.Client
}

type User struct {
	Name      string
	Username  string
	AccountID string
	Email     string
}

func NewAdapter(config yunxiao.ClientConfig) *Adapter {
	return &Adapter{client: yunxiao.NewClient(config)}
}

func (a *Adapter) CurrentUser(ctx context.Context) (User, error) {
	data, err := a.client.DoJSON(ctx, http.MethodGet, "/oapi/v1/platform/user", nil, nil)
	if err != nil {
		return User{}, err
	}

	var envelope struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Username   string `json:"username"`
		Email      string `json:"email"`
		AccountID  string `json:"accountId"`
		AccountID2 string `json:"accountID"`
		Data       struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			Username   string `json:"username"`
			Email      string `json:"email"`
			AccountID  string `json:"accountId"`
			AccountID2 string `json:"accountID"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return User{}, err
	}

	user := User{
		Name:      firstNonEmpty(envelope.Data.Name, envelope.Name, envelope.Data.Username, envelope.Username),
		Username:  firstNonEmpty(envelope.Data.Username, envelope.Username, envelope.Data.Email, envelope.Email),
		AccountID: firstNonEmpty(envelope.Data.AccountID, envelope.Data.AccountID2, envelope.Data.ID, envelope.AccountID, envelope.AccountID2, envelope.ID),
		Email:     firstNonEmpty(envelope.Data.Email, envelope.Email),
	}
	return user, nil
}

func (a *Adapter) ListOrganizations(ctx context.Context) ([]app.Organization, error) {
	organizations := []app.Organization{}
	for page := 1; ; page++ {
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("perPage", "100")
		data, err := a.client.DoJSON(ctx, http.MethodGet, "/oapi/v1/platform/organizations", query, nil)
		if err != nil {
			return nil, err
		}
		items, err := decodeOrganizations(data)
		if err != nil {
			return nil, err
		}
		organizations = append(organizations, items...)
		if len(items) < 100 {
			break
		}
	}
	return organizations, nil
}

func (a *Adapter) ListMembers(ctx context.Context, input app.MemberListInput) ([]app.Member, error) {
	members := []app.Member{}
	for page := 1; ; page++ {
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("perPage", "100")
		if input.Status != "" {
			query.Set("status", input.Status)
		}
		data, err := a.client.DoJSON(ctx, http.MethodGet, a.organizationMembersPath(), query, nil)
		if err != nil {
			return nil, err
		}
		items, err := decodeMembers(data)
		if err != nil {
			return nil, err
		}
		members = append(members, items...)
		if len(items) < 100 {
			break
		}
	}
	return members, nil
}

func (a *Adapter) SearchMembers(ctx context.Context, input app.MemberSearchInput) ([]app.Member, error) {
	body, err := json.Marshal(map[string]string{
		"keyword": firstNonEmpty(input.Name, input.Email),
		"name":    input.Name,
		"email":   input.Email,
		"status":  input.Status,
	})
	if err != nil {
		return nil, err
	}
	data, err := a.client.DoJSON(ctx, http.MethodPost, a.organizationMembersSearchPath(), nil, body)
	if err != nil {
		return nil, err
	}
	members, err := decodeMembers(data)
	if err != nil {
		return nil, err
	}
	return filterMembers(members, input), nil
}

func (a *Adapter) GetMember(ctx context.Context, input app.MemberGetInput) (app.Member, error) {
	members, err := a.ListMembers(ctx, app.MemberListInput{})
	if err != nil {
		return app.Member{}, err
	}
	var found []app.Member
	for _, member := range members {
		if member.UserID == input.UserID {
			found = append(found, member)
		}
	}
	if len(found) == 0 {
		return app.Member{}, fmt.Errorf("member with user id %q was not found", input.UserID)
	}
	if len(found) > 1 {
		return app.Member{}, fmt.Errorf("member with user id %q matched multiple records", input.UserID)
	}
	return found[0], nil
}

func (a *Adapter) organizationMembersPath() string {
	return "/oapi/v1/platform/organizations/" + url.PathEscape(a.client.OrganizationID()) + "/members"
}

func (a *Adapter) organizationMembersSearchPath() string {
	return "/oapi/v1/platform/organizations/" + url.PathEscape(a.client.OrganizationID()) + "/members:search"
}

func decodeOrganizations(data []byte) ([]app.Organization, error) {
	var raw []organizationEnvelope
	if err := json.Unmarshal(data, &raw); err == nil {
		return mapOrganizations(raw), nil
	}

	var envelope struct {
		Data []organizationEnvelope `json:"data"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}
	return mapOrganizations(envelope.Data), nil
}

type organizationEnvelope struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
}

type memberEnvelope struct {
	ID             string   `json:"id"`
	OrganizationID string   `json:"organizationId"`
	UserID         string   `json:"userId"`
	Name           string   `json:"name"`
	Email          string   `json:"email"`
	Status         string   `json:"status"`
	RoleIDs        []string `json:"roleIds"`
}

func mapOrganizations(raw []organizationEnvelope) []app.Organization {
	organizations := make([]app.Organization, 0, len(raw))
	for _, item := range raw {
		organizations = append(organizations, app.Organization{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			CreatedAt:   item.CreatedAt,
		})
	}
	return organizations
}

func decodeMembers(data []byte) ([]app.Member, error) {
	var raw []memberEnvelope
	if err := json.Unmarshal(data, &raw); err == nil {
		return mapMembers(raw), nil
	}

	var envelope struct {
		Data    []memberEnvelope `json:"data"`
		Members []memberEnvelope `json:"members"`
		Items   []memberEnvelope `json:"items"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}
	switch {
	case envelope.Data != nil:
		return mapMembers(envelope.Data), nil
	case envelope.Members != nil:
		return mapMembers(envelope.Members), nil
	default:
		return mapMembers(envelope.Items), nil
	}
}

func mapMembers(raw []memberEnvelope) []app.Member {
	members := make([]app.Member, 0, len(raw))
	for _, member := range raw {
		members = append(members, app.Member{
			UserID:   member.UserID,
			MemberID: member.ID,
			Name:     member.Name,
			Email:    member.Email,
			Status:   member.Status,
			RoleIDs:  member.RoleIDs,
		})
	}
	return members
}

func filterMembers(members []app.Member, input app.MemberSearchInput) []app.Member {
	filtered := make([]app.Member, 0, len(members))
	for _, member := range members {
		if input.Name != "" && !strings.Contains(member.Name, input.Name) {
			continue
		}
		if input.Email != "" && !strings.Contains(member.Email, input.Email) {
			continue
		}
		if input.Status != "" && member.Status != input.Status {
			continue
		}
		filtered = append(filtered, member)
	}
	return filtered
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
