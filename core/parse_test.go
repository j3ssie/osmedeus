package core

import (
	"fmt"
	"github.com/flosch/pongo2/v6"
	"runtime"
	"testing"
)

func TestParseTarget(t *testing.T) {
	fmt.Println("----> INPUT:", "http://example.com")
	result := ParseTarget("http://exmaple.com")
	fmt.Println(result)
	if len(result) == 0 {
		t.Errorf("Error RunMasscan")
	}
	// case 2
	fmt.Println("----> INPUT:", "example.com")
	result = ParseTarget("exmaple.com")
	fmt.Println(result)
	if len(result) == 0 {
		t.Errorf("Error RunMasscan")
	}

	// case 2
	fmt.Println("----> INPUT:", "http://exmaple.com/123?q=1")
	result = ParseTarget("http://exmaple.com/123?q=1")
	fmt.Println(result)
	if len(result) == 0 {
		t.Errorf("Error RunMasscan")
	}

	// case 2
	fmt.Println("----> INPUT:", "1.2.3.4")
	result = ParseTarget("1.2.3.4")
	fmt.Println(result)
	if len(result) == 0 {
		t.Errorf("Error RunMasscan")
	}

	// case 2
	fmt.Println("----> INPUT:", "1.2.3.4/24")
	result = ParseTarget("1.2.3.4/24")
	fmt.Println(result)
	if len(result) == 0 {
		t.Errorf("Error RunMasscan")
	}
}

func TestRenderTemplate(t *testing.T) {
	formatString := "Hello {{name}}! \n"
	formatString += "--> Express: {{ 2 * cpu }} \n"
	formatString += "--> Bool: {{ Skip }} \n"
	target := map[string]any{
		"name": "World",
		"num":  "2",
		"cpu":  runtime.NumCPU(),
		"Skip": true,
	}
	fmt.Println(target)

	if tpl, err := pongo2.FromString(formatString); err == nil {
		// Now you can render the template with the given
		// pongo2.Context how often you want to.
		out, err := tpl.Execute(target)
		if err != nil {
			panic(err)
		}
		fmt.Println(out) // Output: Hello Florian!
	}
}
