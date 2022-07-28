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
	PRSuiteOptions  PRSuiteOptions
	ExpectedComment string
	ExpectedLabels  []string
}

func TestNewTestSuite(t *testing.T) {
	prSuiteOptionsDefault := PRSuiteOptions{
		Paths: []string{"../../kodata/features/"},
	}
	cases := []testSuiteCase{
		{
			PullRequest:     &PullRequest{},
			PRSuiteOptions:  prSuiteOptionsDefault,
			ExpectedComment: ``,
			ExpectedLabels:  []string{"conformance-product-submission", "release-documents-checked"},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			s := NewPRSuite(c.PullRequest)
			s.NewTestSuite(c.PRSuiteOptions).Run()
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
