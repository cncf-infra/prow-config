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
	"context"
	"regexp"
	"strings"
	"time"
	//"fmt"
	githubql "github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"
	"k8s.io/test-infra/prow/plugins"
	"net/http"
	// 	"github.com/golang-collections/collections/set"
	"encoding/xml"
	"github.com/hashicorp/go-version"
	"path"
)

const (
	PluginName = "verify-conformance-tests"
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

type ConformanceTestMetaData struct {
	Testname    string `yaml:"testname"`
	Codename    string `yaml:"codename"`
	Description string `yaml:"description"`
	Release     string `yaml:"release"`
	File        string `yaml:"file"`
}

func getRequiredTests(log *logrus.Entry, k8sRelease string) map[string]bool {
	// TODO we are effectively hardcoding this and we may layer this out
	// Key'd by k8s release map that points to URLs containing the required conformance tests for that release

	//var requiredConformanceSuite []ConformanceTestMetaData
	//confTestSuiteUrl := "https://raw.githubusercontent.com/kubernetes/kubernetes/master/test/conformance/testdata/conformance.yaml"

	var conformanceTests = map[string]string{
		"v1.15":  "https://raw.githubusercontent.com/cncf-infra/prow-config/master/docs/conformance_v1.15.yaml",
		"v1.16":  "https://raw.githubusercontent.com/cncf-infra/prow-config/master/docs/conformance_v1.16.yaml",
		"v1.17":  "https://raw.githubusercontent.com/cncf-infra/prow-config/master/docs/conformance_v1.17.yaml",
		"v1.18":  "https://raw.githubusercontent.com/cncf-infra/prow-config/master/tests/conformance-1.18.yaml",
		"v1.19":  "https://raw.githubusercontent.com/kubernetes/kubernetes/release-1.19/test/conformance/testdata/conformance.yaml",
		"v1.20":  "https://raw.githubusercontent.com/kubernetes/kubernetes/release-1.20/test/conformance/testdata/conformance.yaml",
		"v1.21":  "https://raw.githubusercontent.com/kubernetes/kubernetes/release-1.21/test/conformance/testdata/conformance.yaml",
		"v1.22":  "https://raw.githubusercontent.com/kubernetes/kubernetes/release-1.22/test/conformance/testdata/conformance.yaml",
		"v1.23":  "https://raw.githubusercontent.com/kubernetes/kubernetes/release-1.23/test/conformance/testdata/conformance.yaml",
		"master": "https://raw.githubusercontent.com/kubernetes/kubernetes/master/test/conformance/testdata/conformance.yaml",
	}

	var requiredConformanceSuite []ConformanceTestMetaData
	confTestSuiteUrl := conformanceTests[k8sRelease]

	resp, err := http.Get(confTestSuiteUrl)
	if resp.StatusCode > 199 && resp.StatusCode < 300 {
		// TODO check body for 404
		if err != nil {
			log.Errorf("Error retrieving conformance tests metadata from : %s", confTestSuiteUrl)
			log.Errorf("HTTP Response was: %+v", resp)
			log.Errorf("getRequiredTests : %+v", err)
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body) // TODO Handle err

		err = yaml.Unmarshal(body, &requiredConformanceSuite)
		if err != nil {
			log.Errorf("Cannot unmarshal data. Reason:  %v\n", err)
		}
	}
	requiredConformanceSuiteForRelease := make(map[string]bool, 0)

	for _, testcase := range requiredConformanceSuite {
		if k8sRelease != "" {
			v1, err := version.NewVersion(k8sRelease)
			if err != nil {
				log.Errorf("Unable to parse the version string %v:  %v\n", k8sRelease, err)
			}
			testcaseRelease := strings.Split(testcase.Release, ",")
			if testcaseRelease[0] != "" {
				// testcase.Codename = strings.Replace(testcase.Codename, "\"", "`", -1)
				//                log.Println(testcase)
				v2, err := version.NewVersion(testcaseRelease[0])
				if err != nil {
					log.Errorf("Unable to parse the testcase version string  %v:  %v\n", testcase, err)
					continue
				}
				if v1.GreaterThanOrEqual(v2) {

					requiredConformanceSuiteForRelease[testcase.Codename] = false
				}
			}
		}
	}
	return requiredConformanceSuiteForRelease
}

// HandleAll checks verifiable certification pull requests to insure that all required
// tests for the stated k8s release (e.g. release-1.19) have been executed and have passed.
// the following labels will be added depending on the outcome of checking the tests
// TODO add labels
func HandleAll(log *logrus.Entry, ghc githubClient, config *plugins.Configuration) error {
	var queryString = "archived:false is:pr is:open label:release-documents-checked"
	//	var queryString = "1026 repo:\"cncf-infra/k8s-conformance\""
	pullRequests, err := getPullRequests(log, ghc, queryString)
	if err != nil {
		log.Error(err)
	}

	for _, pr := range pullRequests {

		org := string(pr.Repository.Owner.Login)
		repo := string(pr.Repository.Name)
		prNumber := int(pr.Number)

		releaseVersion := getReleaseFromLabel(log, org, repo, prNumber, ghc)
		changes, _ := getChangeMap(ghc, org, repo, prNumber)

		// Add fields from this PR to logger
		prLogger := log.WithFields(logrus.Fields{"pr": prNumber, "title": pr.Title, "release": releaseVersion})

		requiredTests := getRequiredTests(prLogger, releaseVersion) // retrieves the conformance.yaml for this release
		submittedTests, err := getSubmittedConformanceTests(prLogger, changes["junit_01.xml"])
		if err != nil {
			prLogger.WithError(err)
		}
		submittedTestsPresent, missingTests := checkAllRequiredTestsArePresent(prLogger, requiredTests, submittedTests)
		prLogger.Infof("submittedTestsPresent %t : missingTests are %v\n", submittedTestsPresent, missingTests)

		if submittedTestsPresent {
			testRunEvidenceCorrect, err := checkE2eLogHasZeroTestFailures(prLogger, changes["e2e.log"])
			hasVerifiedLabel, err := HasVerifiedLabel(log, org, repo, prNumber, ghc, "tests-verified-"+releaseVersion)
			if err != nil {
				prLogger.WithError(err)
			}
			if !hasVerifiedLabel {
				hasNoEvidenceMissingLabel, _ := HasNoEvidenceMissingLabel(log, org, repo, prNumber, ghc, "evidence-missing")
				hasNoRequiredTestsMissingLabel, _ := HasNoRequiredTestsMissingLabel(log, org, repo, prNumber, ghc, "required-tests-missing")
				githubClient.AddLabel(ghc, org, repo, prNumber, "tests-verified-"+releaseVersion)
				githubClient.CreateComment(ghc, org, repo, prNumber, "Automatically verified as having all required tests present and passed")
				if hasNoRequiredTestsMissingLabel {
					githubClient.RemoveLabel(ghc, org, repo, prNumber, "required-tests-missing")
				}
				if hasNoEvidenceMissingLabel {
					githubClient.RemoveLabel(ghc, org, repo, prNumber, "evidence-missing")
				}
			}
			// Versions v1.16 and older do not log failures in e2e log, so the testRunEvidenceCorrect check is skipped for those older versions
			v1, err := version.NewVersion("v1.17")
			if err != nil {
				log.Errorf("Unable to set version to 1.17 \n")
			}
			v2, err := version.NewVersion(releaseVersion)
			if err != nil {
				log.Errorf("Unable to parse the version string %v:  %v\n", releaseVersion, err)
			}

			if v2.GreaterThanOrEqual(v1) {
				if testRunEvidenceCorrect {
					hasNoTestFaiLabel, err := HasNoTestFailLabel(log, org, repo, prNumber, ghc, "no-failed-tests-"+releaseVersion)
					if err != nil {
						prLogger.WithError(err)
					}

					if !hasNoTestFaiLabel {
						githubClient.AddLabel(ghc, org, repo, prNumber, "no-failed-tests-"+releaseVersion)
						githubClient.CreateComment(ghc, org, repo, prNumber, "Automatically verified as having all required tests present and passed")
					}
				} else { // specifiedRelease not present in logs

					hasNoEvidenceMissingLabel, err := HasNoEvidenceMissingLabel(log, org, repo, prNumber, ghc, "evidence-missing")
					if err != nil {
						prLogger.WithError(err)
					}

					if !hasNoEvidenceMissingLabel {
						githubClient.AddLabel(ghc, org, repo, prNumber, "evidence-missing")
						githubClient.CreateComment(ghc, org, repo, prNumber,
							"This conformance request has the correct list of tests present in the junit file but at least one of the tests in e2e.log failed")
					}
				}
			}
		} else {
			hasNoRequiredTestsMissingLabel, err := HasNoRequiredTestsMissingLabel(log, org, repo, prNumber, ghc, "required-tests-missing")
			if err != nil {
				prLogger.WithError(err)
			}
			if !hasNoRequiredTestsMissingLabel {

				githubClient.AddLabel(ghc, org, repo, prNumber, "required-tests-missing")
				githubClient.CreateComment(ghc, org, repo, prNumber,
					"This conformance request failed to include all of the required tests for "+releaseVersion)

				githubClient.CreateComment(ghc, org, repo, prNumber, "The first test found to be missing was "+missingTests[0])
			}
		}
	}
	return nil
}

// hasNoRequiredTestsMissingLabel checks if the evidence-missing label has been set
func HasNoRequiredTestsMissingLabel(prLogger *logrus.Entry, org, repo string, prNumber int, ghc githubClient, verifiedLabel string) (bool, error) {
	hasReleaseLabel := false
	labels, err := ghc.GetIssueLabels(org, repo, prNumber)

	if err != nil {
		prLogger.WithError(err).Error("Failed to find labels")
	}

	for foundLabel := range labels {
		releaseCheck := strings.Compare(labels[foundLabel].Name, verifiedLabel)
		if releaseCheck == 0 {
			hasReleaseLabel = true
			break
		}
	}

	return hasReleaseLabel, err
}

// hasNoEvidenceMissingLabel checks if the evidence-missing label has been set
func HasNoEvidenceMissingLabel(prLogger *logrus.Entry, org, repo string, prNumber int, ghc githubClient, verifiedLabel string) (bool, error) {
	hasReleaseLabel := false
	labels, err := ghc.GetIssueLabels(org, repo, prNumber)

	if err != nil {
		prLogger.WithError(err).Error("Failed to find labels")
	}

	for foundLabel := range labels {
		releaseCheck := strings.Compare(labels[foundLabel].Name, verifiedLabel)
		if releaseCheck == 0 {
			hasReleaseLabel = true
			break
		}
	}

	return hasReleaseLabel, err
}

// hasNoTestFailLabel checks if the noTestFail-releaseVersion label has been set
func HasNoTestFailLabel(prLogger *logrus.Entry, org, repo string, prNumber int, ghc githubClient, verifiedLabel string) (bool, error) {
	hasReleaseLabel := false
	labels, err := ghc.GetIssueLabels(org, repo, prNumber)

	if err != nil {
		prLogger.WithError(err).Error("Failed to find labels")
	}

	for foundLabel := range labels {
		releaseCheck := strings.Compare(labels[foundLabel].Name, verifiedLabel)
		if releaseCheck == 0 {
			hasReleaseLabel = true
			break
		}
	}

	return hasReleaseLabel, err
}

// hasVerifiedLabel checks if the verified-releaseVersion has been set
func HasVerifiedLabel(prLogger *logrus.Entry, org, repo string, prNumber int, ghc githubClient, verifiedLabel string) (bool, error) {
	hasReleaseLabel := false
	labels, err := ghc.GetIssueLabels(org, repo, prNumber)

	if err != nil {
		prLogger.WithError(err).Error("Failed to find labels")
	}

	for foundLabel := range labels {
		releaseCheck := strings.Compare(labels[foundLabel].Name, verifiedLabel)
		if releaseCheck == 0 {
			hasReleaseLabel = true
			break
		}
	}

	return hasReleaseLabel, err
}

// getPullRequests sends a github query to retrieve an array of PullRequest
func getPullRequests(log *logrus.Entry, ghc githubClient, queryString string) ([]PullRequest, error) {

	pullRequests, err := prSearch(context.Background(), log, ghc, queryString)

	if err != nil {
		return nil, err
	}

	log.Infof("Considering %d verifiable PRs.", len(pullRequests))

	return pullRequests, nil
}

// getSubmittedConformanceTests returns an array of test names that are tagged as [Conformance]
// in the junit_01.xml file submitted by the vendor in the changes associated with the certification request PR
func getSubmittedConformanceTests(prLogger *logrus.Entry, junitFile github.PullRequestChange) ([]string, error) {

	jUnitUrl := patchUrlToFileUrl(junitFile.BlobURL)

	resp, err := http.Get(jUnitUrl)
	if err != nil {
		prLogger.Errorf("gSTTP: %#v", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	type TestCase struct {
		XMLName xml.Name  `xml:"testcase"`
		Name    string    `xml:"name,attr"`
		Skipped *struct{} `xml:"skipped"`
	}

	var conformanceRequirement struct {
		TestSuite []TestCase `xml:"testcase"`
	}

	if err := xml.Unmarshal(body, &conformanceRequirement); err != nil {
		prLogger.Fatal(err)
	}

	submittedTestMap := make([]string, len(conformanceRequirement.TestSuite))
	for _, testcase := range conformanceRequirement.TestSuite {
		if testcase.Skipped == nil {
			if strings.Contains(testcase.Name, "[Conformance]") {
				testcase.Name = strings.Replace(testcase.Name, "&#39;", "'", -1)
				testcase.Name = strings.Replace(testcase.Name, "&#34;", "\"", -1)
				testcase.Name = strings.Replace(testcase.Name, "&gt;", ">", -1)
				testcase.Name = strings.Replace(testcase.Name, "'cat /tmp/health'", "\"cat /tmp/health\"", -1)
				submittedTestMap = append(submittedTestMap, testcase.Name)
			}
		}
	}

	//fmt.Println(submittedTestMap)
	return submittedTestMap, nil
}

// getChangeMap returns a map of base filenames to the github.PullRequestChange and nil
// returns an err if there is a problem talking to Github
func getChangeMap(ghc githubClient, org, repo string, prNumber int) (map[string]github.PullRequestChange, error) {
	changes, err := ghc.GetPullRequestChanges(org, repo, prNumber)

	if err != nil {
		return nil, err
	}

	var supportingFiles = make(map[string]github.PullRequestChange)

	for _, change := range changes {
		// https://developer.github.com/v3/pulls/#list-pull-requests-files
		supportingFiles[path.Base(change.Filename)] = change
	}
	return supportingFiles, nil
}

// checkAllRequiredTestsArePresent returns true if the test array submitted by the vendor has all tests that
// are required for certification conformance, otherwise returns false and an array of missing tests.
func checkAllRequiredTestsArePresent(log *logrus.Entry, required map[string]bool, submitted []string) (bool, []string) {
	allTestsPresent := true
	localRequired := required
	missingTests := []string{}

	for _, test := range submitted {
		if _, found := localRequired[test]; found {
			localRequired[test] = true
			//		} else {
			// tempTestCountMap[test]++
		}
	}

	for test, val := range localRequired {
		if val != true {
			allTestsPresent = false
			missingTests = append(missingTests, test)
			log.Infof("Missing Test: i%s", test)
		}
	}
	return allTestsPresent, missingTests

}

// checkE2eLogHasZeroTestFailures returns true if the e2eLog has a zero count for failed tests
func checkE2eLogHasZeroTestFailures(log *logrus.Entry, e2eChange github.PullRequestChange) (bool, error) {
	zeroTestFailures := false
	e2eNoTestsFailed := "\"failed\":0"
	e2eMainTestSuite := "\"Test Suite completed\""

	fileUrl := patchUrlToFileUrl(e2eChange.BlobURL)
	resp, err := http.Get(fileUrl)
	if err != nil {
		log.Errorf("cELHR : %+v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	for _, line := range strings.Split(string(body), "\n") {
		if strings.Contains(line, e2eMainTestSuite) {
			if strings.Contains(line, e2eNoTestsFailed) {
				log.Infof("found evidence that no tests have failed %s", line)
				zeroTestFailures = true
				break
			}
		}
	}
	return zeroTestFailures, nil
}

// TODO Consolidate this and the next function to cerate a map of labels
func HasNotVerifiableLabel(prLogger *logrus.Entry, org, repo string, prNumber int, ghc githubClient) (bool, error) {
	hasNotVerifiableLabel := false
	labels, err := ghc.GetIssueLabels(org, repo, prNumber)

	if err != nil {
		prLogger.WithError(err).Error("Failed to find labels")
	}

	for foundLabel := range labels {
		notVerifiableCheck := strings.Compare(labels[foundLabel].Name, "not-verifiable")
		if notVerifiableCheck == 0 {
			hasNotVerifiableLabel = true
			break
		}
	}

	return hasNotVerifiableLabel, err
}
func getReleaseFromLabel(prLogger *logrus.Entry, org, repo string, prNumber int, ghc githubClient) string {
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

func findRelease(log *logrus.Entry, word string) (bool, string) {
	hasRelease := false
	k8sRelease := ""
	k8sVerRegExp := regexp.MustCompile(`v[0-9]\.[0-9][0-9]*`)
	containsVersion, err := regexp.MatchString(`v[0-9]\.[0-9][0-9]*`, word)
	if err != nil {
		log.WithError(err).Error("Error matching k8s version in %s", word)
	}
	if containsVersion {
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
func patchUrlToFileUrl(patchUrl string) string {
	fileUrl := strings.Replace(patchUrl, "github.com", "raw.githubusercontent.com", 1)
	fileUrl = strings.Replace(fileUrl, "/blob", "", 1)
	return fileUrl
}

// Retrieves e2eLogfile and checks that it contains k8sRelease
func checkE2eLogHasRelease(log *logrus.Entry, e2eChange github.PullRequestChange, k8sRelease string) bool {
	e2eLogHasStatedRelease := false

	fileUrl := patchUrlToFileUrl(e2eChange.BlobURL)
	resp, err := http.Get(fileUrl)
	if err != nil {
		log.Errorf("cELHR : %+v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// Make a set that contains all the key fields in the Product YAML file
	// TODO Check to see if string(body) performant
	for _, line := range strings.Split(string(body), "\n") {
		if strings.Contains(line, k8sRelease) {
			log.Infof("cELHR found stated release!! %s", line)
			e2eLogHasStatedRelease = true
			break
		}
	}
	return e2eLogHasStatedRelease

}

// Retrves conformance.yaml so we can create a map from it
func createMapOfRequirements(log *logrus.Entry, k8sRelease string) bool {
	e2eLogHasStatedRelease := false

	return e2eLogHasStatedRelease

}

// prSearch executes a search query q using GitHub client ghc to find PullRequests that match the query in q
// return array of PullRequests found and err is set if there is a problem
func prSearch(ctx context.Context, log *logrus.Entry, ghc githubClient, q string) ([]PullRequest, error) {
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
	log.Infof("Search for query \"%s\" cost %d point(s). %d remaining. ", q, totalCost, remaining)
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
	Title   githubql.String
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
