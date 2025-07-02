package router

import (
	"feather/handler"
	"feather/service"

	"github.com/gin-gonic/gin"
)

func RegisterRouter(engine *gin.Engine, s *service.Service) {
	userHandler := handler.NewUserHandler(s.UserService)
	basecampHandler := handler.NewBasecampHandler(s.BasecampService)
	projectHandler := handler.NewProjectHandler(s.ProjectService)
	gitHandler := handler.NewGitHandler(s.GitService)
	ciHandler := handler.NewCiHandler(s.ArgoSensorService)
	cdHandler := handler.NewCdHandler(s.ArgoCdService)

	apiV1 := engine.Group("/api/v1")
	{
		userGroup := apiV1.Group("/users")
		{
			userGroup.GET("/:id", userHandler.User)
			userGroup.POST("/", userHandler.CreateUser)
		}

		basecampGroup := apiV1.Group("/basecamps")
		{
			basecampGroup.GET("/:id", basecampHandler.Basecamp)
			basecampGroup.GET("/user/:userId", basecampHandler.BasecampByUserId)
			basecampGroup.POST("/", basecampHandler.CreateBasecamp)
		}

		projectGroup := apiV1.Group("/projects")
		{
			projectGroup.GET("/:id", projectHandler.Project)
			projectGroup.GET("/basecamp/:basecampId", projectHandler.ProjectByBasecampId)
			projectGroup.POST("/", projectHandler.CreateProject)
		}

		gitGroup := apiV1.Group("/git")
		{
			gitGroup.POST("/repo", gitHandler.CreateRepo)
		}

		ciGroup := apiV1.Group("/ci")
		{
			ciGroup.POST("/", ciHandler.CreateArgoCi)
		}

		cdGroup := apiV1.Group("/cd")
		{
			cdGroup.POST("/", cdHandler.CreateArgoCd)
		}
	}
}
