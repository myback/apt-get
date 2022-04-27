package main

import (
	"github.com/myback/apt-get"
)

func main() {
	m, err := apt.Load("source.list", apt.ArchAMD64)
	if err != nil {
		panic(err)
	}

	if err = m.FetchPackage("temp", "nginx-light"); err != nil {
		panic(err)
	}
}
