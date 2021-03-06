# pd-cli
A Product Delivery CLI toolbox of awesomeness

## Install

Install [dep](https://github.com/golang/de)

```bash
go get github.com/mozilla/pd-cli
cd $GOPATH/src/github.com/mozilla/pd-cli
dep ensure && go install .
pd-cli -h
```

## Github Access

The pd-cli requires a Github access token to perform most of its commands.

* Create a [GH Personal Access Token](https://github.com/settings/tokens/)
* Give the token *all* repo permissions
* Set the token in the environment variable `GH_ACCESS_TOKEN`

## Usage

* Print help:
  * `pd-cli help`
  * `pd-cli repo help`
  * `pd-cli create-milestone help`
* Check a repo (these are the same):
  * `pd-cli repo check all github.com/mozilla/pd-cli`
  * `pd-cli repo check all mozilla/pd-cli`
* Specific checks:
  * `pd-cli repo check help` -- prints help
  * `pd-cli repo check topic owner/repo` -- checks for Product Delivery repo topics
  * `pd-cli repo check labels owner/repo` -- verifies Product Delivery labels are correct
  * `pd-cli repo check unassigned owner/repo`` -- checks all P1 issues are assigned to somebody
  * `pd-cli repo check unlabled owner/repo`` -- finds issues that do not have a label
  * `pd-cli repo check milestones owner/repo`` -- verifies milestones have a project to track them
* Initializing a repo:
  * `pd-cli repo init mozilla/pd-cli` - creates labels and a default project
  * `pd-cli repo init -p "first project" mozilla/pd-cli` - creates labels and a specific project
  * `pd-cli repo create-milestone -m "Version 2.0" mozilla/pd-cli` - creates a new milestone
  * `pd-cli repo create-project -p "first project" mozilla/pd-cli` - creates a project
