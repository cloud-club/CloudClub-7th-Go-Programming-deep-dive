package service

import (
	"feather/repository"
	"feather/types"
	"fmt"
	"log"
)

type ProjectService interface {
	CreateProject(name string, url string, owner string, private bool, baseCampId int64) error
	ProjectsByBaseCampId(baseCampId int64) ([]*types.Project, error)
	Project(projectId int64) (*types.Project, error)
}

type projectServiceImpl struct {
	repository *repository.Repository
}

func NewProjectService(repository *repository.Repository) ProjectService {
	return &projectServiceImpl{
		repository: repository,
	}
}

func (s *projectServiceImpl) CreateProject(name string, url string, owner string, private bool, baseCampId int64) error {
	err := s.repository.CreateProject(name, url, owner, private, baseCampId)
	if err != nil {
		log.Println("프로젝트 생성에 실패했습니다. : ", "err", err.Error())
		return fmt.Errorf("프로젝트 생성 실패: %w", err)
	}
	return nil
}

func (s *projectServiceImpl) ProjectsByBaseCampId(baseCampId int64) ([]*types.Project, error) {
	res, err := s.repository.ProjectsByBaseCampId(baseCampId)
	if err != nil {
		log.Println("프로젝트 조회에 실패했습니다. : ", "err", err.Error())
		return nil, fmt.Errorf("프로젝트 조회 실패: %w", err)
	}
	return res, nil
}

func (s *projectServiceImpl) Project(projectId int64) (*types.Project, error) {
	res, err := s.repository.Project(projectId)
	if err != nil {
		log.Println("프로젝트 조회에 실패했습니다. : ", "err", err.Error())
		return nil, fmt.Errorf("프로젝트 조회 실패: %w", err)
	}
	return res, nil
}
