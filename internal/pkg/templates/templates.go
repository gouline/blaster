package templates

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Config struct {
	Logger *zap.Logger

	RootPath   string
	LayoutFile string
}

type Templates struct {
	config    Config
	templates map[string]*template.Template
}

// NewRenderer creates new renderer and parses templates directory recursively
// Relative path including extension is used as template name.
func New(config Config) (*Templates, error) {
	t := &Templates{
		config:    config,
		templates: map[string]*template.Template{},
	}

	if f, err := os.Stat(config.RootPath); os.IsNotExist(err) {
		return t, fmt.Errorf("root not found: %w", err)
	} else if err == nil && !f.IsDir() {
		return t, fmt.Errorf("root not directory")
	}

	layoutPath := config.RootPath + "/" + config.LayoutFile
	if f, err := os.Stat(layoutPath); os.IsNotExist(err) {
		return t, fmt.Errorf("layout not found: %w", err)
	} else if err == nil && f.IsDir() {
		return t, fmt.Errorf("layout is directory")
	}

	err := filepath.WalkDir(config.RootPath, func(path string, d fs.DirEntry, err error) error {
		name := d.Name()

		if d.IsDir() || name == config.LayoutFile {
			return nil
		} else if filepath.Ext(path) != filepath.Ext(layoutPath) {
			t.config.Logger.Info("discarding template", zap.String("path", path))
			return nil
		}

		t.templates[name] = template.Must(template.ParseFiles(layoutPath, path))
		t.config.Logger.Info("compiled template", zap.String("name", name), zap.String("path", path))

		return nil
	})
	if err != nil {
		return t, fmt.Errorf("failed to walk root: %w", err)
	}

	return t, nil
}

// Render implements [echo.Renderer] interface.
func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		return fmt.Errorf("template '%s' not found", name)
	}
	return tmpl.Execute(w, data)
}
