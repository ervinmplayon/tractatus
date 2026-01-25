package github

import "strings"

type Detector struct{}

func NewDetector() *Detector {
	return &Detector{}
}

// CI/CD platform indicators (root level only, PERHAPS rethink because what if its not root?)
// For files, match exactly; for directories, match prefix
var cicdFiles = map[string]string{
	".circleci":               "CircleCI",
	".github/workflows":       "GitHub Actions",
	".github":                 "GitHub Actions",
	"bitbucket-pipelines.yml": "Bitbucket Pipelines",
	".gitlab-ci.yml":          "GitLab CI",
	"Jenkinsfile":             "Jenkins",
	".travis.yml":             "Travis CI",
	"azure-pipelines.yml":     "Azure Pipelines",
}

// Test directory patterns
var testDirs = []string{
	"test",
	"tests",
	"__tests__",
	"spec",
	"test_suite",
	"testing",
}

// EKS platform indicators (if found, skip the repo)
var eksIndicators = []string{
	"k8s",
	"kubernetes",
	".kube",
	"helm",
	"Chart.yaml",
	"kustomization.yaml",
	"kustomization.yml",
}

// Platform-specific files
var ecsIndicators = []string{
	"ecs-task-definition.json",
	"ecs-service.json",
	"Dockerfile",
}

var lambdaIndicators = []string{
	"serverless.yml",
	"serverless.yaml",
	"template.yaml",
	"template.yml", // SAM
	"lambda",       // directory
	"functions",    // directory
}

var beanstalkIndicators = []string{
	".ebextensions/",
	"Procfile",
	".elasticbeanstalk",
}

// Checks for CI/CD configuration files at root level
func (d *Detector) DetectCICD(files []string) (bool, string) {
	for _, file := range files {
		for pattern, platform := range cicdFiles {
			if file == pattern || strings.HasPrefix(file, pattern) {
				return true, platform
			}

			// Directory match (for .circleci, .github)
			// Files list contains directory names without trailing slash
			if strings.HasPrefix(pattern, file+"/") || file == strings.TrimSuffix(pattern, "/") {
				return true, platform
			}
		}
	}
	return false, ""
}

// Checks for test directories or files
func (d *Detector) DetectTests(files []string) (bool, string) {
	for _, file := range files {
		for _, testDir := range testDirs {
			// Directory match (without trailing slash in files list)
			if file == testDir {
				return true, "detected test directory"
			}
		}
	}
	// Check for common test file patterns
	testPatterns := []string{
		"_test.go",
		".spec.js",
		".test.js",
		".spec.ts",
		".test.ts",
		"Test.java",
		"test_",
	}

	for _, file := range files {
		for _, pattern := range testPatterns {
			if strings.Contains(file, pattern) {
				return true, "detected test files"
			}
		}
	}

	return false, ""
}

// Checks if the repo is an EKS application
func (d *Detector) IsEKS(files []string) bool {
	for _, file := range files {
		for _, indicator := range eksIndicators {
			// Check both exact match and directory name
			if file == indicator || file == strings.TrimSuffix(indicator, "/") {
				return true
			}
		}
	}
	return false
}

// Determines the deployment platform
func (d *Detector) DetectPlatform(files []string) string {
	platforms := []string{}

	// Check ECS
	for _, file := range files {
		for _, indicator := range ecsIndicators {
			if file == indicator {
				platforms = append(platforms, "ECS")
				break
			}
		}
	}
	// Check Lambda
	for _, file := range files {
		for _, indicator := range lambdaIndicators {
			if file == indicator {
				platforms = append(platforms, "Lambda")
				break
			}
		}
	}

	// Check Elastic Beanstalk
	for _, file := range files {
		for _, indicator := range beanstalkIndicators {
			if file == indicator {
				platforms = append(platforms, "Elastic Beanstalk")
				break
			}
		}
	}

	if len(platforms) == 0 {
		return "Unknown"
	}

	// Return comma-separated if multiple platforms detected
	return strings.Join(platforms, ", ")
}

// Checks if CODEOWNERS file exists
func (d *Detector) DetectCodeOwners(files []string) bool {
	codeownersFiles := []string{
		"CODEOWNERS",
		".github/CODEOWNERS",
		"docs/CODEOWNERS",
	}

	for _, file := range files {
		for _, codeownerFile := range codeownersFiles {
			if file == codeownerFile {
				return true
			}
		}
	}
	return false
}

// Extracts team/owner information from codeowners content
func (d *Detector) ParseCodeOwners(content string) []string {
	var owners []string
	ownerSet := make(map[string]bool) // Track unique owners
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// CODEOWNERS format: path @owner1 @owner2 email@example.com
		parts := strings.Fields(line)
		if len(parts) > 1 {
			// Skip the first part (it's the file pattern like * or /src/*)
			for _, part := range parts[1:] {
				var owner string

				if strings.HasPrefix(part, "@") {
					// GitHub username: @username or @org/team
					owner = strings.TrimPrefix(part, "@")
				} else if strings.Contains(part, "@") {
					// Email address: user@domain.com
					owner = part
				} else {
					// Skip non-owner parts (like file patterns)
					continue
				}

				// Add to set (deduplicates automatically)
				if !ownerSet[owner] {
					ownerSet[owner] = true
					owners = append(owners, owner)
				}
			}
		}
	}

	return owners
}
