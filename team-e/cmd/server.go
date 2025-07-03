package cmd

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type Server struct {
	Engine *gin.Engine
	port   string
	ip     string
}

func NewServer(port string) *Server {
	s := &Server{Engine: gin.New(), port: port}
	s.Engine.Use(gin.Logger())
	s.Engine.Use(gin.Recovery())
	s.Engine.Use(cors.New(cors.Config{
		AllowWebSockets:  true,
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))
	return s
}

func (s *Server) StartServer() error {
	log.Println("Go Server Starting...")
	return s.Engine.Run(s.port)
}
