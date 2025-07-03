package service

import (
	"feather/repository"
	"feather/types"
	"fmt"
	"log"
)

type UserService interface {
	CreateUser(email string, password string, nickname string) error
	User(userId int64) (*types.User, error)
}

type userServiceImpl struct {
	repository *repository.Repository
}

func NewUserService(repository *repository.Repository) UserService {
	return &userServiceImpl{
		repository: repository,
	}
}

// func (service *Service) AuthUser(req *types.AuthUserReq) (*types.Response, error) {
// 	url := req.Url
// 	prefix := "/api/v1/user"

// 	authReq, err := http.NewRequest("GET", url + prefix, nil)
// 	if err != nil {
// 		log.Println("요청 생성 실패 : ", err.Error())
// 		return types.NewRes(401, nil, "요청 생성 실패")
// 	}

// 	authReq.Header.Set("Authorization", "token " + req.Token)

// 	client := &http.Client{}
// 	res, err := client.Do(authReq)
// 	if err != nil {
// 		log.Println("요청 실패 : ", err.Error())
// 		return types.NewRes(401, nil, "요청 실패")
// 	}

// 	defer res.Body.Close()

// 	if res.StatusCode != http.StatusOK {
// 		fmt.Printf("인증 실패: HTTP %d\n", res.StatusCode)
// 		return types.NewRes(401, nil, "인증 실패")
// 	}

// 	body, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		fmt.Println("응답 읽기 실패:", err)
// 		os.Exit(1)
// 	}

// 	var expectedUser types.GiteaUser
// 	if err := json.Unmarshal(body, &expectedUser); err != nil {
// 		fmt.Println("JSON 파싱 실패:", err)
// 		os.Exit(1)
// 	}

// 	if expectedUser.Login == req.Username {
// 		fmt.Println("토큰과 사용자 이름이 일치")
// 	} else {
// 		fmt.Printf("사용자 이름 불일치")
// 	}

// 	return types.NewRes(200, nil, "사용자 인증 성공")
// }

func (s *userServiceImpl) CreateUser(email string, password string, nickname string) error {
	err := s.repository.CreateUser(email, password, nickname)
	if err != nil {
		log.Println("회원 생성에 실패했습니다. : ", "err", err.Error())
		return fmt.Errorf("회원 생성 실패: %w", err)
	}
	return nil
}

func (s *userServiceImpl) User(userId int64) (*types.User, error) {
	res, err := s.repository.User(userId)
	if err != nil {
		log.Println("회원 조회에 실패했습니다. : ", "err", err.Error())
		return nil, fmt.Errorf("회원 조회 실패: %w", err)
	}
	return res, nil
}
