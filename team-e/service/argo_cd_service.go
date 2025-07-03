package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"feather/repository"
	"feather/types"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

type ArgoCdService interface {
	CreateProjectManifestRepo(req *types.CreateCdRequest) error
	ensureApplicationSet(res *types.ProjectWithBaseCampInfo, repoName string, filePath string, namespace string) error
	ensureArgoCdRepo(res *types.ProjectWithBaseCampInfo, repoName string) error
}

type argoCdServiceImpl struct {
	repository *repository.Repository
	gitService GitService
}

func NewArgoCdService(repository *repository.Repository, gitService GitService) ArgoCdService {
	return &argoCdServiceImpl{
		repository: repository,
		gitService: gitService,
	}
}

func (s *argoCdServiceImpl) CreateProjectManifestRepo(req *types.CreateCdRequest) error {
	const (
		repoName         = "feather-argocd"
		appSetFolderPath = "application-sets"
		appSetFileName   = "application-set.yaml"
	)

	applicationSetFilePath := fmt.Sprintf("%s/%s", appSetFolderPath, appSetFileName)

	res, err := s.repository.ProjectWithBaseCampInfo(req.ProjectId)
	if err != nil {
		return fmt.Errorf("Get BaseCamp failed: %w", err)
	}

	log.Println(res.Token)
	log.Println(res.BaseCampURL)
	log.Println(res.ProjectURL)

	if err := s.ensureArgoCdRepo(res, repoName); err != nil {
		return err
	}

	namespace, err := s.ensureProjectManifest(req, res, repoName)
	if err != nil {
		return err
	}

	if err := s.ensureApplicationSet(res, repoName, applicationSetFilePath, namespace); err != nil {
		return err
	}

	return nil
}

func (s *argoCdServiceImpl) ensureProjectManifest(req *types.CreateCdRequest, res *types.ProjectWithBaseCampInfo, repoName string) (string, error) {
	const manifestFileName = "manifest.yaml"
	filePath := fmt.Sprintf("%s/%s/%s", repoName, res.ProjectName, manifestFileName)

	checkProjectManifestReq := &types.CheckFileRequest{
		URL:      res.BaseCampURL,
		Token:    res.Token,
		Owner:    res.BaseCampOwner,
		Repo:     repoName,
		FilePath: filePath,
	}

	exists, err := s.gitService.FileExists(checkProjectManifestReq)
	if err != nil {
		return "", fmt.Errorf("file check failed: %w", err)
	}
	log.Print("File Check Complete \n")

	if err != nil {
		return "", fmt.Errorf("failed to parsing application-set template: %w", err)
	}

	if exists {
		log.Printf("Project Manifest already exists at %s", filePath)
		return req.Namespace, nil
	}

	var buf bytes.Buffer
	tmpl, err := template.New("cd.tmpl").Funcs(sprig.TxtFuncMap()).ParseFiles("assets/templates/argo/cd.tmpl")

	if err := tmpl.Execute(&buf, req.CdTemplateConfig); err != nil {
		return "", fmt.Errorf("failed to execute application-set template: %w", err)
	}

	generatedYAML := buf.String()
	encodedYaml := base64.StdEncoding.EncodeToString([]byte(generatedYAML))

	author := &types.Author{
		Email: "feather@feather.com",
		Name:  "feather",
	}

	fileCommitDetails := &types.FileCommitDetails{
		Author:  *author,
		Content: encodedYaml,
		Message: "Create Project Manifest YAML",
	}

	createReq := &types.CreateFileRequest{
		URL:      res.BaseCampURL,
		Token:    res.Token,
		Owner:    res.BaseCampOwner,
		Repo:     repoName,
		FilePath: filePath,
		Details:  *fileCommitDetails,
	}

	if err := s.gitService.CreateFile(createReq); err != nil {
		return "", fmt.Errorf("failed to create project manifest   file: %w", err)
	}

	return req.Namespace, nil
}

func (s *argoCdServiceImpl) ensureApplicationSet(res *types.ProjectWithBaseCampInfo, repoName string, filePath string, namespace string) error {
	config, err := GetKubeConfig()
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes config: %w", err)
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	checkApplicationSetDirReq := &types.CheckFileRequest{
		URL:      res.BaseCampURL,
		Token:    res.Token,
		Owner:    res.BaseCampOwner,
		Repo:     repoName,
		FilePath: filePath,
	}

	exists, err := s.gitService.FileExists(checkApplicationSetDirReq)
	if err != nil {
		return fmt.Errorf("file check failed: %w", err)
	}
	log.Print("File Check Complete \n")

	if exists {
		log.Printf("ApplicationSet file already exists at %s", filePath)
		baseCampNameLower := strings.ToLower(res.BaseCampName)
		applicationSetName := fmt.Sprintf("%s-appset", baseCampNameLower)
		applicationSetURL := fmt.Sprintf("%s/%s.git", res.BaseCampURL, repoName)
		params := struct {
			ApplicationSetName string
			URL                string
			ClusterURL         string
		}{
			ApplicationSetName: applicationSetName,
			URL:                applicationSetURL,
			ClusterURL:         config.Host,
		}

		tmpl, err := template.New("application-set.tmpl").Funcs(sprig.TxtFuncMap()).ParseFiles("assets/templates/argo/application-set.tmpl")
		if err != nil {
			return err
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, params); err != nil {
			return fmt.Errorf("failed to execute application-set template: %w", err)
		}

		var obj unstructured.Unstructured
		if err := yaml.Unmarshal(buf.Bytes(), &obj); err != nil {
			return fmt.Errorf("failed to decode sensor YAML: %w", err)
		}

		gvr := schema.GroupVersionResource{
			Group:    "argoproj.io",
			Version:  "v1alpha1",
			Resource: "applicationsets",
		}

		resource, err := client.Resource(gvr).Namespace(namespace).Create(context.Background(), &obj, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create applicationset resource: %w", err)
		}

		log.Printf("Application Set created: %s", resource.GetName())
		return nil
	}

	baseCampNameLower := strings.ToLower(res.BaseCampName)
	applicationSetName := fmt.Sprintf("%s-appset", baseCampNameLower)
	applicationSetURL := fmt.Sprintf("%s/%s.git", res.BaseCampURL, repoName)
	params := struct {
		ApplicationSetName string
		URL                string
		ClusterURL         string
	}{
		ApplicationSetName: applicationSetName,
		URL:                applicationSetURL,
		ClusterURL:         config.Host,
	}

	tmpl, err := template.New("application-set.tmpl").Funcs(sprig.TxtFuncMap()).ParseFiles("assets/templates/argo/application-set.tmpl")
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return fmt.Errorf("failed to execute application-set template: %w", err)
	}

	generatedYAML := buf.String()
	encodedYaml := base64.StdEncoding.EncodeToString([]byte(generatedYAML))

	author := &types.Author{
		Email: "feather@feather.com",
		Name:  "feather",
	}

	fileCommitDetails := &types.FileCommitDetails{
		Author:  *author,
		Content: encodedYaml,
		Message: "Create Application Set YAML",
	}

	createReq := &types.CreateFileRequest{
		URL:      res.BaseCampURL,
		Token:    res.Token,
		Owner:    res.BaseCampOwner,
		Repo:     repoName,
		FilePath: filePath,
		Details:  *fileCommitDetails,
	}

	if err := s.gitService.CreateFile(createReq); err != nil {
		return fmt.Errorf("failed to create application set file: %w", err)
	}

	var obj unstructured.Unstructured
	if err := yaml.Unmarshal(buf.Bytes(), &obj); err != nil {
		return fmt.Errorf("failed to decode sensor YAML: %w", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "applicationsets",
	}

	resource, err := client.Resource(gvr).Namespace(namespace).Create(context.Background(), &obj, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create applicationset resource: %w", err)
	}

	log.Printf("Application Set created: %s", resource.GetName())

	return nil
}

func (s *argoCdServiceImpl) ensureArgoCdRepo(res *types.ProjectWithBaseCampInfo, repoName string) error {
	log.Println("=====1=====")
	log.Println(res)

	checkArgoCdRepoReq := &types.CheckRepoRequest{
		URL:   res.BaseCampURL,
		Token: res.Token,
		Owner: res.BaseCampOwner,
		Name:  repoName,
	}

	exists, err := s.gitService.RepoExists(checkArgoCdRepoReq)
	if err != nil {
		return fmt.Errorf("repository check failed: %w", err)
	}

	if !exists {
		createReq := &types.CreateRepoRequest{
			URL:         res.BaseCampURL,
			Description: "Repository for ArgoCD manifest management",
			Name:        repoName,
			Owner:       res.BaseCampOwner,
			Private:     false,
			Token:       res.Token,
		}
		log.Println("=====2=====")
		log.Println(createReq.URL)
		log.Println(createReq.Name)
		log.Println(createReq.Owner)

		if err := s.gitService.CreateRepo(createReq); err != nil {
			return fmt.Errorf("failed to create ArgoCD repository: %w", err)
		}
		log.Printf("Repository '%s' created successfully.", repoName)
	} else {
		log.Printf("Repository '%s' already exists.", repoName)
	}

	return nil
}
