package main

import (
	_ "github.com/angelolily/htyhelpertools/boot"
	_ "github.com/angelolily/htyhelpertools/router"

	"github.com/gogf/gf/frame/g"
)

func main() {
	g.Server().Run()
}
