package suite

import (
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

func TestNewTestSuite(t *testing.T) {
	s := NewPRSuite(&PullRequest{}).NewTestSuite(PRSuiteOptions{})
	if s.Name != "how-are-the-prs" {
		t.Fatalf("Unknown name")
	}
}
