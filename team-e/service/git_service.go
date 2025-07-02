package service

import (
	"encoding/json"
	"feather/repository"
	"feather/types"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type GitService interface {
	CreateRepoBasedTemplate(req *types.RepoFromTemplateRequest) (*types.Response, error)

	RepoExists(req *types.CheckRepoRequest) (bool, error)
	CreateRepo(req *types.CreateRepoRequest) error

	FileExists(req *types.CheckFileRequest) (bool, error)
	CreateFile(req *types.CreateFileRequest) error
	GetFileContent(req *types.GetFileRequest) (string, error)
}

type gitServiceImpl struct {
	httpClient *HTTPClient
	repository *repository.Repository
}

func NewGitService(repository *repository.Repository) GitService {
	return &gitServiceImpl{
		httpClient: NewHTTPClient(),
		repository: repository,
	}
}

func (s *gitServiceImpl) CreateRepoBasedTemplate(req *types.RepoFromTemplateRequest) (*types.Response, error) {
	repoURL := fmt.Sprintf("%s/api/v1/repos/%s/%s/generate", req.URL, req.Template.Owner, req.Template.Repo)

	fmt.Println(req.URL)

	token, err := s.repository.TokenByBasecampId(req.BaseCampId)
	if err != nil {
		return nil, fmt.Errorf("get token by basecampId failed: %w", err)
	}

	payload := map[string]interface{}{
		"avatar":           req.Options.Avatar,
		"default_branch":   req.Options.DefaultBranch,
		"description":      req.Options.Description,
		"git_content":      req.Options.GitContent,
		"git_hooks":        req.Options.GitHooks,
		"labels":           req.Options.Labels,
		"name":             req.Name,
		"owner":            req.Owner,
		"private":          req.Private,
		"protected_branch": req.Options.ProtectedBranch,
		"topics":           req.Options.Topics,
		"webhooks":         req.Options.Webhooks,
	}

	res, err := s.httpClient.JSONPost(repoURL, token, payload)
	if err != nil {
		return nil, fmt.Errorf("repository creation failed: %w", err)
	}
	defer res.Body.Close()

	fmt.Println("====3====")

	var result types.Response
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if req.WebhookEnabled && req.Webhook != nil {
		if err := s.attachWebhook(req); err != nil {
			return nil, err
		}
	}

	log.Printf("Repository created: %s/%s", req.Owner, req.Name)
	return &result, nil
}

func (s *gitServiceImpl) attachWebhook(req *types.RepoFromTemplateRequest) error {
	hookURL := fmt.Sprintf("%s/api/v1/repos/%s/%s/hooks", req.URL, req.Owner, req.Name)

	hookType := req.Webhook.Type
	if hookType == "" {
		hookType = "gitea"
	}

	token, err := s.repository.TokenByBasecampId(req.BaseCampId)
	if err != nil {
		return fmt.Errorf("get token by basecampId failed: %w", err)
	}

	payload := map[string]interface{}{
		"type": hookType,
		"config": map[string]string{
			"url":          req.Webhook.URL,
			"content_type": "json",
		},
		"events":        []string{"push"},
		"branch_filter": req.Webhook.BranchFilter,
		"active":        true,
	}

	if _, err := s.httpClient.JSONPost(hookURL, token, payload); err != nil {
		return fmt.Errorf("webhook creation failed: %w", err)
	}
	log.Printf("Webhook created for: %s/%s", req.Owner, req.Name)
	return nil
}

func (s *gitServiceImpl) RepoExists(req *types.CheckRepoRequest) (bool, error) {
	repoURL := fmt.Sprintf("%s/api/v1/repos/%s/%s", req.URL, req.Owner, req.Name)
	_, err := s.httpClient.JSONGet(repoURL, req.Token)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (s *gitServiceImpl) CreateRepo(req *types.CreateRepoRequest) error {
	repoURL := fmt.Sprintf("%s/api/v1/user/repos", req.URL)
	payload := map[string]interface{}{
		"Description": req.Description,
		"Name":        req.Name,
		"Private":     req.Private,
	}
	_, err := s.httpClient.JSONPost(repoURL, req.Token, payload)
	return err
}

func (s *gitServiceImpl) FileExists(req *types.CheckFileRequest) (bool, error) {
	repoURL := fmt.Sprintf("%s/api/v1/repos/%s/%s/contents/%s", req.URL, req.Owner, req.Repo, req.FilePath)
	_, err := s.httpClient.JSONGet(repoURL, req.Token)

	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func isNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found")
}

func (s *gitServiceImpl) CreateFile(req *types.CreateFileRequest) error {
	repoURL := fmt.Sprintf("%s/api/v1/repos/%s/%s/contents/%s", req.URL, req.Owner, req.Repo, req.FilePath)

	payload := map[string]interface{}{
		"Author":    req.Details.Author,
		"Branch":    req.Details.Branch,
		"NewBranch": req.Details.NewBranch,
		"Content":   req.Details.Content,
		"Message":   req.Details.Message,
	}

	_, err := s.httpClient.JSONPost(repoURL, req.Token, payload)
	return err
}

func (s *gitServiceImpl) GetFileContent(req *types.GetFileRequest) (string, error) {
	// URL 경로 인코딩
	encodedFilePath := url.PathEscape(req.FilePath)

	// Gitea API URL 구성
	// GET /repos/{owner}/{repo}/contents/{filepath}
	apiURL := fmt.Sprintf("%s/api/v1/repos/%s/%s/contents/%s",
		strings.TrimRight(req.URL, "/"),
		req.Owner,
		req.Repo,
		encodedFilePath)

	result, err := s.httpClient.JSONGet(apiURL, req.Token)
	if err != nil {
		return "", fmt.Errorf("failed to get file content: %w", err)
	}

	if result.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(result.Body)
		return "", fmt.Errorf("gitea API error: status %d, body: %s", result.StatusCode, string(body))
	}

	body, err := io.ReadAll(result.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// JSON 응답 파싱
	var fileResponse *types.GiteaFileResponse
	if err := json.Unmarshal(body, &fileResponse); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// 파일 내용 반환 (Gitea는 base64로 인코딩된 내용을 반환)
	return fileResponse.Content, nil
}
