package core

import (
	"fmt"
	"testing"

	"github.com/j3ssie/osmedeus/libs"
)

func TestListMode(t *testing.T) {
	var options libs.Options
	options.Env.WorkFlowsFolder = "~/go/src/github.com/j3ssie/osmedeus/workflow/"
	result := ListFlow(options)
	fmt.Println(result)
	if len(result) == 0 {
		t.Errorf("Error ListMode")
	}

	selectedMode := SelectFlow("general", options)
	fmt.Println(selectedMode)
	if len(selectedMode) == 0 {
		t.Errorf("Error selectedMode")
	}
}
