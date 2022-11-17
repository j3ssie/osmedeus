package core

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/go-playground/validator/v10"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"path"
	"strings"
)

func (r *Runner) Validator() error {
	if r.RequiredInput == "" || r.Opt.DisableValidateInput {
		return nil
	}

	r.RequiredInput = strings.ToLower(strings.TrimSpace(r.RequiredInput))
	//if r.RequiredInput == "file" {
	//	if utils.FileExists(r.Input) {
	//		return nil
	//	}
	//}

	var inputAsFile bool
	// cidr, cidr-file
	if strings.HasSuffix(r.RequiredInput, "-file") || r.RequiredInput == "file" {
		inputAsFile = true
	}
	v := validator.New()

	// if input as a file
	if utils.FileExists(r.Input) && inputAsFile {
		r.InputType = "file"
		inputs := utils.ReadingLines(r.Input)

		for index, input := range inputs {
			if strings.TrimSpace(input) == "" {
				continue
			}

			inputType, err := validate(v, input)
			if err == nil {
				// really validate thing
				if !strings.HasPrefix(r.RequiredInput, inputType) {
					utils.DebugF("validate: %v -- %v", input, inputType)
					errString := fmt.Sprintf("line %v in %v file not match the require input: %v -- %v", index, r.Input, input, inputType)
					utils.ErrorF(errString)
					return fmt.Errorf(errString)
				}
			}
		}
		return nil

	}

	utils.InforF("Start validating input: %v", color.HiCyanString("%v -- %v", r.Input, r.InputType))
	var err error
	r.InputType, err = validate(v, r.Input)
	if err != nil {
		utils.ErrorF("unrecognized input: %v", r.Input)
		return err
	}

	if !strings.HasPrefix(r.RequiredInput, r.InputType) {
		return fmt.Errorf("input does not match the require validation: inputType:%v -- requireType:%v", r.InputType, r.RequiredInput)
	}

	if inputAsFile {
		utils.MakeDir(libs.TEMP)
		dest := path.Join(libs.TEMP, fmt.Sprintf("%v-%v", utils.StripPath(r.Input), utils.RandomString(4)))
		if r.Opt.Scan.CustomWorkspace != "" {
			dest = path.Join(libs.TEMP, fmt.Sprintf("%v-%v", utils.StripPath(r.Opt.Scan.CustomWorkspace), utils.RandomString(4)))
		}
		utils.WriteToFile(dest, r.Input)
		utils.InforF("Convert input to a file: %v", dest)
		r.Input = dest
		r.Target = ParseInput(r.Input, r.Opt)
	}

	utils.DebugF("validator: input:%v -- type: %v -- require:%v", r.Input, r.InputType, r.RequiredInput)
	return nil
}

func validate(v *validator.Validate, raw string) (string, error) {
	var err error
	var inputType string

	if utils.FileExists(raw) {
		inputType = "file"
	}

	err = v.Var(raw, "required,url")
	if err == nil {
		inputType = "url"
	}

	err = v.Var(raw, "required,ipv4")
	if err == nil {
		inputType = "ip"
	}

	err = v.Var(raw, "required,fqdn")
	if err == nil {
		inputType = "domain"
	}

	err = v.Var(raw, "required,hostname")
	if err == nil {
		inputType = "domain"
	}

	err = v.Var(raw, "required,cidr")
	if err == nil {
		inputType = "cidr"
	}

	err = v.Var(raw, "required,uri")
	if err == nil {
		inputType = "url"
	}

	err = v.Var(raw, "required,uri")
	if err == nil {
		inputType = "url"
		if strings.HasPrefix(raw, "https://github.com") || strings.HasPrefix(raw, "https://gitlab.com") {
			inputType = "git-url"
		}
	}

	if strings.HasPrefix(raw, "git@") {
		inputType = "git-url"
	}

	if inputType == "" {
		return "", fmt.Errorf("unrecognized input")
	}

	return inputType, nil
}
