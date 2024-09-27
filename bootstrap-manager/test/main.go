package main

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Test ...
type Test struct {
	Yaddas []*Yadda `validate:"required,dive`
}

type Yadda struct {
	Name string `validate:"required"`
}

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

func main() {

	validate = validator.New()

	var test Test

	yadda := &Yadda{
		Name: "ish",
	}
	test.Yaddas = append(test.Yaddas, yadda)

	val(test)
}

func val(test Test) {
	fmt.Println("testing")
	err := validate.Struct(test)
	fmt.Println(err)
}
