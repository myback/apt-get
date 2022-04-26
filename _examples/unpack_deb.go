package main

import (
	"os"

	"github.com/myback/apt-get/dpkg"
)

func main() {
	file := "temp/nginx-light_1.18.0-6.1_amd64.deb"
	f, _ := os.Open(file)
	defer f.Close()

	if err := dpkg.Unpack(f, file[:len(file)-4]); err != nil {
		panic(err)
	}
}
