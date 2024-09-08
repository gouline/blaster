package templates

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

type Templates struct {
	templates map[string]*template.Template
}

// NewRenderer creates new renderer and parses templates directory recursively
// Relative path including extension is used as template name.
func New(root string, layout string) (*Templates, error) {
	t := &Templates{
		templates: map[string]*template.Template{},
	}

	if _, err := os.Stat(root); os.IsNotExist(err) {
		return t, err
	}

	layoutPath := root + "/" + layout
	layoutExt := filepath.Ext(layoutPath)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		name := d.Name()

		if d.IsDir() || name == layout || filepath.Ext(path) != layoutExt {
			return nil
		}

		t.templates[name] = template.Must(template.ParseFiles(layoutPath, path))

		return nil
	})

	return t, err
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		return fmt.Errorf("template '%s' not found", name)
	}
	return tmpl.Execute(w, data)
}
