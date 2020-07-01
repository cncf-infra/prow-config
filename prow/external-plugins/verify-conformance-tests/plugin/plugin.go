/*
Copyright 2020 CNCF TODO Check how this code should be licensed

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugin

import (
	"bytes"
	"context"
	"fmt"
        "regexp"
        "strings"
	"time"

	githubql "github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"

	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"
	"k8s.io/test-infra/prow/plugins"
	"net/http"
	"io/ioutil"
	// 	"github.com/golang-collections/collections/set"
)

const (
	PluginName     = "verify-conformance-tests"
)

var sleep = time.Sleep

type githubClient interface {
	GetIssueLabels(org, repo string, number int) ([]github.Label, error)
	CreateComment(org, repo string, number int, comment string) error
	BotName() (string, error)
	AddLabel(org, repo string, number int, label string) error
	RemoveLabel(org, repo string, number int, label string) error
	DeleteStaleComments(org, repo string, number int, comments []github.IssueComment, isStale func(github.IssueComment) bool) error
	Query(context.Context, interface{}, map[string]interface{}) error
	GetPullRequest(org, repo string, number int) (*github.PullRequest, error)
	GetPullRequestChanges(org, repo string, number int) ([]github.PullRequestChange, error)
}

type commentPruner interface {
	PruneComments(shouldPrune func(github.IssueComment) bool)
}

// HelpProvider constructs the PluginHelp for this plugin that takes into account enabled repositories.
// HelpProvider defines the type for the function that constructs the PluginHelp for plugins.
func HelpProvider(_ []config.OrgRepo) (*pluginhelp.PluginHelp, error) {
	return &pluginhelp.PluginHelp{
			Description: `The Verify Conformance Tests plugin checks that all required conformance tests have been run for the stated version of Kubernetes`,
		},
		nil
}

func getRequiredTests(log *logrus.Entry, k8sRelease string) [] string {
	// TODO we are effectively hardcoding this and we may layer this out
	// Key'd by k8s release map that points to URLs containing the required conformance tests for that release
	var conformanceTests = map[string]string {
		"v1.15": "https://raw.githubusercontent.com/kubernetes/kubernetes/release-1.15/test/conformance/testdata/conformance.txt",
			"v1.16": "https://raw.githubusercontent.com/kubernetes/kubernetes/blob/release-1.16/test/conformance/testdata/conformance.txt",
			"v1.17": "https://raw.githubusercontent.com/kubernetes/kubernetes/blob/release-1.17/test/conformance/testdata/conformance.txt",
			"v1.18": "https://raw.githubusercontent.com/kubernetes/kubernetes/blob/release-1.18/test/conformance/testdata/conformance.yaml",
			"master": "https://raw.githubusercontent.com/kubernetes/kubernetes/master/test/conformance/testdata/conformance.yaml",
		}

	var requiredTests []string ;
        fileUrl := conformanceTests[k8sRelease]
	resp, err := http.Get(fileUrl)
	if err != nil {
		log.Errorf("getRequiredTests : %+v",err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body) // TODO Handle err

        // TODO How do we get a list of tests returned.
        // TODO Assume YAML all the way for now
	// Unmarshall the YAML ref https://godoc.org/gopkg.in/yaml.v2#example-Unmarshal--Embedded
	log.Info(body)
	return requiredTests
}

// HandleAll is called periodically and the period is setup in main.go
// It runs a Github Query to get all open PRs for this repo which contains k8s conformance requests
//
// Each PR is checked in turn, we check
//   - for the presence of a Release Version in the PR title
//- then we take that version and verify that the e2e test logs refer to that same release version.
//
// if all is in order then we add the verifiable label and a release-Vx.y label
// if there is an inconsistency we add a comment that explains the problem
// and tells the PR submitter to review the documentation
func HandleAll(log *logrus.Entry, ghc githubClient, config *plugins.Configuration) error {
	log.Infof("%v : HandleAll : Checking all PRs for handling", PluginName)

	orgs, repos := config.EnabledReposForExternalPlugin(PluginName) // TODO : Overkill see below

	if len(orgs) == 0 && len(repos) == 0 {
		log.Warnf("HandleAll : No repos have been configured for the %s plugin", PluginName)
		return nil
	}

        // TODO simplify queryOpenPRs
        //      - more general than required
        //      - we deal with a single org and repo
        //      - we target k8s conformance requests sent to the cncf
	var queryOpenPRs bytes.Buffer
	fmt.Fprint(&queryOpenPRs, "archived:false is:pr is:open label:verifiable")
	for _, org := range orgs {
		fmt.Fprintf(&queryOpenPRs, " org:\"%s\"", org)
	}
	for _, repo := range repos {
		fmt.Fprintf(&queryOpenPRs, " repo:\"%s\"", repo)
	}
	prs, err := search(context.Background(), log, ghc, queryOpenPRs.String())

	if err != nil {
		return err
	}
	log.Infof("Considering %d verifiable PRs.", len(prs))

	for _, pr := range prs {
		org := string(pr.Repository.Owner.Login)
		repo := string(pr.Repository.Name)
		prNumber := int(pr.Number)
		sha := string(pr.Commits.Nodes[0].Commit.Oid)

                releaseVersion := getReleaseFromLabel(log, org, repo, prNumber, ghc)

		prLogger := log.WithFields(logrus.Fields{
			//"org":  org,
			//"repo": repo,
			"pr":   prNumber,
                        "title": pr.Title,
                        "release": releaseVersion ,
		})

		// githubClient.CreateComment(ghc, org, repo, prNumber, "Please include the release in the title of this Pull Request" )
		requiredTests := getRequiredTestsForRelease(releaseVersion)
		submittedTests := getSubmittedTestsFromPullReq(prLogger, ghc, org, repo, prNumber, sha)
		submittedTestsPresentInJUnit := checkAllSubmittedTestsArePresent(requiredTests, submittedTests)
		if submittedTestsPresentInJUnit {
                        testRunEvidenceCorrect , err := checkE2eLogHasEvidenceOfTestRuns(prLogger, ghc, org, repo, prNumber, sha, requiredTests)

                        if err != nil {
                                prLogger.WithError(err)
                        }

                        if testRunEvidenceCorrect {
                                githubClient.AddLabel(ghc, org, repo, prNumber, "verified-"+releaseVersion)
                                githubClient.CreateComment(ghc, org, repo, prNumber, "Well done you! You no slouch! VERIFIED!"  )
                        } else { // specifiedRelease not present in logs
				githubClient.AddLabel(ghc, org, repo, prNumber, "evidence-missing")
				githubClient.CreateComment(ghc, org, repo, prNumber,
					"This conformance request has the correct list of tests present in the junit file but is missing evidence from the e2e log file")
			}
                } else {
                        githubClient.AddLabel(ghc, org, repo, prNumber, "required-tests-missing")
                        githubClient.CreateComment(ghc, org, repo, prNumber, "This conformance request failed to include all of the required tests for " +releaseVersion)
		}
        }
	return nil
}

func getRequiredTestsForRelease(release string) []string{
	requiredTests := []string {"itest"}
	return requiredTests
}
func getSubmittedTestsFromPullReq(prLogger *logrus.Entry, ghc githubClient, org,repo string, prNumber int, sha string) []string{
	submittedTests := []string {"itest"}
	return submittedTests
}
func checkAllSubmittedTestsArePresent(required,submitted []string) bool {
	allTestsPresent := false
	return allTestsPresent
}
func checkE2eLogHasEvidenceOfTestRuns (prLogger *logrus.Entry, ghc githubClient, org,repo string, prNumber int, sha string, requiredTests []string) (bool,error) {
	allEvidencePresent := false
	return allEvidencePresent, nil
}
// TODO Consolodate this and the next function to cerate a map of labels
func HasNotVerifiableLabel(prLogger *logrus.Entry, org,repo string, prNumber int, ghc githubClient) (bool,error) {
        hasNotVerifiableLabel := false
	labels, err := ghc.GetIssueLabels(org, repo, prNumber)

        if err != nil {
                prLogger.WithError(err).Error("Failed to find labels")
        }

        for foundLabel := range labels {
                notVerifiableCheck := strings.Compare(labels[foundLabel].Name,"not-verifiable")
                if notVerifiableCheck == 0 {
			hasNotVerifiableLabel = true
                        break
                }
        }

        return hasNotVerifiableLabel, err
}
func getReleaseFromLabel(prLogger *logrus.Entry, org,repo string, prNumber int, ghc githubClient) (string) {
	release := ""
	hasRelease := false
	labels, err := ghc.GetIssueLabels(org, repo, prNumber)

        if err != nil {
                prLogger.WithError(err).Error("GetReleaseLabel : Failed to find labels")
        }

        for _, foundLabel := range labels {
		// I had error that release was being declared but not used. I see you changed line 225?
		// What did I miss? I need to figure out foundLabel. Can we spend a few minutes reviewing?
		hasRelease, release = findRelease(prLogger, foundLabel.Name)
		if hasRelease {
			break
		}
        }

        return release
}

func findRelease(log *logrus.Entry, word string)  (bool, string) {
        hasRelease := false
        k8sRelease := ""
        k8sVerRegExp := regexp.MustCompile(`v[0-9]\.[0-9][0-9]*`)
        containsVersion, err := regexp.MatchString(`v[0-9]\.[0-9][0-9]*`, word)
        if err != nil {
                log.WithError(err).Error("Error matching k8s version in %s",word)
        }
        if (containsVersion) {
                k8sRelease = k8sVerRegExp.FindString(word)
                log.WithFields(logrus.Fields{
                        "Version": k8sRelease,
                })
                hasRelease = true
        }
        return hasRelease, k8sRelease
}

// takes a patchUrl from a githubClient.PullRequestChange and transforms it
// to produce the url that delivers the raw file associated with the patch.
// Tested for small files.
func patchUrlToFileUrl(patchUrl string) (string){
	fileUrl := strings.Replace(patchUrl, "github.com", "raw.githubusercontent.com", 1)
	fileUrl = strings.Replace(fileUrl, "/blob", "", 1)
        return fileUrl
}

// Retrieves e2eLogfile and checks that it contains k8sRelease
func checkE2eLogHasRelease(log *logrus.Entry, e2eChange github.PullRequestChange, k8sRelease string) (bool){
        e2eLogHasStatedRelease := false

        fileUrl := patchUrlToFileUrl(e2eChange.BlobURL)
	resp, err := http.Get(fileUrl)
	if err != nil {
		log.Errorf("cELHR : %+v",err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)


	// Make a set that contains all the key fields in the Product YAML file
        // TODO Check to see if string(body) performant
	for _, line := range strings.Split(string(body), "\n") {
                if strings.Contains(line, k8sRelease){
                        log.Infof("cELHR found stated release!! %s",line)
                        e2eLogHasStatedRelease = true
                        break
                }
        }
        return e2eLogHasStatedRelease

}

// Retrves conformance.yaml so we can create a map from it
func createMapOfRequirements(log *logrus.Entry,  k8sRelease string) (bool){
        e2eLogHasStatedRelease := false

        return e2eLogHasStatedRelease

}

// Executes the search query contained in q using the GitHub client ghc
func search(ctx context.Context, log *logrus.Entry, ghc githubClient, q string) ([]PullRequest, error) {
	var ret []PullRequest
	vars := map[string]interface{}{
		"query":        githubql.String(q),
		"searchCursor": (*githubql.String)(nil),
	}
	var totalCost int
	var remaining int
	for {
		sq := SearchQuery{}
		if err := ghc.Query(ctx, &sq, vars); err != nil {
			return nil, err
		}
		totalCost += int(sq.RateLimit.Cost)
		remaining = int(sq.RateLimit.Remaining)
		for _, n := range sq.Search.Nodes {
			ret = append(ret, n.PullRequest)
		}
		if !sq.Search.PageInfo.HasNextPage {
			break
		}
		vars["searchCursor"] = githubql.NewString(sq.Search.PageInfo.EndCursor)
	}
	log.Infof("Search for query \"%s\" cost %d point(s). %d remaining.", q, totalCost, remaining)
	return ret, nil
}

type PullRequest struct {
	Number githubql.Int
	Author struct {
		Login githubql.String
	}
	Repository struct {
		Name  githubql.String
		Owner struct {
			Login githubql.String
		}
	}
	Labels struct {
		Nodes []struct {
			Name githubql.String
		}
	} `graphql:"labels(first:100)"`
        Files struct {
                Nodes []struct {
                        Path githubql.String
                }
	} `graphql:"files(first:10)"`
	Title githubql.String
	Commits struct {
		Nodes []struct {
			Commit struct {
				Oid githubql.String
			}
		}
	} `graphql:"commits(first:5)"`
}

type SearchQuery struct {
	RateLimit struct {
		Cost      githubql.Int
		Remaining githubql.Int
	}
	Search struct {
		PageInfo struct {
			HasNextPage githubql.Boolean
			EndCursor   githubql.String
		}
		Nodes []struct {
			PullRequest PullRequest `graphql:"... on PullRequest"`
		}
	} `graphql:"search(type: ISSUE, first: 100, after: $searchCursor, query: $query)"`
}
