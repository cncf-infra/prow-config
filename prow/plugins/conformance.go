/*
Copyright 2018 Rob and Berno


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

package conformance

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"regexp"

	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/labels"
	"k8s.io/test-infra/prow/pluginhelp"
	"k8s.io/test-infra/prow/plugins"
)

const (
	pluginName             = "conformance"
	conformanceContextName         = "conformance/cncf"
	conformanceNoVersionMessage = `Thanks for your pull request. Before we can look at your pull request, you'll need to ensure that your PR is accurate and internally consistent.

:memo: **Please follow instructions at <https://github.com/cncf/k8s-conformance/blob/master/instructions.md#running> **

<!-- need_version_in_pr_title -->

<details>
%s
</details>
	`
	maxRetries = 5
)

var (
	// checkVersionRe = regexp.MustCompile(`(.*)v[0-9].[0-9][0-9]\s*$`)
	checkVersionRe = regexp.MustCompile(`(.*)v(.*)$`)
)

func init() {
	plugins.RegisterStatusEventHandler(pluginName, handleStatusEvent, helpProvider)
	plugins.RegisterGenericCommentHandler(pluginName, handleCommentEvent, helpProvider)
}

func helpProvider(config *plugins.Configuration, _ []config.OrgRepo) (*pluginhelp.PluginHelp, error) {
	// The {WhoCanUse, Usage, Examples, Config} fields are omitted because this plugin cannot be
	// manually triggered and is not configurable.
	pluginHelp := &pluginhelp.PluginHelp{
		Description: "The conformance plugin manages requests for CNCR Certification of a Kubernetes ditrisbution "
	}
	pluginHelp.AddCommand(pluginhelp.Command{
		Usage:       "/check-conf",
		Description: "Forces rechecking of the Certification PR.",
		Featured:    true,
		WhoCanUse:   "Anyone",
		Examples:    []string{"/check-request"},
	})
	return pluginHelp, nil
}

type gitHubClient interface {
	CreateComment(owner, repo string, number int, comment string) error
	AddLabel(owner, repo string, number int, label string) error
	RemoveLabel(owner, repo string, number int, label string) error
	GetPullRequest(owner, repo string, number int) (*github.PullRequest, error)
	FindIssues(query, sort string, asc bool) ([]github.Issue, error)
	GetIssueLabels(org, repo string, number int) ([]github.Label, error)
	GetCombinedStatus(org, repo, ref string) (*github.CombinedStatus, error)
}

func handleStatusEvent(pc plugins.Agent, se github.StatusEvent) error {
	return handle(pc.GitHubClient, pc.Logger, se)
}

// 1. Check that the status event received from the webhook is for the CNCF-CLA.
// 2. Use the github search API to search for the PRs which match the commit hash corresponding to the status event.
// 3. For each issue that matches, check that the PR's HEAD commit hash against the commit hash for which the status
//    was received. This is because we only care about the status associated with the last (latest) commit in a PR.
// 4. Set the corresponding CLA label if needed.
func handle(gc gitHubClient, log *logrus.Entry, se github.StatusEvent) error {

	log.Info("Status event is %s",se.State)
	if se.State == "" || se.Context == "" {
		return fmt.Errorf("invalid status event delivered with empty state/context")
	}

	if se.Context != conformanceContextName {
		// Not the CNCF CLA context, do not process this.
		return nil
	}

	if se.State == github.StatusPending {
		// do nothing and wait for state to be updated.
		return nil
	}

	org := se.Repo.Owner.Login
	repo := se.Repo.Name
	log.Info("%s/%s Searching for PRs matching the commit.", org,repo)
	// hunting for issues  feels like overkil for our use case
	// we only really want to check PR contents
	// We may have put on the wrong trousers
	var issues []github.Issue
	var err error
	for i := 0; i < maxRetries; i++ {
		issues, err = gc.FindIssues(fmt.Sprintf("%s repo:%s/%s type:pr state:open", se.SHA, org, repo), "", false)
		if err != nil {
			return fmt.Errorf("error searching for issues matching commit: %v", err)
		}
		if len(issues) > 0 {
			break
		}
		time.Sleep(10 * time.Second)
	}
	log.Infof("Found %d PRs matching commit.", len(issues))

	return nil
}

func handleCommentEvent(pc plugins.Agent, ce github.GenericCommentEvent) error {
	return handleComment(pc.GitHubClient, pc.Logger, &ce)
}

func handleComment(gc gitHubClient, log *logrus.Entry, e *github.GenericCommentEvent) error {
	var org, repo string
	var number int
	// Only consider open PRs and new comments.
	if e.IssueState != "open" || e.Action != github.GenericCommentActionCreated {
		return nil
	}
	org = e.Repo.Owner.Login
	repo = e.Repo.Name
	number = e.Number
	// Only consider "/check-cla" comments.
	pr = gc.GetPullRequest(owner , repo , number)
	if !checkVersionRe.MatchString(pr) {
		return nil
	}

	hasCLAYes := false
	hasCLANo := false

	// Check for existing cla labels.
	issueLabels, err := gc.GetIssueLabels(org, repo, number)
	if err != nil {
		log.WithError(err).Errorf("Failed to get the labels on %s/%s#%d.", org, repo, number)
	}
	for _, candidate := range issueLabels {
		if candidate.Name == labels.ClaYes {
			hasCLAYes = true
		}
		// Could theoretically have both yes/no labels.
		if candidate.Name == labels.ClaNo {
			hasCLANo = true
		}
	}

	pr, err := gc.GetPullRequest(org, repo, e.Number)
	if err != nil {
		log.WithError(err).Errorf("Unable to fetch PR-%d from %s/%s.", e.Number, org, repo)
	}

	// Check for the cla in past commit statuses, and add/remove corresponding cla label if necessary.
	ref := pr.Head.SHA
	combined, err := gc.GetCombinedStatus(org, repo, ref)
	if err != nil {
		log.WithError(err).Errorf("Failed to get statuses on %s/%s#%d", org, repo, number)
	}

	return nil
}
