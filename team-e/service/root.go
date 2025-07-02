package service

import (
	"feather/repository"

	"k8s.io/client-go/rest"
)

type Service struct {
	UserService         UserService
	BasecampService     BasecampService
	ProjectService      ProjectService
	GitService          GitService
	ArgoWorkflowService ArgoWorkflowService
	ArgoSensorService   ArgoSensorService
	ArgoCdService       ArgoCdService
	repository          *repository.Repository
}

func NewService(repository *repository.Repository) *Service {
	gitSvc := NewGitService(repository)
	argoWorkflowSvc := NewArgoWorkflowSerivce()

	return &Service{
		UserService:         NewUserService(repository),
		BasecampService:     NewBasecampService(repository),
		ProjectService:      NewProjectService(repository),
		GitService:          gitSvc,
		ArgoWorkflowService: argoWorkflowSvc,
		ArgoSensorService:   NewArgoSensorService(argoWorkflowSvc),
		ArgoCdService:       NewArgoCdService(repository, gitSvc),
		repository:          repository,
	}
}

func GetKubeConfig() (*rest.Config, error) {
	return rest.InClusterConfig()
}
