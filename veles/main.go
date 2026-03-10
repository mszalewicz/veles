package main

import (
	"embed"
	_ "embed"
	"fmt"
	"log"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

func init() {
}

func main() {

	// TODO:
	// - add db connection
	// - add logger (slog)
	// - save windows info in db

	// Window state management:
	didWindowChange := false
	// read this from db
	var windowRelativeX float32 = 0
	// read this from db
	var windowRelativeY float32 = 0
	// read this from db
	var windowRelativeHeight float32 = 0
	// read this from db
	var windowRelativeWidth float32 = 0

	app := application.New(application.Options{
		Name:        "veles",
		Description: "A demo of using raw HTML & CSS",
		Services: []application.Service{
			application.NewService(&GreetService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Window 1",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
		Height:           900,
		Width:            600,
		MinHeight:        500,
		MinWidth:         500,
		// MaxHeight: 1000,
		// MaxWidth: 1000,
	})

	// Periodically saves window positioning and size if there was and update.
	go func(proceed *bool, relativeX *float32, relativeY *float32, relativeHeight *float32, relativeWidth *float32) {
		time.Sleep(5 * time.Second)

		if *proceed {
			// TODO save information to db
			*proceed = false
		}

	}(&didWindowChange, &windowRelativeX, &windowRelativeY, &windowRelativeHeight, &windowRelativeWidth)

	// Adds window event that checks and saves relative position of the window.
	// User can position windows in their favourite place and application will rember it between start ups.
	window.OnWindowEvent(events.Common.WindowDidMove, func(e *application.WindowEvent) {
		x, y := window.Position()


		screen, err := window.GetScreen()
		if err != nil {
			log.Fatal(err)
		}

		screen_height := float32(screen.Size.Height) * screen.ScaleFactor
		screen_width  := float32(screen.Size.Width) * screen.ScaleFactor

		windowRelativeX = float32(x) / float32(screen_width)
		windowRelativeY = float32(y) / float32(screen_height)

		fmt.Printf("Window relative size: width = %f, y = %f\n", windowRelativeX, windowRelativeY)
	})

	// Add window event that check and saves realtive size of the window.
	// User can set up window size as they please and application will rember it between start ups.
	window.OnWindowEvent(events.Common.WindowDidResize, func(e *application.WindowEvent) {
		x, y := window.Size()

		screen, err := window.GetScreen()
		if err != nil {
			log.Fatal(err)
		}

		screen_height := screen.Size.Height
		screen_width := screen.Size.Width

		windowRelativeWidth = float32(x) / float32(screen_width)
		windowRelativeHeight = float32(y) / float32(screen_height)

		fmt.Printf("Window relative size: width = %f, y = %f\n", windowRelativeWidth, windowRelativeHeight)
	})

	err := app.Run()

	if err != nil {
		log.Fatal(err)
	}
}
