package cli

import (
	"context"
	"fmt"

	"github.com/AldenWangExis/yx-cli/internal/yunxiao/platform"
)

type currentUserAssigneeResolver struct {
	platform interface {
		CurrentUser(ctx context.Context) (platform.User, error)
	}
}

func (r currentUserAssigneeResolver) ResolveAssignee(ctx context.Context, assignee string) (string, error) {
	if assignee != "@me" {
		return assignee, nil
	}
	user, err := r.platform.CurrentUser(ctx)
	if err != nil {
		return "", err
	}
	if user.AccountID == "" {
		return "", fmt.Errorf("current user id is unavailable")
	}
	return user.AccountID, nil
}
