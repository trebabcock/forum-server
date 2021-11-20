package main

import "forum-server/app"

func main() {
	app := app.App{}
	app.Init()
	app.Run(":2814")
}
