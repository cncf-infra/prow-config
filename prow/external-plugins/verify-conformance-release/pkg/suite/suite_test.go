package suite

import (
	"fmt"
	"testing"
)

func TestNewPRSuite(t *testing.T) {
	es := NewPRSuite(&PullRequest{
		PullRequestQuery: PullRequestQuery{},
		SupportingFiles:  []*PullRequestFile{},
	})
	s := NewPRSuite(&PullRequest{
		PullRequestQuery: PullRequestQuery{
			Title: "Conformance results for v1.23/cool",
		},
		SupportingFiles: []*PullRequestFile{
			&PullRequestFile{
				Name:     "v1.23/cool/README.md",
				BaseName: "README.md",
				BlobURL:  "https://github.com/cncf-infra/k8s-conformance/raw/2c154f2bd6f0796c4d65f5b623c347b6cc042e59/v1.23/cke/README.md",
				Contents: `# Conformance test for Something`,
			},
		},
	})
	if s.Labels[0] != "conformance-product-submission" {
		t.Fatalf("Label not found")
	}

	newMetadataFolder := "a"
	s.SetMetadataFolder(newMetadataFolder)
	if s.MetadataFolder != newMetadataFolder {
		t.Fatalf("MetadataFolder not set")
	}

	if err := es.thePRTitleIsNotEmpty(); err == nil {
		t.Fatalf("Title should be empty")
	}
	if err := s.thePRTitleIsNotEmpty(); err != nil {
		t.Fatalf("Title error: %v", err)
	}

	if err := s.isIncludedInItsFileList("README.md"); err != nil {
		t.Fatalf("Failed to find file")
	}

	if err := s.isIncludedInItsFileList("README.m"); err == nil {
		t.Fatalf("Shouldn't have found this file")
	}

	if err := s.fileFolderStructureMatchesRegex("^(v1.[0-9]{2})/(.*)$"); err != nil {
		t.Fatalf("Failed to match the folder structure, %v", err)
	}

	if err := s.fileFolderStructureMatchesRegex("verify-conformance-bot-lol"); err == nil {
		t.Fatalf("Folder structure is invalid")
	}

	if err := s.fileFolderStructureMatchesRegex("()/()"); err == nil {
		t.Fatalf("Folder structure is invalid")
	}

	if err := s.thereIsOnlyOnePathOfFolders(); err != nil {
		t.Fatalf("Folder shouldn't have more than one directory structure, %v", err)
	}
}

type testSuiteCase struct {
	Name            string
	PullRequest     *PullRequest
	ExpectedComment string
	ExpectedLabels  []string
}

func TestNewTestSuite(t *testing.T) {
	prSuiteOptions := PRSuiteOptions{
		Paths: []string{"../../kodata/features/"},
	}
	cases := []testSuiteCase{
		{
			Name:        "empty pr",
			PullRequest: &PullRequest{},
			ExpectedComment: `0 of 14 requirements have passed. Please review the following:
- [FAIL] it seems that there is no title set
  - title is empty
- [FAIL] there seems to be some required files missing (https://github.com/cncf/k8s-conformance/blob/master/instructions.md#contents-of-the-pr)
  - missing file &#39;README.md&#39;
  - missing file &#39;PRODUCT.yaml&#39;
  - missing file &#39;e2e.log&#39;
  - missing file &#39;junit_01.xml&#39;
- [FAIL] the submission file directory does not seem to match the Kubernetes release version in the files
  - there were no files found in the submission
- [FAIL] the submission seems to contain files of multiple Kubernetes release versions or products. Each Kubernetes release version and products should be submitted in a separate PRs
  - there were no files found in the submission
- [FAIL] the title of the submission does not seem to contain a Kubernetes release version that matches the release version in the submitted files
  - there were no files found in the submission
- [FAIL] it appears that the PRODUCT.yaml file does not contain all the required fields (https://github.com/cncf/k8s-conformance/blob/master/instructions.md#productyaml)
  - missing required file &#39;PRODUCT.yaml&#39;
- [FAIL] it appears that URL(s) in the PRODUCT.yaml aren't correctly formatted URLs
  - missing required file &#39;PRODUCT.yaml&#39;
- [FAIL] it appears that URL(s) in the PRODUCT.yaml don't resolve to the correct data type
  - missing required file &#39;PRODUCT.yaml&#39;
- [FAIL] the submission title is missing either a Kubernetes release version (v1.xx) or product name
  - title is empty
- [FAIL] it seems the e2e.log does not contain the Kubernetes release version that match the submission title
  - missing required file &#39;e2e.log&#39;
- [FAIL] the Kubernetes release version in this pull request does not qualify for conformance submission anymore (https://github.com/cncf/k8s-conformance/blob/master/terms-conditions/Certified_Kubernetes_Terms.md#qualifying-offerings-and-self-testing)
  - unable to find a Kubernetes release version in the title
- [FAIL] it appears that some tests are missing from the product submission
  - missing required file &#39;junit_01.xml&#39;
- [FAIL] it appears that some tests failed in the product submission
  - missing required file &#39;e2e.log&#39;
- [FAIL] it appears that there is a mismatch of tests in junit_01.xml and e2e.log
  - missing required file &#39;e2e.log&#39;

 for a full list of requirements, please refer to these sections of the docs: [_content of the PR_](https://github.com/cncf/k8s-conformance/blob/master/instructions.md#contents-of-the-pr), and [_requirements_](https://github.com/cncf/k8s-conformance/blob/master/instructions.md#requirements).
`,
			ExpectedLabels: []string{"conformance-product-submission", "release-documents-checked", "missing-file-README.md", "missing-file-PRODUCT.yaml", "missing-file-e2e.log", "missing-file-junit_01.xml", "not-verifiable"},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			s := NewPRSuite(c.PullRequest)
			s.NewTestSuite(prSuiteOptions).Run()
			comment, labels, _ := s.GetLabelsAndCommentsFromSuiteResultsBuffer()
			if comment != c.ExpectedComment {
				fmt.Println(comment)
				t.Fatalf("Comment '%v' expected to match '%v'", comment, c.ExpectedComment)
			}
			missingLabels := []string{}
			for _, l := range labels {
				found := false
				for _, lr := range c.ExpectedLabels {
					if l == lr {
						found = true
					}
				}
				if found != true {
					missingLabels = append(missingLabels, l)
				}
			}
			if len(missingLabels) > 0 {
				t.Fatalf("Labels missing from PR: %v", missingLabels)
			}
		})
	}
}
