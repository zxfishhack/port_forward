package main

import (
	"github.com/kataras/iris"
	"./port_forward"
	"flag"
)

func main() {
	app := iris.New()
	app.StaticWeb("/assets", "./assets")
	console := flag.String("l", ":8872", "listen addr")
	flag.Parse()
	port_forward.RegisterRouter(app)
	app.Run(iris.Addr(*console))
}