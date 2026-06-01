package platform

import (
	"context"
	"encoding/json"
	"net/http"

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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
