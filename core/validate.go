package core

import (
    "fmt"
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

    // cidr, cidr-file
    r.RequiredInput = strings.ToLower(strings.TrimSpace(r.RequiredInput))
    var inputAsFile bool
    if strings.HasSuffix(r.RequiredInput, "-file") {
        inputAsFile = true
    }
    v := validator.New()

    // if input as a file
    if utils.FileExists(r.Input) && inputAsFile {
        //if !inputAsFile {
        //    utils.ErrorF("input required is not a file: %v", r.Input)
        //    return fmt.Errorf("input required is not a file: %v", r.Input)
        //}

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

    var err error
    r.InputType, err = validate(v, r.Input)
    if err != nil {
        utils.ErrorF("unrecognized input: %v", r.Input)
        return err
    }

    if !strings.HasPrefix(r.RequiredInput, r.InputType) {
        return fmt.Errorf("input does not match the require validation: inputType:%v -- requireType:%v", r.InputType, r.RequiredInput)
    }

    utils.InforF("Start validating input: %v -- %v", r.Input, r.InputType)
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

    if inputType == "" {
        return "", fmt.Errorf("unrecognized input")
    }

    return inputType, nil
}
