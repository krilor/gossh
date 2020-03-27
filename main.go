package main

import (
	"flag"
	"fmt"

	"github.com/krilor/gossh/state"

	"plugin"
)

type pluginFunc func() (state.State, error)

func main() {

	recipeName := flag.String("recipeName", "", "Name of the .so file")
	flag.Parse()
	p, err := plugin.Open(*recipeName)

	if err != nil {
		panic(err) //if recipeName is empty, this runs. Better error message can be added
	}
	getState, err := p.Lookup("GetState") //always look for GetState function
	if err != nil {
		panic(err)
	}

	newState, err := getState.(pluginFunc)() //typecast received function & call it

	//err = newState.Apply()
	fmt.Printf("%+v", newState)
	if err != nil {
		fmt.Println("apply gone wrong", err)
	}

}
