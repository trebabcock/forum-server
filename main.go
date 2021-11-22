package main

import (
	"forum-server/app"
	"forum-server/audit"
)

func main() {
	auditor := audit.Auditor{}
	auditor.Init()

	app := app.App{}
	app.Init(&auditor)
	app.Run(":2814")
}
