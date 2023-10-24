package utils

import (
	"fmt"
	"testing"
)

func TestTar(t *testing.T) {
	err := MakeTar("../static", "./static.tar")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = UnpackTar("./static.tar", "./OK", true)
	if err != nil {
		fmt.Println(err)
		return
	}
}
