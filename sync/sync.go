package sync

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/funnydog/website/backend"
)

type Element struct {
	Name string
	Hash string
}

func (e *Element) isDir() bool {
	return e.Hash == ""
}

func Create(b backend.Backend, chksumName, sourceDir, destDir string) (*Sync, error) {
	return &Sync{
		backend:    b,
		chksumName: chksumName,
		sourceDir:  sourceDir,
		destDir:    destDir,
	}, nil
}

type Sync struct {
	backend    backend.Backend
	chksumName string
	sourceDir  string
	destDir    string
}

func (u *Sync) GetRemoteChecksums() ([]Element, error) {
	elements := []Element{}

	res, err := u.backend.Fetch(path.Join(u.destDir, u.chksumName))
	if err != nil {
		return elements, nil
	}
	defer res.Close()

	sums, err := ioutil.ReadAll(res)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(sums), "\n")
	for _, line := range lines {
		values := strings.Split(line, " ")
		if len(values) > 1 {
			elements = append(elements, Element{
				Name: strings.Join(values[1:], " "),
				Hash: values[0],
			})
		}
	}

	sort.Slice(elements, func(i, j int) bool { return elements[i].Name < elements[j].Name })
	return elements, nil
}

func (u *Sync) GetLocalChecksums() ([]Element, error) {
	pathNames := []string{""}
	elements := []Element{}
	for len(pathNames) > 0 {
		baseDirName := pathNames[len(pathNames)-1]
		pathNames = pathNames[:len(pathNames)-1]

		entries, err := ioutil.ReadDir(path.Join(u.sourceDir, baseDirName))
		if err != nil {
			log.Println(err)
			continue
		}

		for _, entry := range entries {
			name := path.Join(baseDirName, entry.Name())
			if entry.IsDir() {
				pathNames = append(pathNames, name)
				elements = append(elements, Element{
					Name: name,
					Hash: "",
				})
			} else {
				data, err := ioutil.ReadFile(path.Join(u.sourceDir, name))
				if err != nil {
					log.Println(err)
					continue
				}

				hash := sha512.Sum512(data)
				elements = append(elements, Element{
					Name: name,
					Hash: hex.EncodeToString(hash[:]),
				})
			}
		}
	}

	sort.Slice(elements, func(i, j int) bool { return elements[i].Name < elements[j].Name })
	return elements, nil
}

func (u *Sync) StoreChecksums(elements []Element) error {
	var b bytes.Buffer
	for _, element := range elements {
		fmt.Fprintf(&b, "%s %s\n", element.Hash, element.Name)
	}

	err := u.backend.Store(path.Join(u.destDir, u.chksumName), &b)
	if err != nil {
		return err
	}

	return nil
}

func (u *Sync) SaveLocal(element *Element) error {
	log.Println("+", element.Name)
	if element.isDir() {
		_ = u.backend.MakeDir(path.Join(u.destDir, element.Name))
		return nil
	} else {
		file, err := os.Open(path.Join(u.sourceDir, element.Name))
		if err != nil {
			return err
		}
		defer file.Close()
		return u.backend.Store(path.Join(u.destDir, element.Name), file)
	}
}

func (u *Sync) DeleteRemote(remote *Element) error {
	log.Println("-", remote.Name)
	if remote.isDir() {
		_ = u.backend.RemoveDir(path.Join(u.destDir, remote.Name))
	} else {
		_ = u.backend.Delete(path.Join(u.destDir, remote.Name))
	}
	return nil
}

func (u *Sync) Synchronize() error {
	localSums, err := u.GetLocalChecksums()
	if err != nil {
		return err
	}

	remoteSums, err := u.GetRemoteChecksums()
	if err != nil {
		return err
	}

	i := 0
	j := 0
	for i < len(localSums) && j < len(remoteSums) {
		local := &localSums[i]
		remote := &remoteSums[j]

		if local.Name < remote.Name {
			if err := u.SaveLocal(local); err != nil {
				return err
			}
			i++
		} else if local.Name > remote.Name {
			if err := u.DeleteRemote(remote); err != nil {
				return err
			}
			j++
		} else if local.Hash != remote.Hash {
			if err := u.SaveLocal(local); err != nil {
				return err
			}
			i++
			j++
		} else {
			log.Println("=", local.Name)
			i++
			j++
		}
	}

	for i < len(localSums) {
		if err := u.SaveLocal(&localSums[i]); err != nil {
			return err
		}
		i++
	}

	for j < len(remoteSums) {
		if err := u.DeleteRemote(&remoteSums[j]); err != nil {
			return err
		}
		j++
	}

	err = u.StoreChecksums(localSums)
	if err != nil {
		return err
	}

	return nil
}
