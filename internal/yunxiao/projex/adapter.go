package projex

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/AldenWangExis/yx-cli/internal/app"
	"github.com/AldenWangExis/yx-cli/internal/yunxiao"
)

type Adapter struct {
	client *yunxiao.Client
}

func NewAdapter(config yunxiao.ClientConfig) *Adapter {
	return &Adapter{client: yunxiao.NewClient(config)}
}

func (a *Adapter) ListProjects(ctx context.Context) ([]app.Project, error) {
	paths := newProjexPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodPost, paths.projectsSearchPath(), nil, []byte(`{}`))
	if err != nil {
		return nil, err
	}
	return decodeProjects(data)
}

func (a *Adapter) ListProjectTemplates(ctx context.Context) ([]app.ProjectTemplate, error) {
	paths := newProjexPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodGet, paths.projectTemplatesPath(), nil, nil)
	if err != nil {
		return nil, err
	}
	return decodeProjectTemplates(data)
}

func (a *Adapter) CreateProject(ctx context.Context, input app.CreateProjectInput) (app.Project, error) {
	body, err := json.Marshal(projectCreateRequest{
		Name:        input.Name,
		CustomCode:  input.CustomCode,
		Scope:       input.Scope,
		TemplateID:  input.TemplateID,
		Description: input.Description,
	})
	if err != nil {
		return app.Project{}, err
	}
	paths := newProjexPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodPost, paths.projectsPath(), nil, body)
	if err != nil {
		return app.Project{}, err
	}
	return decodeProject(data)
}

func (a *Adapter) ListWorkitems(ctx context.Context, projectID string) ([]app.WorkitemListItem, error) {
	items := []app.WorkitemListItem{}
	paths := newProjexPaths(a.client)
	for _, category := range []string{"Req", "Task", "Bug"} {
		body, err := json.Marshal(map[string]string{"spaceId": projectID, "category": category})
		if err != nil {
			return nil, err
		}
		data, err := a.client.DoJSON(ctx, http.MethodPost, paths.workitemsSearchPath(), nil, body)
		if err != nil {
			return nil, err
		}
		categoryItems, err := decodeWorkitems(data)
		if err != nil {
			return nil, err
		}
		items = append(items, categoryItems...)
	}
	return items, nil
}

func (a *Adapter) GetWorkitem(ctx context.Context, id string) (app.WorkitemDetail, error) {
	paths := newProjexPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodGet, paths.workitemPath(id), nil, nil)
	if err != nil {
		return app.WorkitemDetail{}, err
	}
	return decodeWorkitem(data)
}

func (a *Adapter) CreateWorkitem(ctx context.Context, input app.CreateWorkitemInput) (app.WorkitemDetail, error) {
	body, err := json.Marshal(workitemCreateRequest{
		SpaceID:        input.ProjectID,
		WorkitemTypeID: input.Type,
		Subject:        input.Title,
		Description:    input.Description,
		FormatType:     normalizeDescriptionFormat(input.DescriptionFormat),
		AssignedTo:     input.Assignee,
	})
	if err != nil {
		return app.WorkitemDetail{}, err
	}
	paths := newProjexPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodPost, paths.workitemsPath(), nil, body)
	if err != nil {
		return app.WorkitemDetail{}, err
	}
	return decodeWorkitem(data)
}

func normalizeDescriptionFormat(format string) string {
	switch format {
	case "markdown", "MARKDOWN":
		return "MARKDOWN"
	case "richtext", "RICHTEXT":
		return "RICHTEXT"
	default:
		return format
	}
}

func (a *Adapter) UpdateWorkitem(ctx context.Context, input app.UpdateWorkitemInput) (app.WorkitemDetail, error) {
	body, err := json.Marshal(workitemUpdateRequest{
		Status:      input.Status,
		AssignedTo:  input.Assignee,
		Subject:     input.Title,
		Description: input.Description,
		FormatType:  normalizeDescriptionFormat(input.DescriptionFormat),
	})
	if err != nil {
		return app.WorkitemDetail{}, err
	}
	paths := newProjexPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodPut, paths.workitemPath(input.ID), nil, body)
	if err != nil {
		return app.WorkitemDetail{}, err
	}
	if len(data) == 0 {
		return app.WorkitemDetail{ID: input.ID, Status: input.Status, Assignee: input.Assignee}, nil
	}
	return decodeWorkitem(data)
}

func (a *Adapter) DeleteWorkitem(ctx context.Context, id string) (app.WorkitemDetail, error) {
	paths := newProjexPaths(a.client)
	data, err := a.client.DoJSON(ctx, http.MethodDelete, paths.workitemPath(id), nil, nil)
	if err != nil {
		return app.WorkitemDetail{}, err
	}
	if len(data) == 0 {
		return app.WorkitemDetail{ID: id}, nil
	}
	detail, err := decodeWorkitem(data)
	if err != nil {
		return app.WorkitemDetail{}, err
	}
	if detail.ID == "" {
		detail.ID = id
	}
	return detail, nil
}
