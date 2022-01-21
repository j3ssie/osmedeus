package utils

import (
    "archive/zip"
    "bufio"
    "context"
    "crypto/sha1"
    "encoding/base64"
    "fmt"
    "io"
    "io/ioutil"
    "math/rand"
    "net/url"
    "os"
    "os/exec"
    "path"
    "path/filepath"
    "regexp"
    "strconv"
    "strings"
    "syscall"
    "time"

    "github.com/mitchellh/go-homedir"
)

// CalcTimeout calculate timeout
func CalcTimeout(raw string) int {
    raw = strings.ToLower(strings.TrimSpace(raw))
    seconds := raw
    multiply := 1

    matched, _ := regexp.MatchString(`.*[a-z]`, raw)
    if matched {
        unitTime := fmt.Sprintf("%c", raw[len(raw)-1])
        seconds = raw[:len(raw)-1]
        switch unitTime {
        case "s":
            multiply = 1
            break
        case "m":
            multiply = 60
            break
        case "h":
            multiply = 3600
            break
        }
    }

    timeout, err := strconv.Atoi(seconds)
    if err != nil {
        return 0
    }
    return timeout * multiply
}

// GetDomain get domain from the URL
func GetDomain(raw string) (string, error) {
    u, err := url.Parse(raw)
    if err == nil {
        return u.Hostname(), nil
    }
    return raw, err
}

// EmptyDir check if directory is empty or not
func EmptyDir(dir string) bool {
    if !FolderExists(dir) {
        return true
    }
    f, err := os.Open(NormalizePath(dir))
    if err != nil {
        return false
    }
    defer f.Close()

    _, err = f.Readdirnames(1)
    if err == io.EOF {
        return true
    }
    return false
}

// EmptyFile check if file is empty or not
func EmptyFile(filename string, num int) bool {
    filename = NormalizePath(filename)
    if !FileExists(filename) {
        return true
    }
    data := ReadingLines(filename)
    if len(data) > num {
        return false
    }
    return true
}

// StrToInt string to int
func StrToInt(data string) int {
    i, err := strconv.Atoi(data)
    if err != nil {
        return 0
    }
    return i
}

// GetOSEnv get environment variable
func GetOSEnv(name string, alt string) string {
    variable, ok := os.LookupEnv(name)
    if !ok {
        if alt != "" {
            return alt
        }
        return name
    }
    return variable
}

// MakeDir just make a folder
func MakeDir(folder string) {
    folder = NormalizePath(folder)
    os.MkdirAll(folder, 0750)
}

// GetCurrentDay get current day
func GetCurrentDay() string {
    currentTime := time.Now()
    return fmt.Sprintf("%v", currentTime.Format("2006-01-02_3:4:5"))
}

// NormalizePath the path
func NormalizePath(path string) string {
    if strings.HasPrefix(path, "~") {
        path, _ = homedir.Expand(path)
    }
    return path
}

// GetFileContent Reading file and return content of it
func GetFileContent(filename string) string {
    var result string
    if strings.Contains(filename, "~") {
        filename, _ = homedir.Expand(filename)
    }
    file, err := os.Open(filename)
    if err != nil {
        return result
    }
    defer file.Close()
    b, err := ioutil.ReadAll(file)
    if err != nil {
        return result
    }
    return string(b)
}

// ReadingLines Reading file and return content as []string
func ReadingLines(filename string) []string {
    var result []string
    if strings.HasPrefix(filename, "~") {
        filename, _ = homedir.Expand(filename)
    }
    file, err := os.Open(filename)
    if err != nil {
        return result
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        val := strings.TrimSpace(scanner.Text())
        if val == "" {
            continue
        }
        result = append(result, val)
    }

    if err := scanner.Err(); err != nil {
        return result
    }
    return result
}

// Cat Reading file and return content as []string
func Cat(filename string) {
    filename = NormalizePath(filename)
    if !FileExists(filename) {
        return
    }
    file, err := os.Open(filename)
    if err != nil {
        return
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" {
            continue
        }
        fmt.Println(line)
    }
    return
}

// ReadingFileUnique Reading file and return content as []string
func ReadingFileUnique(filename string) []string {
    var result []string
    if strings.Contains(filename, "~") {
        filename, _ = homedir.Expand(filename)
    }
    file, err := os.Open(filename)
    if err != nil {
        return result
    }
    defer file.Close()

    seen := make(map[string]bool)
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        val := strings.TrimSpace(scanner.Text())
        // unique stuff
        if val == "" {
            continue
        }
        if seen[val] {
            continue
        }

        seen[val] = true
        result = append(result, val)
    }

    if err := scanner.Err(); err != nil {
        return result
    }
    return result
}

// WriteToFile write string to a file
func WriteToFile(filename string, data string) (string, error) {
    file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
    if err != nil {
        return "", err
    }
    defer file.Close()

    _, err = io.WriteString(file, data+"\n")
    if err != nil {
        return "", err
    }
    return filename, file.Sync()
}

// AppendToContent append string to a file
func AppendToContent(filename string, data string) (string, error) {
    // If the file doesn't exist, create it, or append to the file
    f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return "", err
    }
    if _, err := f.Write([]byte(data + "\n")); err != nil {
        return "", err
    }
    if err := f.Close(); err != nil {
        return "", err
    }
    return filename, nil
}

// FileExists check if file is exist or not
func FileExists(filename string) bool {
    filename = NormalizePath(filename)
    _, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return true
}

// FolderExists check if file is exist or not
func FolderExists(foldername string) bool {
    foldername = NormalizePath(foldername)
    if _, err := os.Stat(foldername); os.IsNotExist(err) {
        return false
    }
    return true
}

// FileLength count len of file
func FileLength(filename string) int {
    filename = NormalizePath(filename)
    if !FileExists(filename) {
        return 0
    }
    return CountLines(filename)
}

// DirLength count len of file
func DirLength(dir string) int {
    dir = NormalizePath(dir)
    files, err := ioutil.ReadDir(dir)
    if err != nil {
        dir = dir + "/"
        files, err = ioutil.ReadDir(dir)
        if err == nil {
            return len(files)
        }
        return 0
    }
    return len(files)
}

// Copy append content to a file
func Copy(src string, dest string) {
    src = NormalizePath(src)
    dest = NormalizePath(dest)
    if !FileExists(src) || FileLength(src) <= 0 {
        return
    }
    input, _ := ioutil.ReadFile(src)
    ioutil.WriteFile(dest, input, 0644)
}

// GetTS get current timestamp and return a string
func GetTS() string {
    return strconv.FormatInt(time.Now().Unix(), 10)
}

// GenHash gen SHA1 hash from string
func GenHash(text string) string {
    h := sha1.New()
    h.Write([]byte(text))
    hashed := h.Sum(nil)
    return fmt.Sprintf("%x", hashed)
}

// GetFileSize get file size of a file in GB
func GetFileSize(src string) float64 {
    var sizeGB float64
    fi, err := os.Stat(NormalizePath(src))
    if err != nil {
        return sizeGB
    }
    // get the size
    size := fi.Size()
    sizeGB = float64(size) / (1024 * 1024 * 1024)
    return sizeGB
}

// RandomString return a random string with length
func RandomString(n int) string {
    var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
    var letter = []rune("abcdefghijklmnopqrstuvwxyz")
    b := make([]rune, n)
    for i := range b {
        b[i] = letter[seededRand.Intn(len(letter))]
    }
    return string(b)
}

// runCmdWithOutput just run os command
func runCmdWithOutput(cmd string) string {
    DebugF("Execute: %s", cmd)
    command := []string{
        "bash",
        "-c",
        cmd,
    }
    realCmd := exec.Command(command[0], command[1:]...)
    // output command output to std too
    output, _ := realCmd.CombinedOutput()
    return string(output)
}

// RunCmdWithOutput run command with timeout
func RunCmdWithOutput(command string, timeoutRaw ...string) string {
    if len(timeoutRaw) == 0 {
        return runCmdWithOutput(command)
    }

    timeout := CalcTimeout(timeoutRaw[0])
    DebugF("Run command with %v seconds timeout", timeout)
    var out string

    c := context.Background()
    deadline := time.Now().Add(time.Duration(timeout) * time.Second)
    c, cancel := context.WithDeadline(c, deadline)
    defer cancel()
    go func() {
        out = RunCmdWithOutput(command)
        cancel()
    }()

    select {
    case <-c.Done():
        return out
    case <-time.After(time.Duration(timeout) * time.Second):
        return out + "\n[err] command got timeout"
    }
}

func runCommandWithError(cmd string) (string, error) {
    DebugF("Execute: %s", cmd)
    command := []string{
        "bash",
        "-c",
        cmd,
    }
    var output string
    realCmd := exec.Command(command[0], command[1:]...)

    // output command output to std too
    cmdReader, _ := realCmd.StdoutPipe()
    scanner := bufio.NewScanner(cmdReader)
    go func() {
        for scanner.Scan() {
            out := scanner.Text()
            DebugF(out)
            output += out + "\n"
        }
    }()
    if err := realCmd.Start(); err != nil {
        return output, err
    }
    if err := realCmd.Wait(); err != nil {
        return output, err
    }
    return output, nil
}

// RunCommandWithErr Run a command
func RunCommandWithErr(command string, timeoutRaw ...string) (string, error) {
    if len(timeoutRaw) == 0 {
        return runCommandWithError(command)
    }
    var output string
    var err error

    timeout := CalcTimeout(timeoutRaw[0])
    DebugF("Run command with %v seconds timeout", timeout)
    var out string

    c := context.Background()
    deadline := time.Now().Add(time.Duration(timeout) * time.Second)
    c, cancel := context.WithDeadline(c, deadline)
    defer cancel()
    go func() {
        out, err = runCommandWithError(command)
        cancel()
    }()

    select {
    case <-c.Done():
        return output, err
    case <-time.After(time.Duration(timeout) * time.Second):
        return out, fmt.Errorf("command got timeout")
    }
}

func RunOSCommand(cmd string) (string, error) {
    DebugF("Execute: %s", cmd)
    command := []string{
        "bash",
        "-c",
        cmd,
    }
    var output string
    realCmd := exec.Command(command[0], command[1:]...)

    // output command output to std too
    cmdReader, _ := realCmd.StdoutPipe()
    scanner := bufio.NewScanner(cmdReader)
    go func() {
        for scanner.Scan() {
            out := scanner.Text()
            DebugF(out)
            output += out
        }
    }()
    if err := realCmd.Start(); err != nil {
        return output, err
    }
    if err := realCmd.Wait(); err != nil {
        return output, err
    }
    return output, nil
}

// RunCommandWithoutOutput Run a command
func RunCommandWithoutOutput(cmd string) error {
    command := []string{
        "bash",
        "-c",
        cmd,
    }
    DebugF("[Exec] %v", command)
    realCmd := exec.Command(command[0], command[1:]...)
    cmdReader, _ := realCmd.StdoutPipe()
    scanner := bufio.NewScanner(cmdReader)
    go func() {
        for scanner.Scan() {
            InforF(scanner.Text())
        }
    }()
    if err := realCmd.Start(); err != nil {
        return err
    }
    if err := realCmd.Wait(); err != nil {
        return err
    }
    return nil
}

// StripPath just Base64 Encode
func StripPath(raw string) string {
    raw = NormalizePath(raw)
    raw = strings.Replace(raw, "/", "_", -1)
    raw = strings.Replace(raw, "..", "_", -1)
    return raw
}

// ZippedFolder zip a folder
func ZippedFolder(src string, dest string) error {
    baseDest := path.Base(dest)
    if FileExists(baseDest) {
        os.RemoveAll(baseDest)
    }
    file, err := os.Create(dest + ".zip")
    if err != nil {
        return err
    }
    defer file.Close()

    w := zip.NewWriter(file)
    defer w.Close()

    walker := func(asbPath string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() {
            return nil
        }
        file, err := os.Open(asbPath)
        if err != nil {
            return err
        }
        defer file.Close()
        relPath := strings.Replace(asbPath, src, baseDest, -1)
        f, err := w.Create(relPath)
        if err != nil {
            return err
        }

        _, err = io.Copy(f, file)
        if err != nil {
            return err
        }

        return nil
    }
    err = filepath.Walk(src, walker)
    if err != nil {
        return err
    }
    return nil
}

// CountLines Return the lines amount of the file
func CountLines(filename string) int {
    var amount int
    if strings.HasPrefix(filename, "~") {
        filename, _ = homedir.Expand(filename)
    }
    file, err := os.Open(filename)
    if err != nil {
        return amount
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        val := strings.TrimSpace(scanner.Text())
        if val == "" {
            continue
        }
        amount++
    }

    if err := scanner.Err(); err != nil {
        return amount
    }
    return amount
}

// CleanPath get environment variable
func CleanPath(raw string) string {
    raw = NormalizePath(raw)
    base := raw
    if FileExists(base) {
        base = filepath.Base(raw)
    }
    out := strings.ReplaceAll(base, "/", "_")
    out = strings.ReplaceAll(out, ":", "_")
    return out
}

func IsFile(src string) bool {
    fi, err := os.Stat(NormalizePath(src))
    if err != nil {
        return false
    }
    switch mode := fi.Mode(); {
    case mode.IsDir():
        return false
    case mode.IsRegular():
        if FileLength(src) > 0 {
            return true
        }
        return false
    }
    return false
}

func FolderLength(dir string) int {
    dir = NormalizePath(dir)
    var length int
    if FileExists(dir) {
        dir = path.Dir(dir)
    }
    files, err := ioutil.ReadDir(dir)
    if err != nil {
        return length
    }
    length = len(files)
    return length
}

// ImageAsBase64 read image file as a string
func ImageAsBase64(src string) string {
    src = NormalizePath(src)
    f, err := os.Open(src)
    if err != nil {
        ErrorF("File not found: %v", src)
        return ""
    }

    // Read entire JPG into byte slice.
    reader := bufio.NewReader(f)
    content, err := ioutil.ReadAll(reader)
    if err != nil {
        return ""
    }
    // Encode as base64.
    encoded := base64.StdEncoding.EncodeToString(content)
    return encoded
}

// Base64Encode read image file as a string
func Base64Encode(raw string) string {
    return base64.StdEncoding.EncodeToString([]byte(raw))
}

const bufSize = 1024

// OffsetRange represents a content block of a file.
type OffsetRange struct {
    File  string
    Start int64
    Stop  int64
}

// SplitLineChunks splits file into chunks.
// The whole line are guaranteed to be split in the same chunk.
func SplitLineChunks(filename string, chunks int) ([]OffsetRange, error) {
    info, err := os.Stat(filename)
    if err != nil {
        return nil, err
    }

    if chunks <= 1 {
        return []OffsetRange{
            {
                File:  filename,
                Start: 0,
                Stop:  info.Size(),
            },
        }, nil
    }

    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var ranges []OffsetRange
    var offset int64
    // avoid the last chunk too few bytes
    preferSize := info.Size()/int64(chunks) + 1
    for {
        if offset+preferSize >= info.Size() {
            ranges = append(ranges, OffsetRange{
                File:  filename,
                Start: offset,
                Stop:  info.Size(),
            })
            break
        }

        offsetRange, err := nextRange(file, offset, offset+preferSize)
        if err != nil {
            return nil, err
        }

        ranges = append(ranges, offsetRange)
        if offsetRange.Stop < info.Size() {
            offset = offsetRange.Stop
        } else {
            break
        }
    }

    return ranges, nil
}

func nextRange(file *os.File, start, stop int64) (OffsetRange, error) {
    offset, err := skipPartialLine(file, stop)
    if err != nil {
        return OffsetRange{}, err
    }

    return OffsetRange{
        File:  file.Name(),
        Start: start,
        Stop:  offset,
    }, nil
}

func skipPartialLine(file *os.File, offset int64) (int64, error) {
    for {
        skipBuf := make([]byte, bufSize)
        n, err := file.ReadAt(skipBuf, offset)
        if err != nil && err != io.EOF {
            return 0, err
        }
        if n == 0 {
            return 0, io.EOF
        }

        for i := 0; i < n; i++ {
            if skipBuf[i] != '\r' && skipBuf[i] != '\n' {
                offset++
            } else {
                for ; i < n; i++ {
                    if skipBuf[i] == '\r' || skipBuf[i] == '\n' {
                        offset++
                    } else {
                        return offset, nil
                    }
                }
                return offset, nil
            }
        }
    }
}

// A RangeReader is used to read a range of content from a file.
type RangeReader struct {
    file  *os.File
    start int64
    stop  int64
}

// NewRangeReader returns a RangeReader, which will read the range of content from file.
func NewRangeReader(file *os.File, start, stop int64) *RangeReader {
    return &RangeReader{
        file:  file,
        start: start,
        stop:  stop,
    }
}

// Read reads the range of content into p.
func (rr *RangeReader) Read(p []byte) (n int, err error) {
    stat, err := rr.file.Stat()
    if err != nil {
        return 0, err
    }

    if rr.stop < rr.start || rr.start >= stat.Size() {
        return 0, fmt.Errorf("exceed file size")
    }

    if rr.stop-rr.start < int64(len(p)) {
        p = p[:rr.stop-rr.start]
    }

    n, err = rr.file.ReadAt(p, rr.start)
    if err != nil {
        return n, err
    }

    rr.start += int64(n)
    return
}

func Move(src string, dest string) error {
    src = NormalizePath(src)
    if !IsFile(src) && DirLength(src) == 0 {
        return fmt.Errorf("source does not exist: %v", src)
    }

    dest = NormalizePath(dest)
    os.RemoveAll(dest)
    DebugF("Moving %v --> %v", src, dest)
    return os.Rename(src, dest)
}

func IsWritable(filename string) (isWritable bool, err error) {
    isWritable = false
    info, err := os.Stat(filename)
    if err != nil {
        return
    }

    err = nil
    if !info.IsDir() {
        return
    }

    // Check if the user bit is enabled in file permission
    if info.Mode().Perm()&(1<<(uint(7))) == 0 {
        return
    }

    var stat syscall.Stat_t
    if err = syscall.Stat(filename, &stat); err != nil {
        return
    }

    err = nil
    if uint32(os.Geteuid()) != stat.Uid {
        isWritable = false
        return
    }

    isWritable = true
    return
}
