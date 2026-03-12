package main

import (
	"embed"
	_ "embed"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/mszalewicz/veles/backend"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

func init() {
}

func main() {

	// TODO:
	// - save windows info in db

	//      App directory scheme under different OS:
	// 		Mac:       ~/Library/Application\ Scripts/<appName>/log
	// 		Windows:   C:\Users\<username>\AppData\Local\<appName>\log
	// 		Linux:     /var/lib/<appName>/log

	appName := "veles"
	current_os := runtime.GOOS
	logPath := ""
	applicationDBPath := ""
	appDirectory := ""

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	switch current_os {
	case "darwin":
		appDirectory = filepath.Join(usr.HomeDir, "/Library/Application Support", appName)
		logPath = filepath.Join(appDirectory, "log")
		applicationDBPath = filepath.Join(appDirectory, "application.sqlite")

	case "windows":
		appDirectory := filepath.Join(usr.HomeDir, "AppData\\Local", appName)
		logPath = filepath.Join(appDirectory, "log")
		applicationDBPath = filepath.Join(appDirectory, "application.sqlite")

	case "linux":
		appDirectory := "/var/lib/" + appName
		logPath = filepath.Join(appDirectory, "log")
		applicationDBPath = filepath.Join(appDirectory, "application.sqlite")
	}

	_, err = os.Stat(appDirectory)
	if os.IsNotExist(err) {
		err := os.MkdirAll(appDirectory, os.ModePerm)
		if err != nil {
			log.Fatal("Error creating directory:", err)
		}
	} else if err != nil {
		log.Fatal("Error checking directory:", err)
	}

	// create log file
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		slog.Error("Could not create log file", "error", err)
		// TODO: show error in application window
		os.Exit(1)
	}
	defer logFile.Close()

	loggerArgs := &slog.HandlerOptions{AddSource: true}
	logger := slog.New(slog.NewJSONHandler(logFile, loggerArgs))
	slog.SetDefault(logger)

	// create db instance
	err = backend.Connect(applicationDBPath)
	defer backend.Close()

	if err != nil {
		slog.Error("Could not create local database", "error", err)
		// TODO: show error in application window
		os.Exit(1)
	}

	err = backend.ApplySchema()
	if err != nil {
		slog.Error("Could not apply database schema", "error", err)
		// TODO: show error in application window
		os.Exit(1)
	}

	var schemaVersion int
	err = backend.DB.QueryRow("PRAGMA user_version").Scan(&schemaVersion)
    if err != nil {
        log.Fatal(err)
    }

    if schemaVersion == 0 {

    }

	// errToHandleInGUI = backend.CreateStructure()

	// if errToHandleInGUI != nil {
	// 	slog.Error("Could not bootstrap DB from schema.", "error", errToHandleInGUI)
	// 	// TODO: show error in application window
	// }

	// Window state management:
	// didWindowChange := false
	// TODO: read this from db
	var windowRelativeX float32 = 0
	// TODO: read this from db
	var windowRelativeY float32 = 0
	// TODO: read this from db
	var windowRelativeHeight float32 = 0
	// TODO: read this from db
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
	// go func(proceed *bool, relativeX *float32, relativeY *float32, relativeHeight *float32, relativeWidth *float32) {
	// 	time.Sleep(5 * time.Second)

	// 	if *proceed {
	// 		// TODO save information to db
	// 		*proceed = false
	// 	}

	// }(&didWindowChange, &windowRelativeX, &windowRelativeY, &windowRelativeHeight, &windowRelativeWidth)

	// Adds window event that checks and saves relative position of the window.
	// User can position windows in their favourite place and application will rember it between start ups.
	window.OnWindowEvent(events.Common.WindowDidMove, func(e *application.WindowEvent) {
		x, y := window.Position()

		screen, err := window.GetScreen()
		if err != nil {
			log.Fatal(err)
		}

		screen_height := float32(screen.Size.Height) * screen.ScaleFactor
		screen_width := float32(screen.Size.Width) * screen.ScaleFactor

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

	err = app.Run()

	if err != nil {
		log.Fatal(err)
	}
}
