package main

import (
	"fmt"
	"strings"

	"github.com/go-github/github"
	"flag"
"os"
)

func GetRepos(client *github.Client) []github.Repository {
	opt := &github.RepositoryListByOrgOptions{}
	allRepos := []github.Repository{}
	fmt.Println("Gathering repos...")
	for {
		repos, resp, err := client.Repositories.ListByOrg("intelsdi-x", opt)
		if err != nil {panic(err)}
		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}
	return allRepos
}

func GetPull(client *github.Client, repo github.Repository) []github.PullRequest {
	pulls := []github.PullRequest{}
	opt := &github.PullRequestListOptions{State: "open"}
	//fmt.Printf("Gathering open pull requests for %s\n", *repo.Name)
	pulls, _, err := client.PullRequests.List("intelsdi-x", *repo.Name, opt)
	if err != nil {panic(err)}
	return pulls
}

func GetIssues(client *github.Client, repo github.Repository) []github.Issue {
	opt := &github.IssueListByRepoOptions{State: "open"}
	//fmt.Printf("Gathering open issues for %s\n", *repo.Name)
	issues, _, err := client.Issues.ListByRepo("intelsdi-x", *repo.Name, opt)
	if err != nil {panic(err)}
	return issues
}

func main(){

	user := flag.String("user", "", "github user")
	passwd := flag.String("passwd", "", "github password")

	flag.Parse()

	if *user == "" || *passwd == "" {
		fmt.Printf("Usge of %s:\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	basic := github.BasicAuthTransport{Username: *user, Password: *passwd}
	client := github.NewClient(basic.Client())

	f, err := os.Create("gitstats.csv")
	if err != nil {panic(err)}
	defer f.Close()

	repos := GetRepos(client)
	fmt.Printf("Found %d snap repos\n", len(repos))
	for i, repo := range repos {
		fmt.Printf("Done %d%s", int(float32(i+1)/float32(len(repos))*100), "%")
		if strings.Contains(*repo.Name, "snap") {
			pulls := GetPull(client, repo)
			issues := GetIssues(client, repo)
			if len(pulls) > 0 || len(issues) > 0 {
				for _, pull := range pulls {
					f.WriteString(fmt.Sprintf("%s;Pull Request;%d;%s\n", *repo.Name, *pull.Number, *pull.Title))
				}
				for _, issue := range issues {
					f.WriteString(fmt.Sprintf("%s;Issue;%d;%s\n", *repo.Name, *issue.Number, *issue.Title))
				}
			}
			f.Sync()
		}
		fmt.Printf("\033[0J")
		fmt.Printf("\033[%dA\n", 1)
	}
	fmt.Println("")
}
