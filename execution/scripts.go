package execution

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/j3ssie/osmedeus/libs"

	"github.com/j3ssie/osmedeus/utils"
)

// Execution Run a command
func Execution(cmd string, options libs.Options) (string, error) {
	command := []string{
		"bash",
		"-c",
		cmd,
	}
	//var output string
	utils.DebugF("[Exec] %v", command)
	realCmd := exec.Command(command[0], command[1:]...)
	if options.Quite == true {
		realCmd.CombinedOutput()
	} else {
		// output command output to std too
		cmdReader, _ := realCmd.StdoutPipe()
		scanner := bufio.NewScanner(cmdReader)
		//var out string
		go func() {
			for scanner.Scan() {
				utils.InforF(scanner.Text())
			}
		}()
		if err := realCmd.Start(); err != nil {
			return "", err
		}
		if err := realCmd.Wait(); err != nil {
			return "", err
		}
	}
	return "", nil
}

// Echo just print testing
func Echo(script string) {
	fmt.Println("testing:", script)
}

// Sleep just print testing
func Sleep(raw string) {
	time.Sleep(time.Second * time.Duration(utils.StrToInt(raw)))
}

// Printf print out some string
func Printf(block string, content string) {
	if block != "" {
		utils.BlockF(block, content)
	} else {
		fmt.Println(content)
	}
}

// ErrPrintf print out some string
func ErrPrintf(block string, content string) {
	if block != "" {
		utils.BadBlockF(block, content)
	} else {
		fmt.Println(content)
	}
}

// Base just print testing
func Base(raw string) string {
	raw = utils.NormalizePath(raw)
	return path.Base(raw)
}

// StripName strip a file name
func StripName(raw string) string {
	result := strings.Replace(raw, "/", "_", -1)
	result = strings.Replace(result, ":", "_", -1)
	return strings.Trim(result, "_")
}

// DeleteFile delete file
func DeleteFile(filename string) {
	utils.DebugF("Delete: %v", filename)
	os.Remove(utils.NormalizePath(filename))
}

// DeleteFolder delete entire folder
func DeleteFolder(path string) {
	utils.DebugF("Delete: %v", path)
	os.RemoveAll(utils.NormalizePath(path))
}

// Append append content to a file
func Append(dest string, src string) {
	if !utils.FileExists(src) || utils.FileLength(src) <= 0 {
		utils.DebugF("error to append %v", src)
		return
	}
	data := utils.GetFileContent(src)
	utils.AppendToContent(dest, data)
}

// Copy append content to a file
func Copy(src string, dest string) {
	if !utils.FileExists(src) || utils.FileLength(src) <= 0 {
		return
	}
	input, _ := ioutil.ReadFile(src)
	ioutil.WriteFile(dest, input, 0644)
}

// Sort sort content of a file
func Sort(src string) {
	data := utils.ReadingFileUnique(src)
	if len(data) == 0 {
		return
	}
	sort.Strings(data)
	content := strings.Join(data, "\n")
	// remove blank line
	content = regexp.MustCompile(`[\t\r\n]+`).ReplaceAllString(strings.TrimSpace(content), "\n")
	utils.WriteToFile(src, content)
}

// SortU sort content of a file
func SortU(src string) {
	if !utils.FileExists(src) {
		utils.DebugF("File not found: %s", src)
	}
	cmd := fmt.Sprintf("LC_ALL=C sort -u -o %s %s", src, src)
	utils.RunCmdWithOutput(cmd)
}

// Unique unique content of a file and remove blank line
func Unique(filename string) {
	data := utils.ReadingFileUnique(filename)
	if len(data) > 0 {
		content := strings.Join(data, "\n")
		// remove blank line
		content = regexp.MustCompile(`[\t\r\n]+`).ReplaceAllString(strings.TrimSpace(content), "\n")
		utils.WriteToFile(filename, content)
	}
}

// Compress sort content of a file
func Compress(dest string, src string) {
	if utils.FolderLength(src) < 1 {
		utils.DebugF("Source folder is empty or not found: %s", src)
		return
	}
	cmd := fmt.Sprintf("tar --use-compress-program='gzip -9' -C %s -cf %s .", strings.TrimRight(src, "/"), dest)
	utils.RunCmdWithOutput(cmd)
}

// Decompress sort content of a file
func Decompress(dest string, src string) {
	if !utils.FileExists(src) {
		utils.DebugF("File not found: %s", src)
		return
	}
	utils.MakeDir(dest)
	cmd := fmt.Sprintf("tar -xf %s -C %s", src, dest)
	utils.RunCmdWithOutput(cmd)
}

func ExtractTarGz(filename string) error {
	r, err := os.Open(utils.NormalizePath(filename))
	if err != nil {
		return err
	}

	uncompressedStream, err := gzip.NewReader(r)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(uncompressedStream)
	for true {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(header.Name, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(header.Name)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
			outFile.Close()

		default:
			return err

		}

	}

	return nil
}
