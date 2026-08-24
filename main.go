package main

import (
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/unit"

	"kripdroid/ui"
)

func main() {
	go func() {
		w := new(app.Window)
		w.Option(
			app.Title("KripDroid"),
			app.Size(unit.Dp(420), unit.Dp(780)),
			app.MinSize(unit.Dp(320), unit.Dp(480)),
		)
		if err := ui.Run(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}
