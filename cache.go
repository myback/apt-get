package apt

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"

	"github.com/xi2/xz"
)

const (
	RepoFileName  = "Packages.xz"
	IndexFileName = "Packages"
)

func Load(noCache bool) (*Packages, error) {
	var file io.ReadCloser
	var err error

	if noCache {
		file, err = remote("http://ftp.debian.org/debian/dists/bullseye/main/binary-amd64/Packages.gz")
	} else {
		file, err = os.Open(IndexFileName)
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	var m map[string]Package
	if noCache {
		m, err = parse(gzReader)
		if err != nil {
			return nil, err
		}
	} else {
		err = json.NewDecoder(gzReader).Decode(&m)
		if err != nil {
			return nil, err
		}
	}

	return &Packages{info: m}, nil
}

func Update() error {
	body, err := remote("http://ftp.debian.org/debian/dists/bullseye/main/binary-amd64/Packages.xz")
	if err != nil {
		return err
	}
	defer body.Close()

	tmpFile, err := ioutil.TempFile("", "apt-package-")
	if err != nil {
		return err
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	xzfile, err := xz.NewReader(body, 0)
	if err != nil {
		return err
	}

	if _, err := io.CopyBuffer(tmpFile, xzfile, make([]byte, 1048576)); err != nil {
		return err
	}

	if _, err := tmpFile.Seek(0, 0); err != nil {
		return err
	}

	m, err := parse(tmpFile)
	if err != nil {
		return err
	}

	indexFile, err := os.Create(IndexFileName)
	if err != nil {
		return err
	}
	defer indexFile.Close()

	gzipWriter, err := gzip.NewWriterLevel(indexFile, gzip.BestSpeed)
	if err != nil {
		return err
	}
	defer gzipWriter.Close()

	if err := json.NewEncoder(gzipWriter).Encode(m); err != nil {
		return err
	}

	return nil
}
