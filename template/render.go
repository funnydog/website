package template

import (
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/funnydog/website/config"
)

var ErrMissingBaseTemplates = errors.New("Missing base templates")

func Render(c *config.Configuration) error {
	layoutPath := filepath.Join(c.TemplateDir, c.LayoutName)
	barePath := filepath.Join(c.TemplateDir, c.BareName)

	entries, err := ioutil.ReadDir(c.TemplateDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if name == c.LayoutName || name == c.BareName ||
			filepath.Ext(name) != ".html" || entry.IsDir() {
			continue
		}

		t := template.New(name)
		t, err = t.ParseFiles(layoutPath, barePath, filepath.Join(c.TemplateDir, name))
		if err != nil {
			return err
		}

		destPath := filepath.Join(c.RenderDir, name)
		_ = os.MkdirAll(filepath.Dir(destPath), 0755)

		full, err := os.Create(destPath)
		if err != nil {
			log.Println(err)
			continue
		}
		defer full.Close()

		err = t.ExecuteTemplate(full, c.LayoutName, nil)
		if err != nil {
			log.Println(err)
			continue
		}

		destPath = filepath.Join(c.RenderDir, "nav", name)
		_ = os.MkdirAll(filepath.Dir(destPath), 0755)

		sub, err := os.Create(destPath)
		if err != nil {
			log.Println(err)
			continue
		}
		defer sub.Close()

		err = t.ExecuteTemplate(sub, c.BareName, nil)
		if err != nil {
			log.Println(err)
			continue
		}
	}

	return nil
}
