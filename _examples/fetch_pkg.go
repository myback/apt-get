package main

import (
	"github.com/myback/apt-get"
)

func main() {
	if err := apt.Update(); err != nil {
		panic(err)
	}

	m, err := apt.Load(false)
	if err != nil {
		panic(err)
	}

	if err = m.FetchPackage("nginx-light", "temp"); err != nil {
		panic(err)
	}
}
