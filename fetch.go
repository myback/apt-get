package apt

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

func fetchPackageList(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "plain/text")

	client := &http.Client{Timeout: 3 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		respContent, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%d: %s", resp.StatusCode, respContent)
	}

	return resp.Body, nil
}
