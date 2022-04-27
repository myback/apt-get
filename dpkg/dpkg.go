package dpkg

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/blakesmith/ar"
	"github.com/myback/go-docker-pull/archive"
	"github.com/xi2/xz"
)

type archType int

const (
	Tar archType = iota
	GZ
	XZ
)

func Unpack(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	extractDir := file[:len(file)-4]

	if err = os.Mkdir(extractDir, 0755); err != nil && os.IsNotExist(err) {
		return err
	}

	arFile := ar.NewReader(f)
	for {
		hdrAr, err := arFile.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		archFilename := strings.Split(filepath.Base(hdrAr.Name), ".")[0]

		var ft archType
		if strings.HasSuffix(hdrAr.Name, ".xz") {
			ft = XZ
		} else if strings.HasSuffix(hdrAr.Name, ".gz") {
			ft = GZ
		} else {
			if err := save(filepath.Join(extractDir, hdrAr.Name), arFile); err != nil {
				return err
			}
			continue
		}

		if err = Decompress(arFile, ft, filepath.Join(extractDir, archFilename)); err != nil {
			return err
		}
	}
}

func Decompress(r io.Reader, fileType archType, extractDir string) (err error) {
	switch fileType {
	case Tar:
	case GZ:
		gz, err := gzip.NewReader(r)
		if err != nil {
			return err
		}
		defer gz.Close()

		r = gz
	case XZ:
		r, err = xz.NewReader(r, 0)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("decompress: unknown archive type")
	}

	return archive.Untar(extractDir, r)
}

func save(file string, data io.Reader) error {
	out, err := os.Create(file)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, data)

	return err
}
