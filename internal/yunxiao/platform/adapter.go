package platform

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
