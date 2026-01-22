package heuristics

import (
    "os"
    "testing"
)

func TestDetectType_File(t *testing.T) {
    // Create temp file
    f, err := os.CreateTemp("", "test-file*.txt")
    if err != nil {
        t.Fatal(err)
    }
    defer func() { _ = os.Remove(f.Name()) }()
    _ = f.Close()

    got := DetectType(f.Name())
    if got != TargetTypeFile {
        t.Errorf("DetectType() = %v, want %v", got, TargetTypeFile)
    }
}

func TestDetectType_DomainNotFile(t *testing.T) {
    got := DetectType("example.com")
    if got != TargetTypeDomain {
        t.Errorf("DetectType() = %v, want %v", got, TargetTypeDomain)
    }
}

func TestDetectType_URL(t *testing.T) {
    got := DetectType("https://example.com/path")
    if got != TargetTypeURL {
        t.Errorf("DetectType() = %v, want %v", got, TargetTypeURL)
    }
}

func TestParseFileTarget(t *testing.T) {
    tests := []struct {
        name     string
        filePath string
        wantRoot string
    }{
        {
            name:     "simple filename",
            filePath: "/tmp/urls-input.txt",
            wantRoot: "urls-input-file",
        },
        {
            name:     "underscore replacement",
            filePath: "/tmp/my_target_list.txt",
            wantRoot: "my-target-list-file",
        },
        {
            name:     "multiple underscores",
            filePath: "/path/to/some_long_file_name.csv",
            wantRoot: "some-long-file-name-file",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            info, err := ParseFileTarget(tt.filePath)
            if err != nil {
                t.Errorf("ParseFileTarget() error = %v", err)
                return
            }
            if info.RootDomain != tt.wantRoot {
                t.Errorf("ParseFileTarget() RootDomain = %v, want %v", info.RootDomain, tt.wantRoot)
            }
            if info.Type != TargetTypeFile {
                t.Errorf("ParseFileTarget() Type = %v, want %v", info.Type, TargetTypeFile)
            }
            if info.Original != tt.filePath {
                t.Errorf("ParseFileTarget() Original = %v, want %v", info.Original, tt.filePath)
            }
        })
    }
}

func TestAnalyze_FileTarget(t *testing.T) {
    // Create temp file
    f, err := os.CreateTemp("", "test_analysis*.txt")
    if err != nil {
        t.Fatal(err)
    }
    defer func() { _ = os.Remove(f.Name()) }()
    _ = f.Close()

    info, err := Analyze(f.Name(), "basic")
    if err != nil {
        t.Errorf("Analyze() error = %v", err)
        return
    }
    if info.Type != TargetTypeFile {
        t.Errorf("Analyze() Type = %v, want %v", info.Type, TargetTypeFile)
    }
}
