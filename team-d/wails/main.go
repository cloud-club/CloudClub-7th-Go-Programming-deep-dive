package main

import (
	"context"
	"embed"

	"wails/backend/client"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()
	client, err := client.NewClient()

	if err != nil {
		println("Error:", err.Error())
	}

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "wails",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			client.SetWailsContext(ctx) // 여기서 context를 전달
		},
		Bind: []interface{}{
			app,
			client,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
