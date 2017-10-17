package main

import (
	"github.com/kataras/iris"
	"./port_forward"
)

func main() {
	app := iris.New()
	app.StaticWeb("/assets", "./assets")
	port_forward.RegisterRouter(app)
	app.Run(iris.Addr("0.0.0.0:8872"))
}