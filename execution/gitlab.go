package execution

import (
	"errors"
	"fmt"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"log"
	"strings"

	"github.com/xanzy/go-gitlab"
)

// GitlabAuth do authentication with gitlab
func GitlabAuth(options libs.Options) (*gitlab.Client, error) {
	if options.NoGit {
		return nil, nil
	}
	// prefer using username
	if options.Git.Username != "GITLAB_USER" {
		utils.DebugF("Do authen with user: %s", options.Git.Username)
		git, err := gitlab.NewBasicAuthClient(
			options.Git.Username,
			options.Git.Password,
			gitlab.WithBaseURL(options.Git.BaseURL),
		)
		if err != nil {
			utils.WarnF("Error authen in Gitlab %v: %v", options.Git.BaseURL, err)
			return nil, err
		}
		return git, nil
	}

	if options.Git.Token == "" {
		return nil, errors.New("no credentials provided")
	}
	utils.DebugF("Do authen with token: %s", options.Git.Token)
	git, err := gitlab.NewClient(
		options.Git.Token,
		gitlab.WithBaseURL(options.Git.BaseURL),
	)
	if err != nil {
		utils.WarnF("Error authen in Gitlab %v: %v", options.Git.BaseURL, err)
		return nil, err
	}
	return git, nil
}

// CreateGitlabRepo create gitlab repo
func CreateGitlabRepo(repo string, tag string, options libs.Options) string {
	if options.NoGit {
		return ""
	}
	git, err := GitlabAuth(options)
	if err != nil {
		return ""
	}

	utils.InforF("Create repo: %v", repo)
	tags := []string{options.Git.DefaultTag}
	if tag != "" {
		if strings.Contains(tag, ",") {
			tags = append(tags, strings.Split(tag, ",")...)
		} else {
			tags = append(tags, tag)
		}
	}

	opt := gitlab.CreateProjectOptions{
		Name:          gitlab.String(repo),
		Path:          nil,
		DefaultBranch: gitlab.String("master"),
		Visibility:    gitlab.Visibility(gitlab.PrivateVisibility),
		TagList:       &tags,
	}
	gRepo, _, err := git.Projects.CreateProject(&opt, nil)
	if err != nil {
		utils.WarnF("Error to create %v: %v", repo, err)
		return ""
	}

	optMember := gitlab.AddProjectMemberOptions{
		UserID:      gitlab.Int(options.Git.DefaultUID),
		AccessLevel: gitlab.AccessLevel(gitlab.MaintainerPermissions),
	}
	git.ProjectMembers.AddProjectMember(gRepo.ID, &optMember)
	utils.InforF("Created repo at: %s with pid: %v", gRepo.SSHURLToRepo, gRepo.ID)
	return gRepo.SSHURLToRepo
}

// DeleteRepo delete repo by id or name
func DeleteRepo(repo string, pid int, options libs.Options) {
	if options.NoGit {
		return
	}
	git, err := GitlabAuth(options)
	if err != nil {
		return
	}
	// select pid first
	if pid != 0 {
		utils.InforF("Delete repo with id: %v", pid)
		git.Projects.DeleteProject(pid)
		return
	}

	user, _, err := git.Users.CurrentUser()
	projects, _, err := git.Projects.ListUserProjects(user.ID, nil)
	for _, project := range projects {
		if project.Name == repo {
			utils.InforF("Delete repo: %v", repo)
			git.Projects.DeleteProject(project.ID)
			return
		}
	}
	utils.WarnF("Project not found: %v", repo)
}

// ListProjects delete repo by id or name
func ListProjects(gitUser int, options libs.Options) {
	if options.NoGit {
		return
	}
	git, err := GitlabAuth(options)
	if err != nil {
		utils.WarnF("Err get do authen user: %v", err)
		return
	}
	uid := gitUser
	var username string
	if gitUser == 0 {
		user, _, err := git.Users.CurrentUser()
		if err != nil {
			utils.WarnF("Err get current user: %v", err)
			return
		}
		uid = user.ID
		username = user.Username
	} else {
		//user, _, err := git.Users.GetUser(uid, git.Users.)
		//if err != nil {
		//	utils.WarnF("Err get current user: %v", err)
		//	return
		//}
		//username = user.Username
	}
	utils.InforF("Listing projects of uid: %v", uid)
	projects, _, err := git.Projects.ListUserProjects(uid, nil)
	if err != nil {
		utils.WarnF("Error listing projects: %v", err)
		return
	}
	for _, project := range projects {
		fmt.Printf("%30s     --   %10d\n", fmt.Sprintf("%s/%s", username, project.Name), project.ID)
	}
}

// This example shows how to create a client with username and password.
func GitAuthSample() {
	// git, err := gitlab.NewBasicAuthClient(
	// 	"user",
	// 	"password",
	// 	gitlab.WithBaseURL("https://gitlab.com"),
	// )
	git, err := gitlab.NewClient(
		"token-here-s",
		gitlab.WithBaseURL("https://gitlab.com"),
	)

	if err != nil {
		log.Fatal(err)
	}

	// List all projects
	user, _, err := git.Users.CurrentUser()
	//spew.Dump(user)

	// Create
	projects, _, err := git.Projects.ListUserProjects(user.ID, nil)
	////git.Projects.
	//group, _, _ := git.Groups.ListAllGroupMembers(9182310, nil)
	//spew.Dump(group)
	//tags := []string{"osm", "test"}
	////git.Projects.ListUserProjects()
	//opt := gitlab.CreateProjectOptions{
	//	Name:                        gitlab.String("test-project-tag-osm"),
	//	Path:                        nil,
	//	DefaultBranch:               gitlab.String("master"),
	//	//GroupWithProjectTemplatesID: gitlab.Int(9182310),
	//	Visibility:                  gitlab.Visibility(gitlab.PrivateVisibility),
	//	TagList:                     &tags,
	//}
	//git.Projects.CreateProject(&opt, nil)

	if err != nil {
		log.Fatal(err)
	}

	//git.Projects.DeleteProject(pid)

	log.Printf("Found %d projects", len(projects))
	for _, project := range projects {
		fmt.Printf("%30s   --   %10d\n", fmt.Sprintf("%s/%s", user.Username, project.Name), project.ID)
	}

}
