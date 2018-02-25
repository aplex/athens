package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/arschles/vgoprox/pkg/payloads"
)

const help = `Usage:
vgp <basepath> <module> <version>

Details:

- The directory from which code will be uploaded is <basepath>/<module>
- ... and that directory must have a go.mod file in ot
- ... and if there's a vendor directory under that directory, it won't be ignored right now
- ... and the go.mod file will be uploaded with the source
`

func main() {
	if len(os.Args) != 4 {
		log.Println(help)
		os.Exit(1)
	}

	basePath := os.Args[1]
	module := os.Args[2]
	version := os.Args[3]

	fullDirectory, err := filepath.Abs(filepath.Join(basePath, module))
	if err != nil {
		log.Fatalf("couldn't get full directory (%s)", err)
	}
	log.Printf("found directory %s", fullDirectory)
	modFilePath := filepath.Join(fullDirectory, "go.mod")
	log.Printf("reading %s file", modFilePath)
	modBytes, err := ioutil.ReadFile(modFilePath)
	if err != nil {
		log.Fatalf("couldn't find go.mod file (%s)", err)
	}

	zipBytes, err := makeZip(fullDirectory)
	if err != nil {
		log.Fatalf("couldn't make zip (%s)", err)
	}

	url := fmt.Sprintf("http://localhost:3000/admin/upload/%s/%s/%s", basePath, module, version)
	postBody := &payloads.Upload{
		Module: modBytes,
		Zip:    zipBytes,
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(postBody); err != nil {
		log.Fatalf("error encoding json (%s)", err)
	}
	resp, err := http.Post(url, "application/json", buf)
	if err != nil {
		log.Fatalf("error uploading (%s)", err)
	} else if resp.StatusCode != 200 {
		log.Fatalf("upload failed because status code was %d", resp.StatusCode)
	}
}
