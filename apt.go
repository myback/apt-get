package apt

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Packages struct {
	info     map[string]Package
	depsList []string
}

type Package struct {
	Depends  string `yaml:"Depends" json:"d,omitempty"`
	Filename string `yaml:"Filename" json:"f"`
	Package  string `yaml:"Package" json:"-"`
	Version  string `yaml:"Version" json:"v"`
}

func (p *Packages) orDeps(d string) string {
	var deps []string
	for _, orD := range strings.Split(d, "|") {
		pkg := strings.TrimSpace(orD)
		idx := strings.IndexByte(pkg, ' ')
		if idx > 0 {
			pkg = pkg[:idx]
		}

		if InSlice(p.depsList, pkg) {
			return pkg
		}

		deps = append(deps, pkg)
	}

	for _, d := range deps {
		if _, ok := p.info[d]; ok {
			return d
		}
	}

	return ""
}

func (p *Packages) packageDependencies() {
	for _, dep := range p.depsList {
		pkg := p.info[dep]

		var deps []string
		for _, d := range strings.Split(pkg.Depends, ",") {
			var pkg string
			if strings.IndexByte(d, '|') > -1 {
				pkg = p.orDeps(d)
			} else {
				pkg = strings.TrimSpace(d)
				idx := strings.IndexByte(pkg, ' ')
				if idx > 0 {
					pkg = pkg[:idx]
				}
			}

			if InSlice(p.depsList, pkg) {
				continue
			}

			if _, ok := p.info[pkg]; ok {
				deps = append(deps, pkg)
			}
		}

		if len(deps) > 0 {
			p.depsList = append(p.depsList, deps...)
			p.packageDependencies()
		}
	}
}

func (p *Packages) GetPackagesDependency(pkgs ...string) ([]string, error) {
	if len(pkgs) == 0 {
		return []string{}, nil
	}

	for _, pkg := range pkgs {
		_, ok := p.info[pkg]
		if !ok {
			return nil, fmt.Errorf("%s: package not found", p)
		}

		p.depsList = append(p.depsList, pkg)
	}

	p.packageDependencies()
	sort.Strings(p.depsList)

	return p.depsList, nil
}

func (p *Packages) FetchPackage(outDir, pkg string) error {
	pkgInfo, ok := p.info[pkg]
	if !ok {
		return fmt.Errorf("package %s not found", pkg)
	}

	if _, err := os.Stat(outDir); err != nil {
		if err = os.MkdirAll(outDir, 0755); err != nil {
			return err
		}
	}

	body, err := fetchPackageList("http://ftp.debian.org/debian/" + pkgInfo.Filename)
	if err != nil {
		return err
	}
	defer body.Close()

	filename := filepath.Base(pkgInfo.Filename)
	outFile, err := os.Create(filepath.Join(outDir, filename))
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, body)
	return err
}
