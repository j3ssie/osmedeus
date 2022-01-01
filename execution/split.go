package execution

import (
    "fmt"
    "github.com/j3ssie/osmedeus/utils"
    "github.com/robertkrimen/otto"
    "os"
    "path"
    "path/filepath"
    "strings"
)

/*

	SplitFile("src", "dest")
 	SplitFile("src", "base/dest", 500, "another/dest")
 	next step should be loop in file base/index

*/

// SplitFile split file into multiple file
func SplitFile(kind string, arguments []otto.Value) {
    source := arguments[0].String()
    dest := arguments[1].String()
    // prefix name of the output file
    prefix := fmt.Sprintf("%v-chunked", filepath.Base(source))
    destDir := filepath.Dir(utils.NormalizePath(dest))
    if len(arguments) >= 4 {
        destDir = arguments[3].String()
    }
    utils.MakeDir(destDir)

    chunk := 200
    // check if we change the size of it
    if len(arguments) >= 3 {
        chunkSize, _ := arguments[2].ToInteger()
        chunk = int(chunkSize)
    }

    chunkParts := chunk
    // get number of part
    if kind == "size" {
        length := utils.FileLength(source)
        chunkParts = length / chunk
    }

    utils.DebugF("Splitting %v to %v %v", source, chunkParts, kind)
    rawChunks, err := utils.SplitLineChunks(source, chunkParts)
    if err != nil || len(rawChunks) == 0 {
        utils.ErrorF("error to split input file: %v", source)
        return
    }
    fp, err := os.Open(source)
    if err != nil {
        utils.ErrorF("error to open input file: %v", source)
        return
    }

    var sumFile []string
    for index, offset := range rawChunks {
        targetName := path.Join(destDir, fmt.Sprintf("%v-%v", prefix, index))
        reader := utils.NewRangeReader(fp, offset.Start, offset.Stop)
        body := make([]byte, offset.Stop-offset.Start+1)
        _, err := reader.Read(body)
        if err != nil {
            utils.ErrorF("error to read chunk file: %s", err)
            continue
        }
        sumFile = append(sumFile, targetName)
        utils.DebugF("writing %v part to: %v", index, targetName)
        utils.WriteToFile(targetName, string(body))
    }

    // write summary file
    indexFile := path.Join(destDir, dest)
    _, err = utils.WriteToFile(indexFile, strings.Join(sumFile, "\n"))
    if err != nil {
        utils.ErrorF("Error writing to %v", indexFile)
    }
}

// ChunkFileByPart chunk file to multiple part
func ChunkFileByPart(source string, chunk int) [][]string {
    var divided [][]string
    data := utils.ReadingLines(source)
    if len(data) <= 0 || chunk > len(data) {
        if len(data) > 0 {
            divided = append(divided, data)
        }
        return divided
    }

    chunkSize := (len(data) + chunk - 1) / chunk
    for i := 0; i < len(data); i += chunkSize {
        end := i + chunkSize
        if end > len(data) {
            end = len(data)
        }

        divided = append(divided, data[i:end])
    }
    return divided
}

// ChunkFileBySize chunk file to multiple part
func ChunkFileBySize(source string, chunk int) [][]string {
    var divided [][]string
    data := utils.ReadingLines(source)
    if len(data) <= 0 || chunk > len(data) {
        if len(data) > 0 {
            divided = append(divided, data)
        }
        return divided
    }

    chunkSize := chunk
    for i := 0; i < len(data); i += chunkSize {
        end := i + chunkSize
        if end > len(data) {
            end = len(data)
        }

        divided = append(divided, data[i:end])
    }
    return divided
}
