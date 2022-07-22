package main

import (
	_ "htyhelpertools/boot"
	_ "htyhelpertools/router"

	"github.com/gogf/gf/frame/g"
)

func main() {
	g.Server().Run()
}
