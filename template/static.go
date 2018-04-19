package template

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/funnydog/website/config"
)

func CopyStatic(c *config.Configuration) error {
	pathNames := []string{""}
	for len(pathNames) > 0 {
		baseDirName := pathNames[len(pathNames)-1]
		pathNames = pathNames[:len(pathNames)-1]

		entries, err := ioutil.ReadDir(filepath.Join(c.StaticDir, baseDirName))
		if err != nil {
			log.Println(err)
			continue
		}

		for _, entry := range entries {
			name := filepath.Join(baseDirName, entry.Name())
			srcName := filepath.Join(c.StaticDir, name)
			dstName := filepath.Join(c.RenderDir, name)
			if entry.IsDir() {
				pathNames = append(pathNames, name)
				_ = os.Mkdir(dstName, 0755)
			} else {
				in, err := os.Open(srcName)
				if err != nil {
					return err
				}
				defer in.Close()

				out, err := os.Create(dstName)
				if err != nil {
					return err
				}
				defer out.Close()

				if _, err = io.Copy(out, in); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
