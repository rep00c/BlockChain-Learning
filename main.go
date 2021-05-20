package main

import "fmt"

type Test struct {
	a int
	b int
}


func main() {
	a := uint8(128)
	b := uint8(128)
	fmt.Println(a + b)
	//a := make([]uint32, 4)
	//a[0] = 123
	//fmt.Println(a)
}


