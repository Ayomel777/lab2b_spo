package main

import (
	"context"
	"os"

	"ruby-analyzer/internal/analyzer" // Убедитесь, что путь совпадает с вашим go.mod

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Analyze(source string) analyzer.Result {
	return analyzer.Analyze(source)
}

func (a *App) SelectFile() (string, error) {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Выберите Ruby-файл",
		Filters: []runtime.FileFilter{
			{DisplayName: "Ruby files", Pattern: "*.rb"},
		},
	})
	if err != nil {
		return "", err
	}
	return selection, nil
}

func (a *App) ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
