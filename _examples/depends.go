package main

import (
	"fmt"
	"github.com/myback/apt-get"
)

func main() {
	if err := apt.Update("source.list", apt.ArchAMD64); err != nil {
		panic(err)
	}

	m, err := apt.Load("source.list", apt.ArchAMD64)
	if err != nil {
		panic(err)
	}

	pkgs, err := m.GetPackagesDependency("openssh-server")
	if err != nil {
		panic(err)
	}

	for _, pkg := range pkgs {
		fmt.Println(pkg)
	}
}
