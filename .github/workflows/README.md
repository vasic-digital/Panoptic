# GitHub Actions Manual Execution

This repository's GitHub Actions are configured to run **ONLY ON MANUAL DISPATCH**. Automatic triggers on push/pull requests have been disabled.

## How to Run GitHub Actions Manually

### Method 1: Using GitHub Web Interface

1. Go to the **Actions** tab in your GitHub repository
2. Select **"Panoptic CI/CD Pipeline"** from the workflow list
3. Click the **"Run workflow"** button
4. Configure the run parameters:
   - **Reason**: Optional reason for running CI
   - **Test Type**: Choose what tests to run
     - `all` (default) - Run all tests and build
     - `unit` - Run only unit tests
     - `integration` - Run only integration tests
     - `e2e` - Run only end-to-end tests
     - `security` - Run only security tests
     - `performance` - Run only performance tests
   - **Environment**: Target environment (affects some workflows)
     - `development` (default)
     - `staging`
     - `production`
5. Click **"Run workflow"** to start the CI/CD pipeline

### Method 2: Using GitHub CLI

```bash
# Run all tests with default settings
gh workflow run "Panoptic CI/CD Pipeline"

# Run specific test type
gh workflow run "Panoptic CI/CD Pipeline" --field test_type=unit

# Run with environment
gh workflow run "Panoptic CI/CD Pipeline" --field test_type=integration --field environment=staging

# Run with custom reason
gh workflow run "Panoptic CI/CD Pipeline" --field test_type=security --field reason="Security audit before release"
```

## Workflow Execution Logic

The workflow will execute different jobs based on the selected test type:

| Test Type | Jobs Executed |
|-----------|---------------|
| `all` | All jobs: quality-checks, unit-tests, integration-tests, e2e-tests, security-tests, performance-tests, build, docker, docs |
| `unit` | quality-checks, unit-tests |
| `integration` | quality-checks, unit-tests, integration-tests |
| `e2e` | quality-checks, unit-tests, integration-tests, e2e-tests |
| `security` | quality-checks, unit-tests, security-tests |
| `performance` | quality-checks, unit-tests, integration-tests, performance-tests |

## Environment-Specific Behavior

- **Development**: All features enabled, comprehensive testing
- **Staging**: Production-like testing, some features limited
- **Production**: Security-focused testing, limited diagnostic jobs

## Benefits of Manual Execution

1. **Cost Control**: No unnecessary CI runs on every push
2. **Resource Management**: GitHub Actions minutes conserved
3. **Focused Testing**: Run only what you need when you need it
4. **Flexibility**: Choose test scope and environment
5. **Intentional Testing**: CI runs only when explicitly triggered

## When to Run CI Manually

- **Before Releases**: Run `all` tests with `production` environment
- **After Major Changes**: Run `all` tests to verify everything works
- **Security Audits**: Run `security` tests to check for vulnerabilities
- **Performance Reviews**: Run `performance` tests to measure impact
- **Quick Validation**: Run `unit` or `integration` tests for fast feedback
- **Documentation Updates**: No tests needed, just check documentation generation

## Monitoring and Logs

- Monitor progress in the Actions tab
- Each job provides detailed logs and artifacts
- Failed jobs will upload diagnostic artifacts
- Success/failure notifications can be configured

## Troubleshooting

### Workflow Not Available
- Ensure you have repository write permissions
- Check that the workflow file exists in `.github/workflows/`

### Jobs Skipped
- Verify your test_type selection matches job conditions
- Check environment requirements for specific jobs
- Review job dependencies in the workflow

### Permission Issues
- Ensure GitHub Actions are enabled in repository settings
- Check for any required secrets (DOCKER_PASSWORD, etc.)
- Verify workflow file syntax is correct

## Customization

The workflow can be customized by modifying:
- Test type options in the `workflow_dispatch` inputs
- Job conditions based on your preferences
- Environment-specific behaviors
- Notification and cleanup logic

For questions or issues with manual execution, please create an issue in the repository.