package service

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

func getKeys () map[string]string {

	usr, err := user.Current()
	if err != nil {
		log.Fatal( err )
	}

	keys := findKeys(usr.HomeDir + "/keys", ".pem")
	if len(keys) == 0 {
		keys = findKeys(usr.HomeDir, ".pem")
	}

	return keys
}

func findKeys(root string, ext string) map[string]string {
	var files = make(map[string]string, 0)
	filepath.Walk(root, func(path string, f os.FileInfo, _ error) error {
		if !strings.HasPrefix(f.Name(), ".") {
			if !f.IsDir() {
				r, err := regexp.MatchString(ext, f.Name())
				if err == nil && r {
					if _, ok := files[f.Name()]; !ok {
						files[f.Name()] = path
					}
				}
			}
		}
		return nil
	})
	return files
}