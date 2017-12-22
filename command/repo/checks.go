package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/go-github/github"
	"gopkg.in/urfave/cli.v1"
)

// this file contains all the verification checking functions. Look in repo.go for
// where all these private functions are wired up. These functions depend on
// preflight() which sets ghClient, owner and repo

func checkAll(c *cli.Context) error {
	funcs := []cli.ActionFunc{
		checkTopic,
		checkLabels,
		checkUnassigned,
		checkUnlabled,
		checkMilestones,
	}

	for _, f := range funcs {
		if err := f(c); err != nil {
			return err
		}
	}
	return nil
}

// checkTopic ensures "product-delivery" topic is assigned to the repo
func checkTopic(c *cli.Context) error {
	fmt.Fprintf(c.App.Writer, "Checking for [product-delivery] topic\n")
	ctx := context.Background()
	topics, _, err := ghClient.Repositories.ListAllTopics(ctx, owner, repo)
	if err != nil {
		fmt.Fprintf(c.App.Writer, " - Error: %s\n", err.Error())
		return err
	}

	for _, name := range topics.Names {
		if name == "product-delivery" {
			fmt.Fprintf(c.App.Writer, " - OK. Found product-delivery topic\n")
			return nil
		}
	}

	fmt.Fprintf(c.App.Writer, " - Error: product-delivery topic not set\n")
	return errors.New("product-delivery topic not set")
}

func checkLabels(c *cli.Context) error {
	fmt.Fprintf(c.App.Writer, "Checking Labels\n")
	labels, _, err := ghClient.Issues.ListLabels(context.Background(), owner, repo, nil)

	if err != nil {
		fmt.Fprintf(c.App.Writer, " - Error: %s\n", err.Error())
		return err
	}

	// expected labels and their color
	standardLabels := map[string]string{
		"bug":             "b60205",
		"security":        "b60205",
		"documentation":   "0e8a16",
		"fix":             "0e8a16",
		"new-feature":     "0e8a16",
		"P1":              "ffa32c",
		"P2":              "ffa32c",
		"P3":              "ffa32c",
		"P5":              "ffa32c",
		"proposal":        "1d76db",
		"question":        "1d76db",
		"support-request": "1d76db",
	}

	for _, label := range labels {
		if label == nil || label.Name == nil || label.Color == nil {
			continue
		}

		name := *label.Name
		color := *label.Color

		if expectedColor, ok := standardLabels[name]; !ok {
			// not a standard label
			if color != "5319e7" {
				fmt.Fprintf(c.App.Writer, " - Error: [%s] should have color #5319e7\n", name)
			} else {
				fmt.Fprintf(c.App.Writer, " - OK. [%s] verified\n", name)
			}
		} else {
			// check standard label has correct color
			if color != expectedColor {
				fmt.Fprintf(c.App.Writer, " - Error: standard label [%s] should have color #%s\n", name, expectedColor)
			} else {
				fmt.Fprintf(c.App.Writer, " - OK. [%s] verified\n", name)
			}

			// delete it so we know how many are missing
			delete(standardLabels, name)
		}
	}

	// check for missing standard labels
	for missing, color := range standardLabels {
		fmt.Fprintf(c.App.Writer, " - Error: missing %s (%s)\n", missing, color)
	}

	return nil
}

func checkUnassigned(c *cli.Context) error {
	fmt.Fprintf(c.App.Writer, "Checking Unassigned Issues\n")

	query := fmt.Sprintf("repo:%s/%s is:open no:assignee label:P1", owner, repo)
	results, _, err := ghClient.Search.Issues(context.Background(), query, nil)
	if err != nil {
		fmt.Fprintf(c.App.Writer, " - Error: %s\n", err.Error())
		return err
	}

	count := *results.Total
	if count > 0 {
		fmt.Fprintf(c.App.Writer, " - Error: %d unassigned P1 issues\n", count)
		for _, issue := range results.Issues {
			fmt.Fprintf(c.App.Writer, "  #%-4d %s", *issue.Number, *issue.Title)
		}
	} else {
		fmt.Fprintf(c.App.Writer, " - OK. All P1 issues assigned\n")
	}
	return nil
}
func checkUnlabled(c *cli.Context) error {
	fmt.Fprintf(c.App.Writer, "Checking Unlabled\n")
	query := fmt.Sprintf("repo:%s/%s is:open no:label is:issue", owner, repo)
	results, _, err := ghClient.Search.Issues(context.Background(), query, nil)
	if err != nil {
		fmt.Fprintf(c.App.Writer, " - Error: %s\n", err.Error())
		return err
	}

	unassigned := *results.Total
	if unassigned > 0 {
		fmt.Fprintf(c.App.Writer, " - Error: %d issues unlabeled\n", unassigned)
		for _, issue := range results.Issues {
			fmt.Fprintf(c.App.Writer, "   #%-4d %s\n", *issue.Number, *issue.Title)
		}
	} else {
		fmt.Fprintf(c.App.Writer, " - OK. All issues are labeled\n")
	}

	return nil
}

func checkMilestones(c *cli.Context) error {
	fmt.Fprintf(c.App.Writer, "Checking Milestones\n")
	ctx := context.Background()

	milestones, _, err := ghClient.Issues.ListMilestones(ctx, owner, repo, nil)
	if err != nil {
		fmt.Fprintf(c.App.Writer, " - Error: Feteching milestones, %s\n", err.Error())
		return err
	}
	projects, _, err := ghClient.Repositories.ListProjects(ctx, owner, repo, nil)
	if err != nil {
		fmt.Fprintf(c.App.Writer, " - Error: Fetching projects, %s\n", err.Error())
		return err
	}

	pMap := make(map[string]*github.Project)
	for _, p := range projects {
		pMap[*p.Name] = p
	}

	errHappend := false
	for _, milestone := range milestones {
		if project, found := pMap[*milestone.Title]; !found {
			fmt.Fprintf(c.App.Writer, " - Error: %s does not have a matching project", *milestone.Title)
		} else {
			// check the project's columns
			pCols, _, err := ghClient.Projects.ListProjectColumns(ctx, *project.ID, nil)
			if err != nil {
				fmt.Fprintf(c.App.Writer, " - Error: Fetching project columns, %s\n", err.Error())
				return err
			}

			flags := 0
			for _, col := range pCols {
				switch *col.Name {
				case "Backlog":
					flags |= 0x01
				case "In Progress":
					flags |= 0x02
				case "Blocked":
					flags |= 0x04
				case "Completed":
					flags |= 0x08
				default:
					fmt.Fprintf(c.App.Writer, ` - Error: Project "%s" has unexpected column %s\n`,
						*project.Name, *col.Name)
					errHappend = true
				}
			}

			if flags&0x01 == 0 {
				fmt.Fprintf(c.App.Writer, ` - Error: Project "%s" missing "Backlog" column\n`, *project.Name)
				errHappend = true
			}
			if flags&0x02 == 0 {
				fmt.Fprintf(c.App.Writer, ` - Error: Project "%s" missing "In Progress" column\n`, *project.Name)
				errHappend = true
			}
			if flags&0x04 == 0 {
				fmt.Fprintf(c.App.Writer, ` - Error: Project "%s" missing "Blocked" column\n`, *project.Name)
				errHappend = true
			}
			if flags&0x08 == 0 {
				fmt.Fprintf(c.App.Writer, ` - Error: Project "%s" missing "Completed" column\n`, *project.Name)
				errHappend = true
			}
		}
	}

	if !errHappend {
		fmt.Fprintf(c.App.Writer, " - OK. Milestones verified\n")
	}

	return nil
}