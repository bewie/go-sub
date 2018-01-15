package downloader

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"os"
)

// Get ...
func (q *Query) Get(url, destFile string) (err error) {
	response, err := http.Get(url)
	if err == nil {
		defer response.Body.Close()
		if err == nil {
			dir, err := ioutil.TempDir("", "")
			if err == nil {
				defer os.RemoveAll(dir) // clean up
				if reader, err := gzip.NewReader(response.Body); err == nil {
					if b, err := ioutil.ReadAll(reader); err == nil {
						err = ioutil.WriteFile(destFile, b, 0644)
					}
				}
			}
		}
	}
	return err
}
