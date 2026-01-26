# Tractatus: AWS CLI Tool

## Table of Contents
- [Getting Started](#getting-started)  
- [Output Example](#output-example)  
- [Project Structure](#project-structure)
- [FAQ](#faq)


## Getting Started
**Build**
```bash
go mod download
go build -o tractatus ./cmd/main.go
```

**Run**
```bash
# Run with GitHub (default)
export GITHUB_TOKEN=<token_here>
./tractatus --github-org org-name

# Save to markdown
./tractatus --github-org org-name --format markdown --output repos.md

# AWS still works too
./tractatus --source aws --account production
```

## Output Example
```bash
| Repo Name    | Owner     | Last Committer | CODEOWNERS | Platform | CI/CD     | Tests      |
|--------------|-----------|----------------|------------|----------|-----------|------------|
| unity-api    | dashbirds | a.danger  | Yes (2)    | ECS      | CircleCI  | Yes (pytest)|
| web-frontend | web-team  | v.myers   | No         | Unknown  | GitHub Actions | Yes     |
| monorepo     | mono-team | j.biden   | No         | Unknown  | GitHub Actions | Yes     |
| ace          | ace-team  | b.lesnar  | No         | Unknown  | GitHub Actions | Yes     |
```

## Project Structure
```bash
tractatus/
├── cmd/
│   └── main.go                    ← Updated with GitHub support
├── internal/
│   ├── sources/
│   │   ├── source.go             ← DataSource interface
│   │   ├── github/
│   │   │   ├── client.go         ← GitHub API wrapper
│   │   │   ├── detector.go       ← Multi-signal detection
│   │   │   └── source.go         ← GitHub DataSource impl
│   │   └── aws/
│   │       ├── client.go         ← AWS API (existing)
│   │       └── source.go         ← AWS DataSource impl
│   ├── inventory/
│   │   └── collector.go          ← Unified collector
│   └── output/
│       ├── table.go              ← Updated for GitHub fields
│       └── markdown.go           ← Updated for GitHub fields
└── go.mod                         ← Added GitHub libraries
```

## FAQ
1. Confluence Ouput: Direct API push to Confluence or generate Confluence-compatible markdown? Both
2. Tag Fallbacks: If `name` tag is missing should we:
* Use the CloudFormation logical ID?
* Parse the resource ID?
* Mark as unknown?
3. Error Handling: IF one account fails (bad credentials, network issue) should we:
* Continue with other accounts?
* Fail fast?
4. Caching: Should we cache results or always fetch fresh data?
5. App name: If `Name` is `unity-api - ECS Host`. Should we?
* Just use the whole damn string. 
