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

func NewSync(b backend.Backend, chksumName, sourceDir string) (*Sync, error) {
	return &Sync{
		backend:    b,
		chksumName: chksumName,
		sourceDir:  sourceDir,
	}, nil
}

type Sync struct {
	backend    backend.Backend
	chksumName string
	sourceDir  string
}

func (u *Sync) GetRemoteChecksums() ([]Element, error) {
	elements := []Element{}

	res, err := u.backend.Fetch(u.chksumName)
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

	err := u.backend.Store(u.chksumName, &b)
	if err != nil {
		return err
	}

	return nil
}

func (u *Sync) SaveLocal(element *Element) error {
	if element.isDir() {
		log.Println("d", element.Name)
		_ = u.backend.MakeDir(element.Name)
		return nil
	} else {
		log.Println("+", element.Name)
		file, err := os.Open(path.Join(u.sourceDir, element.Name))
		if err != nil {
			return err
		}
		defer file.Close()
		return u.backend.Store(element.Name, file)
	}
}

func (u *Sync) DeleteRemote(remote *Element) error {
	log.Println("-", remote.Name)
	return u.backend.Delete(remote.Name)
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

	j := 0
	for _, local := range localSums {
		if j >= len(remoteSums) {
			// save the local file or create the directory
			if err := u.SaveLocal(&local); err != nil {
				return err
			}
			continue
		}

		remote := remoteSums[j]
		for local.Name > remote.Name {
			// delete the remote file
			if remote.Hash == "" {
				// do not remove the directory
			} else if err := u.DeleteRemote(&remote); err != nil {
				return err
			}
			j++
			remote = remoteSums[j]
		}

		if local.Name < remote.Name {
			// save the local file or create the directory
			if err := u.SaveLocal(&local); err != nil {
				return err
			}
		} else if local.Hash != remote.Hash {
			// save the local file or create the directory
			if err := u.SaveLocal(&local); err != nil {
				return err
			}
			j++
		} else {
			// same contents
			log.Println("=", local.Name)
			j++
		}
	}

	err = u.StoreChecksums(localSums)
	if err != nil {
		return err
	}

	return nil
}
