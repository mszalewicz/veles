package main

import (
	"context"
	"embed"
	_ "embed"
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

var (
	WindowRelativeX      float64
	WindowRelativeY      float64
	WindowRelativeHeight float64
	WindowRelativeWidth  float64
)

func init() {
}

func main() {
	ctx := context.Background()
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

	var schemaVersion int
	err = backend.DB.QueryRow("PRAGMA user_version").Scan(&schemaVersion)
	if err != nil {
		slog.Error("Could not SQLite user_version", "error", err)
		// TODO: show error in application window
		os.Exit(1)
	}

	if schemaVersion == 0 {
		if err := backend.ApplySchema(); err != nil {
			slog.Error("Could not apply schema to database", "error", err)
			// TODO: show error in application window
			os.Exit(1)
		}

		err = backend.SafeInsertDefaultWindow(ctx, backend.InsertDefaultWindowParams{Width: 0.5, Height: 0.5, X: 0.25, Y: 0.25})

		if err != nil {
			slog.Error("Could not apply schema to database", "error", err)
			// TODO: show error in application window
			os.Exit(1)
		}

		_, err = backend.DB.Exec("PRAGMA user_version = 1")

		err = backend.SafeSqliteUserVersion(1)

		if err != nil {
			slog.Error("Could not update user_version", "error", err)
			// TODO: show error in application window
			os.Exit(1)
		}
	}

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
		Height:           100,
		Width:            100,
		MinHeight:        100,
		MinWidth:         100,
		X:                100,
		Y:                100,
		Hidden:           true,
	})

	app.OnShutdown(func() {
		err := backend.SafeUpdateWindowGeometry(ctx, backend.UpdateWindowGeometryParams{
			Height: WindowRelativeHeight,
			Width:  WindowRelativeWidth,
			X:      WindowRelativeX,
			Y:      WindowRelativeY,
		})

		if err != nil {
			slog.Error("Could not update window geometry:", "error", err)
			// TODO: show error in application window
			os.Exit(1)
		}
	})

	// Read saved window position + size and apply before showing window.
	app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(e *application.ApplicationEvent) {
		screen, err := window.GetScreen()
		if err != nil {
			slog.Error("Nil screen object:", "error", err)
			os.Exit(1)
		}

		windowGeometry, err := backend.Q.GetWindowGeometry(ctx)
		if err != nil {
			slog.Error("Could not read window geometry data from database:", "error", err)
			os.Exit(1)
		}

		width := int(windowGeometry.Width * float64(screen.Size.Width))
		height := int(windowGeometry.Height * float64(screen.Size.Height))

		x := int(windowGeometry.X * float64(screen.Size.Width) * float64(screen.ScaleFactor))
		y := int(windowGeometry.Y * float64(screen.Size.Height) * float64(screen.ScaleFactor))

		minHeight := int(0.1 * float64(screen.Size.Height))
		minWidth := int(0.1 * float64(screen.Size.Width))

		window.SetMinSize(minWidth, minHeight)
		window.SetPosition(x, y)
		window.SetSize(width, height)
		window.Show()
	})

	// Adds window event that checks and saves relative position of the window.
	// User can position windows in their favourite place and application will rember it between start ups.
	window.OnWindowEvent(events.Common.WindowDidMove, func(e *application.WindowEvent) {
		x, y := window.Position()

		screen, err := window.GetScreen()
		if err != nil {
			log.Fatal(err)
		}

		screen_height := float64(screen.Size.Height) * float64(screen.ScaleFactor)
		screen_width := float64(screen.Size.Width) * float64(screen.ScaleFactor)

		WindowRelativeX = float64(x) / float64(screen_width)
		WindowRelativeY = float64(y) / float64(screen_height)

		// fmt.Printf("Window relative size: width = %f, y = %f\n", WindowRelativeX, WindowRelativeY)
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

		WindowRelativeWidth = float64(x) / float64(screen_width)
		WindowRelativeHeight = float64(y) / float64(screen_height)

		// fmt.Printf("Window relative size: width = %f, y = %f\n", WindowRelativeWidth, WindowRelativeHeight)
	})

	err = app.Run()

	if err != nil {
		log.Fatal(err)
	}
}
