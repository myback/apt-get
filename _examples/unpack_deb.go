package main

import (
	"github.com/myback/apt-get/dpkg"
)

func main() {
	if err := dpkg.Unpack("temp/nginx-light_1.18.0-6.1_amd64.deb"); err != nil {
		panic(err)
	}
}
