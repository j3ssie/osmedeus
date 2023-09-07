package execution

import (
	"fmt"
	"testing"

	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
)

func TestDiffCompare(t *testing.T) {
	var options libs.Options
	options.Debug = true
	DiffCompare("/Users/j3ssie/.osmedeus/workspaces/duckduckgo.com/subdomain/final-duckduckgo.com.txt", "/Users/j3ssie/.osmedeus/storages/summary/duckduckgo.com/subdomain-duckduckgo.com.txt", "/Users/j3ssie/.osmedeus/workspaces/duckduckgo.com/subdomain/diff-duckduckgo.com-2019-12-28_3:20:9.txt", options)
	// DiffCompare(options)
	// fmt.Println(result)

	data := utils.GetFileContent("/Users/j3ssie/.osmedeus/workspaces/duckduckgo.com/subdomain/diff-duckduckgo.com-2019-12-28_3:20:9.txt")
	fmt.Println(data)

	if !utils.FileExists("/Users/j3ssie/.osmedeus/workspaces/duckduckgo.com/subdomain/diff-duckduckgo.com-2019-12-28_3:20:9.txt") {
		t.Logf("Error DiffCompare")
	}

}
