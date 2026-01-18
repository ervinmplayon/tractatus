Tractatus: AWS CLI Tool

## Table of Contents
- [Getting Started](#getting-started)
- [Command Structure](#command-structure)  
- [Arguments](#arguments)  
- [Resource Types](#resource-types-to-query)  
- [Project Structure](#project-structure)

## Getting Started
**Build**
```bash
go mod download
go build -o tractatus ./cmd/main.go
```

**Run**
```bash
./tractatus --account abc-production
```

## Command Structure
```bash
# basic usage
./tractatus --account abc-prod --format table

# with output file
./tractatus --account abc-staging --format markdown --output inventory.md

# multiple accounts
./tractatus --account abc-prod,abc-staging --format table 
```

## Arguments
```bash
+----------------+---------------------------------------------------+-----------------------+--------------------+
|Argument        |           Description                             |       Default         |           Required |
|----------------+---------------------------------------------------+-----------------------+--------------------+
|--account       | AWS account name(s) from config                   |         -             |           Yes      |
|--format        | Output format: table, markdown                    |       table           |           No       |
|--output        | Output destination: stdout, file path, confluence |       stdout          |           No       |
|--use-profile   | (bool) Use AWS credentials                        |       true            |           No       |
+----------------+---------------------------------------------------+-----------------------+--------------------+
```

## Resource Types to Query
```go
resourceTypes := []string{
    "ec2:instance",
    "lambda:function",
    "ecs:service",
    "ecs:cluster",
    "elasticbeanstalk:application",
    "elasticbeanstalk:environment",
    "lightsail:instance",
    // Potentially also:
    "apprunner:service",  // AWS App Runner
    "batch:job-queue",    // AWS Batch
}

# data extratction
## Tags to Extract
| Tag Key | Column Name | Fallback |
|---------|-------------|----------|
| `Name` | App Name | Parse from ARN or "Unknown" |
| `owned-by` | Owner | Check `team` tag, else "Unknown" |
| `team` | Team | Check `owned-by` tag, else "Unknown" |
| `aws:cloudformation:stack-name` | Stack Name | "None" if not CloudFormation-managed |

### Derived Fields

| Column | How to Calculate |
|--------|------------------|
| **Platform** | Parse from ARN: `ec2`, `lambda`, `ecs`, etc. |
| **Has CI/CD** | `Yes` if `aws:cloudformation:stack-name` exists, else `No` |
| **Account** | From which account it was queried |


## Table Output Structure

+------------------+------------+----------+------------------+---------+--------------+
| App Name         | Owner      | Platform | Stack Name       | CI/CD   | Account      |
+------------------+------------+----------+------------------+---------+--------------+
| unity-api        | dashbirds  | ECS      | unity-api-prod   | Yes     | production   |
| data-processor   | data-team  | Lambda   | None             | No      | production   |
| web-frontend     | web-team   | EC2      | web-prod-stack   | Yes     | staging      |
+------------------+------------+----------+------------------+---------+--------------+
```

## Project Structure
```bash
aws-inventory/
├── cmd/
│   └── main.go                 # CLI entry point
├── internal/
│   ├── config/
│   │   └── config.go          # Load config.json
│   ├── aws/
│   │   └── client.go          # AWS API client
│   ├── inventory/
│   │   ├── collector.go       # Resource collection logic
│   │   └── filter.go          # EKS filtering
│   └── output/
│       ├── table.go           # Table formatter
│       └── markdown.go        # Markdown formatter
├── config.json                 # AWS credentials
├── go.mod
└── README.md
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
