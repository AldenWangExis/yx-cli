package projex

import (
	"encoding/json"
	"fmt"

	"github.com/AldenWangExis/yx-cli/internal/app"
)

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
	Description    string `json:"description,omitempty"`
	FormatType     string `json:"formatType,omitempty"`
	AssignedTo     string `json:"assignedTo,omitempty"`
}

type projectCreateRequest struct {
	Name        string `json:"name"`
	CustomCode  string `json:"customCode"`
	Scope       string `json:"scope"`
	TemplateID  string `json:"templateId"`
	Description string `json:"description,omitempty"`
}

type workitemUpdateRequest struct {
	Status      string `json:"status,omitempty"`
	AssignedTo  string `json:"assignedTo,omitempty"`
	Subject     string `json:"subject,omitempty"`
	Description string `json:"description,omitempty"`
	FormatType  string `json:"formatType,omitempty"`
}

func decodeProjects(data []byte) ([]app.Project, error) {
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

func decodeProjectTemplates(data []byte) ([]app.ProjectTemplate, error) {
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

func decodeProject(data []byte) (app.Project, error) {
	var response projectResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return app.Project{}, fmt.Errorf("decode project: %w", err)
	}
	return app.Project{ID: response.ID, Name: response.Name, CustomCode: response.CustomCode, Scope: response.Scope}, nil
}

func decodeWorkitems(data []byte) ([]app.WorkitemListItem, error) {
	var response []workitemResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("decode workitems: %w", err)
	}
	items := make([]app.WorkitemListItem, 0, len(response))
	for _, item := range response {
		items = append(items, app.WorkitemListItem{
			ID:        item.ID,
			Title:     item.Subject,
			Status:    item.Status.Name,
			Type:      item.WorkitemType.Name,
			ProjectID: item.Space.ID,
		})
	}
	return items, nil
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
