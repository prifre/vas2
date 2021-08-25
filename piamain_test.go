package main

import (
	"fmt"
	"testing"
)

func TestGetint64(t *testing.T) {
	got := Getint64("123")
	if got == 123 {
		fmt.Println("Getint64 ok")
	} else {
		fmt.Println("Getint64 failed")
	}
}

func (g *game) TestCheckforupdate(t *testing.T) {
	fmt.Println("!!!!!!!!!!!!")
	g.Checkforupdate()
}

