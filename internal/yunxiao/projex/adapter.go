package projex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

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
	data, err := a.client.DoJSON(ctx, http.MethodPost, a.orgPath("/projects:search"), nil, []byte(`{}`))
	if err != nil {
		return nil, err
	}
	var response []projectResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("decode projects: %w", err)
	}
	projects := make([]app.Project, 0, len(response))
	for _, project := range response {
		projects = append(projects, app.Project{ID: project.ID, Name: project.Name})
	}
	return projects, nil
}

func (a *Adapter) ListProjectTemplates(ctx context.Context) ([]app.ProjectTemplate, error) {
	data, err := a.client.DoJSON(ctx, http.MethodGet, a.orgPath("/projectTemplates"), nil, nil)
	if err != nil {
		return nil, err
	}
	var response []projectTemplateResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("decode project templates: %w", err)
	}
	templates := make([]app.ProjectTemplate, 0, len(response))
	for _, template := range response {
		templates = append(templates, app.ProjectTemplate{ID: template.ID, Name: template.Name})
	}
	return templates, nil
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
	data, err := a.client.DoJSON(ctx, http.MethodPost, a.orgPath("/projects"), nil, body)
	if err != nil {
		return app.Project{}, err
	}
	var response projectResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return app.Project{}, fmt.Errorf("decode project: %w", err)
	}
	return app.Project{ID: response.ID, Name: response.Name, CustomCode: response.CustomCode, Scope: response.Scope}, nil
}

func (a *Adapter) ListWorkitems(ctx context.Context, projectID string) ([]app.WorkitemListItem, error) {
	items := []app.WorkitemListItem{}
	for _, category := range []string{"Req", "Task", "Bug"} {
		body, err := json.Marshal(map[string]string{"spaceId": projectID, "category": category})
		if err != nil {
			return nil, err
		}
		data, err := a.client.DoJSON(ctx, http.MethodPost, a.orgPath("/workitems:search"), nil, body)
		if err != nil {
			return nil, err
		}
		var response []workitemResponse
		if err := json.Unmarshal(data, &response); err != nil {
			return nil, fmt.Errorf("decode workitems: %w", err)
		}
		for _, item := range response {
			items = append(items, app.WorkitemListItem{
				ID:        item.ID,
				Title:     item.Subject,
				Status:    item.Status.Name,
				Type:      item.WorkitemType.Name,
				ProjectID: item.Space.ID,
			})
		}
	}
	return items, nil
}

func (a *Adapter) GetWorkitem(ctx context.Context, id string) (app.WorkitemDetail, error) {
	data, err := a.client.DoJSON(ctx, http.MethodGet, a.orgPath("/workitems/"+url.PathEscape(id)), nil, nil)
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
	})
	if err != nil {
		return app.WorkitemDetail{}, err
	}
	data, err := a.client.DoJSON(ctx, http.MethodPost, a.orgPath("/workitems"), nil, body)
	if err != nil {
		return app.WorkitemDetail{}, err
	}
	return decodeWorkitem(data)
}

func (a *Adapter) UpdateWorkitem(ctx context.Context, input app.UpdateWorkitemInput) (app.WorkitemDetail, error) {
	body, err := json.Marshal(workitemUpdateRequest{
		Status:     input.Status,
		AssignedTo: input.Assignee,
	})
	if err != nil {
		return app.WorkitemDetail{}, err
	}
	data, err := a.client.DoJSON(ctx, http.MethodPut, a.orgPath("/workitems/"+url.PathEscape(input.ID)), nil, body)
	if err != nil {
		return app.WorkitemDetail{}, err
	}
	if len(data) == 0 {
		return app.WorkitemDetail{ID: input.ID, Status: input.Status, Assignee: input.Assignee}, nil
	}
	return decodeWorkitem(data)
}

func (a *Adapter) orgPath(path string) string {
	return fmt.Sprintf("/oapi/v1/projex/organizations/%s%s", url.PathEscape(a.client.OrganizationID()), path)
}

func decodeWorkitem(data []byte) (app.WorkitemDetail, error) {
	var response workitemResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return app.WorkitemDetail{}, fmt.Errorf("decode workitem: %w", err)
	}
	return app.WorkitemDetail{
		ID:        response.ID,
		Title:     response.Subject,
		Status:    response.Status.Name,
		Type:      response.WorkitemType.Name,
		ProjectID: response.Space.ID,
		Assignee:  response.AssignedTo.ID,
	}, nil
}

type projectResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CustomCode string `json:"customCode"`
	Scope      string `json:"scope"`
}

type projectTemplateResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type workitemResponse struct {
	ID           string `json:"id"`
	Subject      string `json:"subject"`
	Status       named  `json:"status"`
	WorkitemType named  `json:"workitemType"`
	Space        struct {
		ID string `json:"id"`
	} `json:"space"`
	AssignedTo struct {
		ID string `json:"id"`
	} `json:"assignedTo"`
}

type named struct {
	Name string `json:"name"`
}

type workitemCreateRequest struct {
	SpaceID        string `json:"spaceId"`
	WorkitemTypeID string `json:"workitemTypeId"`
	Subject        string `json:"subject"`
}

type projectCreateRequest struct {
	Name        string `json:"name"`
	CustomCode  string `json:"customCode"`
	Scope       string `json:"scope"`
	TemplateID  string `json:"templateId"`
	Description string `json:"description,omitempty"`
}

type workitemUpdateRequest struct {
	Status     string `json:"status,omitempty"`
	AssignedTo string `json:"assignedTo,omitempty"`
}
