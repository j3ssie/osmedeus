package execution

import (
    "errors"
    "fmt"
    "os"
    "path"
    "strings"

    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/utils"
)

/*

Base Git command:
GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no -i /path/to/private/key' git xxx

*/

/*
GitDiff('path-to-diff-file', 'path-to-output')
GitDiff('path-to-diff-folder', 'path-to-output')
GitDiff('path-to-diff-file', 'path-to-output', 'number-of-commit-to-diff')
*/

// GitDiff run git diff command
func GitDiff(dest string, output string, history string, options libs.Options) {
    if options.NoGit || options.Storages["secret_key"] == "" {
        return
    }
    if !utils.FileExists(dest) {
        utils.WarnF("File not found: %v", dest)
        return
    }
    utils.DebugF("Git Diff: %v", dest)
    diffCmd := fmt.Sprintf("GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no -i %v' git diff -U0 HEAD~%v --output=%v %v", options.Storages["secret_key"], history, output, dest)
    Execution(diffCmd, options)
}

// LoopGitDiff like GitDiff but take input as a file
func LoopGitDiff(src string, output string, options libs.Options) {
    lines := utils.ReadingFileUnique(src)
    if len(lines) == 0 {
        return
    }
    for _, line := range lines {
        GitDiff(line, output, "1", options)
    }
}

// DiffCompare run git diff command
func DiffCompare(src string, dest string, output string, options libs.Options) {
    if options.NoGit || options.Storages["secret_key"] == "" {
        return
    }
    // if !utils.FileExists(src) || !utils.FileExists(dest) {
    // 	return
    // }
    diffCmd := fmt.Sprintf("GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no -i %v' git diff --no-index --output=%v %v %v", options.Storages["secret_key"], output, src, dest)
    Execution(diffCmd, options)
}

// PullResult pull latest data from result repo
func PullResult(storageFolder string, options libs.Options) {
    if options.NoGit || options.Storages["secret_key"] == "" {
        return
    }
    if !utils.FolderExists(storageFolder) {
        return
    }
    utils.DebugF("git pull on: %v", storageFolder)
    pullCmd := fmt.Sprintf("GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no -i %v' git -C %v pull -f", options.Storages["secret_key"], storageFolder)
    Execution(pullCmd, options)
}

// PushResult push result to git repo
func PushResult(storageFolder string, commitMess string, options libs.Options) {
    if options.NoGit || options.Storages["secret_key"] == "" {
        return
    }

    if !utils.FolderExists(storageFolder) {
        return
    }
    utils.DebugF("git push on: %v", storageFolder)
    cmds := []string{
        fmt.Sprintf(`GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no -i %v' git -C %v add -A`, options.Storages["secret_key"], storageFolder),
        fmt.Sprintf(`GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no -i %v' git -C %v commit -m "%v"`, options.Storages["secret_key"], storageFolder, commitMess),
        fmt.Sprintf(`GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no -i %v' git -C %v push -f`, options.Storages["secret_key"], storageFolder),
        fmt.Sprintf(`GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no -i %v' git -C %v push -f`, options.Storages["secret_key"], storageFolder),
    }
    // really run git command
    for _, cmd := range cmds {
        Execution(cmd, options)
    }
}

// GitClone update latest UI and Plugins from default repo
func GitClone(url string, dest string, forced bool, options libs.Options) {
    if options.NoGit || options.Storages["secret_key"] == "" {
        utils.WarnF("Storage Disable")
        return
    }

    if url == "" || dest == "" {
        utils.WarnF("Invalid repo or no destination")
        return
    }

    // check if folder is exist or not
    dest = utils.NormalizePath(dest)
    if forced {
        utils.DebugF("Remove: %v", dest)
        os.RemoveAll(dest)
    }

    // if folder exist and have .git folder in it, do git pull instead
    if utils.FolderExists(dest) {
        if utils.FileExists(path.Join(dest, "/.git/HEAD")) {
            pullCmd := fmt.Sprintf("GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no -i %v' git -C %v pull -f", options.Storages["secret_key"], dest)
            Execution(pullCmd, options)
            return
        }
    }

    // cloning new
    utils.DebugF("Cloning: %v", url)
    cloneCmd := fmt.Sprintf("GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no -i %v' git clone --depth=1 %v %v", options.Storages["secret_key"], url, dest)
    Execution(cloneCmd, options)
}

// CloneRepo clone the repo
func CloneRepo(url string, dest string, options libs.Options) error {
    if !ValidGitURL(url) {
        return errors.New("invalid repo name")
    }
    if options.NoGit || options.Storages["secret_key"] == "" {
        return errors.New("storage Disable")
    }

    if url == "" || dest == "" {
        return errors.New("storage didn't setup correctly")
    }
    // check if folder is exist or not
    dest = utils.NormalizePath(dest)
    utils.DebugF("Cloning: %v", url)

    if utils.FolderExists(dest) {
        if utils.FileExists(path.Join(dest, "/.git/HEAD")) {
            return nil
        }
        utils.DebugF("Remove: %v", dest)
        os.RemoveAll(dest)
    }

    // cloning new
    cloneCmd := fmt.Sprintf("GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no -i %v' git clone --depth=1 %v %v", options.Storages["secret_key"], url, dest)
    Execution(cloneCmd, options)
    if utils.FolderExists(dest) {
        if utils.FileExists(path.Join(dest, "/.git/HEAD")) {
            return nil
        }
        return nil
    }
    return errors.New("fail to clone Repo")
}

// ValidGitURL simple validate git repo
func ValidGitURL(raw string) bool {
    if strings.TrimSpace(raw) == "" {
        return false
    }
    if !strings.HasPrefix(raw, "git@") {
        return false
    }

    if !strings.Contains(raw, "github.com") && !strings.Contains(raw, "gitlab.com") {
        return false
    }

    return true
}
