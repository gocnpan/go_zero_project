package main

import "fmt"

func main()  {
	var loaders = map[string]func(){
		".one": mpOne,
		".two": mpTwo,
	}
	loaderTwo, _ := loaders[".two"]
	loaderTwo()
	loaderOne, _ := loaders[".one"]
	loaderOne()
}

func mpOne()  {
	fmt.Println("map one")
}

func mpTwo()  {
	fmt.Println("map two")
}