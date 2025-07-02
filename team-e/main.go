package main

import (
	"feather/cmd"
	"feather/config"
	"feather/repository"
	"feather/router"
	"feather/service"
	"flag"

	_ "github.com/go-sql-driver/mysql"
)

var pathFlag = flag.String("config", "./config.toml", "config set")
var port = flag.String("port", "localhost:8080", "port set")

func main() {
	flag.Parse()
	c := config.NewConfig(*pathFlag)
	r, err := repository.NewRepository(c)
	if err != nil {
		panic(err)
	}

	s := service.NewService(r)
	app := cmd.NewServer(*port)
	router.RegisterRouter(app.Engine, s)

	if err := app.StartServer(); err != nil {
		panic(err)
	}
}
