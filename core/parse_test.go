package core

import (
	"fmt"
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
