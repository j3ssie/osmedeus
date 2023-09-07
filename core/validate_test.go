package core

import (
	"fmt"
	"github.com/j3ssie/osmedeus/libs"
	"testing"
)

func TestValidator(t *testing.T) {
	input := "target.com"
	opt := libs.Options{}
	runner, _ := InitRunner(input, opt)
	runner.RequiredInput = "domain"
	runner.Validator()

	runner.Input = "apple.com"
	runner.Validator()

	fmt.Printf("runner.InputType --> %v:%v -- %s\n", runner.RequiredInput, runner.InputType, runner.Input)

	runner.Input = "1.2.3.4"
	runner.Validator()

	runner.Input = "http://127.0.0.1/q"
	runner.Validator()
	runner.Input = "sub.domain.com"
	runner.Validator()

	runner.Input = "1.2.3.4/24"
	runner.Validator()

	runner.Input = "https://github.com/j3ssie/osmedeus"
	runner.Validator()
	fmt.Printf("==> runner.InputType --> %v:%v -- %s\n\n", runner.RequiredInput, runner.InputType, runner.Input)

	runner.Input = "git@github.com:j3ssie/osmedeus.git"
	runner.Validator()
	fmt.Printf("==> runner.InputType --> %v:%v -- %s\n\n", runner.RequiredInput, runner.InputType, runner.Input)

	//
	////raw := "tcp://git@gitlab.com:j3ssie/osmd-assets"
	//raw := "git@gitlab.com/j3ssie/osmd-assets"
	//v := validator.New()
	//err := v.Var(raw, "required,uri")
	//fmt.Println(err)
	//
	//err = v.Var(raw, "required,datauri")
	//fmt.Println(err)

	if runner.InputType == "" {
		t.Errorf("Error Validator")
	}
}
