package service

import (
	"feather/repository"
	"feather/types"
	"fmt"
	"log"
)

type BasecampService interface {
	CreateBasecamp(name string, url string, token string, owner string, userId int64) error
	BasecampsByUserId(userId int64) ([]*types.Basecamp, error)
	Basecamp(baseCampId int64) (*types.Basecamp, error)
}

type basecampServiceImpl struct {
	repository *repository.Repository
}

func NewBasecampService(repository *repository.Repository) BasecampService {
	return &basecampServiceImpl{
		repository: repository,
	}
}

func (s *basecampServiceImpl) CreateBasecamp(name string, url string, token string, owner string, userId int64) error {
	err := s.repository.CreateBasecamp(name, url, token, owner, userId)
	if err != nil {
		log.Println("베이스캠프 생성에 실패했습니다. : ", "err", err.Error())
		return fmt.Errorf("베이스캠프 생성 실패: %w", err)
	}
	return nil
}

func (s *basecampServiceImpl) BasecampsByUserId(userId int64) ([]*types.Basecamp, error) {
	res, err := s.repository.BasecampsByUserId(userId)
	if err != nil {
		log.Println("베이스캠프 조회에 실패했습니다. : ", "err", err.Error())
		return nil, fmt.Errorf("베이스캠프 조회 실패: %w", err)
	}
	return res, nil
}

func (s *basecampServiceImpl) Basecamp(baseCampId int64) (*types.Basecamp, error) {
	res, err := s.repository.Basecamp(baseCampId)
	if err != nil {
		log.Println("베이스캠프 조회에 실패했습니다. : ", "err", err.Error())
		return nil, fmt.Errorf("베이스캠프 조회 실패: %w", err)
	}
	return res, nil
}
