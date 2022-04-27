package apt

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/xi2/xz"
)

const (
	PackageFileName = "Packages.xz"
	PackageFileExt  = ".xz"
	CacheDir        = "apt_list"

	ArchX86    = "i386"
	ArchAMD64  = "amd64"
	AchARM64   = "arm64"
	ArchARMHF  = "armhf"
	ArchPPC64  = "ppc64el"
	ArchRisc64 = "riscv64"
	ArchS390x  = "s390x"
)

func Load(sourceListFile, arch string) (*Packages, error) {
	lst, err := list(sourceListFile, arch)
	if err != nil {
		return nil, err
	}

	pkgs := new(Packages)
	for _, repo := range lst {
		pkgList, err := load(repo[0])
		if err != nil {
			return nil, err
		}

		if len(pkgs.info) == 0 {
			pkgs.info = pkgList
			continue
		}

		for k, v := range pkgList {
			pkgs.info[k] = v
		}
	}

	return pkgs, nil
}

func load(fileName string) (map[string]Package, error) {
	file, err := os.Open(filepath.Join(CacheDir, fileName+PackageFileExt))
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
	if err = json.NewDecoder(gzReader).Decode(&m); err != nil {
		return nil, err
	}

	return m, nil
}

func fileNameGenerator(repoAddr, dist, channel, arch string) (string, string, error) {
	u, err := url.Parse(repoAddr)
	if err != nil {
		return "", "", err
	}

	u.Path = path.Join(u.Path, "dists", dist, channel, "binary-"+arch, PackageFileName)

	host := strings.Split(u.Host, ":")[0]
	distSlug := strings.Replace(dist, "/", "-", -1)
	tmplFileName := fmt.Sprintf("%s_%s_%s_%s", host, distSlug, channel, arch)

	return tmplFileName, u.String(), nil
}

func list(sourceListFile, arch string) ([][2]string, error) {
	out := make([][2]string, 0)

	if len(arch) == 0 {
		arch = ArchAMD64
	}
	b, err := ioutil.ReadFile(sourceListFile)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(b), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || len(line) == 0 {
			continue
		}

		repoData := strings.Split(line, " ")

		if repoData[0] != "deb" {
			fmt.Printf("WARN: skip repo %s. Repo must start from deb\n", line)
			continue
		}

		tpl, u, err := fileNameGenerator(repoData[1], repoData[2], repoData[3], arch)
		if err != nil {
			return nil, err
		}

		out = append(out, [2]string{tpl, u})
	}

	return out, nil
}

func Update(sourceListFile, arch string) error {
	lst, err := list(sourceListFile, arch)
	if err != nil {
		return err
	}

	for _, repo := range lst {
		if err = update(repo[0], repo[1]); err != nil {
			return err
		}
	}

	return nil
}

func update(tmplFileName, url string) error {

	if _, err := os.Stat(CacheDir); err != nil {
		if err = os.MkdirAll(CacheDir, 0755); err != nil {
			return err
		}
	}

	tmpFile, err := ioutil.TempFile("", tmplFileName+"-")
	if err != nil {
		return err
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	body, err := fetchPackageList(url)
	if err != nil {
		return err
	}
	defer body.Close()

	xzfile, err := xz.NewReader(body, 0)
	if err != nil {
		return err
	}

	if _, err = io.CopyBuffer(tmpFile, xzfile, make([]byte, 1048576)); err != nil {
		return err
	}

	if _, err = tmpFile.Seek(0, 0); err != nil {
		return err
	}

	indexFile, err := os.Create(filepath.Join(CacheDir, tmplFileName+PackageFileExt))
	if err != nil {
		return err
	}
	defer indexFile.Close()

	gzipWriter, err := gzip.NewWriterLevel(indexFile, gzip.BestSpeed)
	if err != nil {
		return err
	}
	defer gzipWriter.Close()

	m, err := parse(tmpFile)
	if err != nil {
		return err
	}

	return json.NewEncoder(gzipWriter).Encode(m)
}
