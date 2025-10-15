# Release Workflow and Versioning Policy

This document outlines the release workflow and versioning policy for the Bazaruto Insurance Platform.

## Versioning Strategy

We follow [Semantic Versioning (SemVer)](https://semver.org/) with the format `MAJOR.MINOR.PATCH`:

- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality in a backwards compatible manner
- **PATCH**: Backwards compatible bug fixes

### Version Examples

- `1.0.0` - Initial stable release
- `1.0.1` - Bug fix release
- `1.1.0` - New feature release
- `2.0.0` - Breaking changes release

## Release Types

### 1. Stable Releases

Stable releases are production-ready versions with:
- Full test coverage
- Documentation updates
- Security patches
- Performance optimizations

### 2. Release Candidates (RC)

Release candidates are pre-release versions for testing:
- `1.1.0-rc.1`
- `1.1.0-rc.2`

### 3. Beta Releases

Beta releases for early testing:
- `1.1.0-beta.1`
- `1.1.0-beta.2`

### 4. Alpha Releases

Alpha releases for internal testing:
- `1.1.0-alpha.1`
- `1.1.0-alpha.2`

## Release Process

### 1. Pre-Release Checklist

Before creating a release, ensure:

- [ ] All tests pass
- [ ] Code is properly formatted and linted
- [ ] Documentation is updated
- [ ] CHANGELOG.md is updated
- [ ] Version is bumped in code
- [ ] Security scan is clean
- [ ] Performance benchmarks are acceptable

### 2. Automated Release Workflow

The release process is automated using GitHub Actions and Goreleaser:

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0
      
      - name: Set up Go
        uses: actions/setup-go@v3
        with:
          go-version: '1.22'
      
      - name: Run tests
        run: make test
      
      - name: Run security scan
        run: |
          go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
          gosec ./...
      
      - name: Run Goreleaser
        uses: goreleaser/goreleaser-action@v3
        with:
          version: latest
          args: release --rm-dist
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### 3. Manual Release Process

#### Step 1: Prepare Release

```bash
# Ensure you're on main branch
git checkout main
git pull origin main

# Run tests
make test

# Update version in code
# Update internal/version/version.go
# Update CHANGELOG.md
# Update README.md if needed

# Commit changes
git add .
git commit -m "chore: prepare release v1.1.0"
git push origin main
```

#### Step 2: Create Release Tag

```bash
# Create and push tag
git tag -a v1.1.0 -m "Release version 1.1.0"
git push origin v1.1.0
```

#### Step 3: Monitor Release

- Check GitHub Actions workflow
- Verify all packages are built
- Test installation packages
- Update release notes if needed

### 4. Post-Release Tasks

- [ ] Update documentation website
- [ ] Announce release on social media
- [ ] Update package repositories
- [ ] Monitor for issues
- [ ] Plan next release

## Release Notes

### Format

Release notes follow this format:

```markdown
# Release v1.1.0

## 🚀 New Features
- Added webhook retry mechanism
- Implemented job queue management
- Added comprehensive CLI commands

## 🐛 Bug Fixes
- Fixed memory leak in job processing
- Resolved database connection issues
- Fixed authentication token refresh

## 🔧 Improvements
- Improved error handling
- Enhanced logging
- Optimized database queries

## 📚 Documentation
- Updated API documentation
- Added deployment guide
- Improved configuration examples

## 🔒 Security
- Updated dependencies
- Fixed security vulnerabilities
- Enhanced authentication

## 📦 Packages
- Linux: bazarutod_1.1.0_amd64.deb, bazarutod-1.1.0-1.x86_64.rpm
- macOS: bazarutod_1.1.0_darwin_amd64.tar.gz
- Windows: bazarutod_1.1.0_windows_amd64.zip
- Docker: bazaruto:1.1.0

## 🏗️ Build Info
- Go version: 1.22.0
- Build date: 2024-01-15T10:30:00Z
- Commit: abc123def456
```

### CHANGELOG.md

Maintain a CHANGELOG.md file following [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- New feature X
- New feature Y

### Changed
- Changed behavior Z

### Fixed
- Fixed bug A
- Fixed bug B

## [1.1.0] - 2024-01-15

### Added
- Webhook retry mechanism with Stripe-inspired logic
- Job queue management CLI commands
- Comprehensive event bus system
- GitHub-style pagination

### Changed
- Updated job system architecture
- Improved error handling across services
- Enhanced logging with Zap integration

### Fixed
- Memory leak in job processing
- Database connection pool issues
- Authentication token refresh logic

## [1.0.0] - 2024-01-01

### Added
- Initial release
- Core insurance platform functionality
- RESTful API
- Authentication and authorization
- Job processing system
- Event-driven architecture
```

## Version Management

### Version Constants

Maintain version information in `internal/version/version.go`:

```go
package version

import (
    "fmt"
    "runtime"
)

var (
    Version   = "dev"
    Commit    = "unknown"
    BuildDate = "unknown"
    GoVersion = runtime.Version()
)

func String() string {
    return fmt.Sprintf("bazarutod %s (commit: %s, built: %s, go: %s)",
        Version, Commit, BuildDate, GoVersion)
}
```

### Build-time Version Injection

Use ldflags to inject version information:

```bash
go build -ldflags "-X main.version=1.1.0 -X main.commit=$(git rev-parse HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" cmd/bazarutod/main.go
```

### Version Command

Implement a version command:

```go
// cmd/version.go
func NewVersionCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "version",
        Short: "Print version information",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println(version.String())
        },
    }
}
```

## Release Branches

### Branch Strategy

- `main` - Production-ready code
- `develop` - Integration branch for features
- `feature/*` - Feature branches
- `hotfix/*` - Hotfix branches
- `release/*` - Release preparation branches

### Release Branch Workflow

```bash
# Create release branch
git checkout -b release/v1.1.0 develop

# Update version and changelog
# ... make changes ...

# Merge to main
git checkout main
git merge --no-ff release/v1.1.0
git tag -a v1.1.0 -m "Release version 1.1.0"

# Merge back to develop
git checkout develop
git merge --no-ff release/v1.1.0

# Delete release branch
git branch -d release/v1.1.0
```

## Hotfix Process

### Emergency Hotfix

```bash
# Create hotfix branch from main
git checkout -b hotfix/v1.0.1 main

# Make hotfix changes
# ... fix critical bug ...

# Update version and changelog
# ... update files ...

# Merge to main
git checkout main
git merge --no-ff hotfix/v1.0.1
git tag -a v1.0.1 -m "Hotfix version 1.0.1"

# Merge to develop
git checkout develop
git merge --no-ff hotfix/v1.0.1

# Delete hotfix branch
git branch -d hotfix/v1.0.1
```

## Release Schedule

### Regular Releases

- **Major releases**: Every 6 months
- **Minor releases**: Every 2 months
- **Patch releases**: As needed for bug fixes
- **Security releases**: Immediately when vulnerabilities are found

### Release Calendar

| Version | Type | Planned Date | Status |
|---------|------|--------------|--------|
| 1.0.0 | Major | 2024-01-01 | ✅ Released |
| 1.0.1 | Patch | 2024-01-15 | ✅ Released |
| 1.1.0 | Minor | 2024-02-01 | ✅ Released |
| 1.1.1 | Patch | 2024-02-15 | 🔄 In Progress |
| 1.2.0 | Minor | 2024-03-01 | 📋 Planned |
| 2.0.0 | Major | 2024-07-01 | 📋 Planned |

## Quality Gates

### Pre-Release Checks

1. **Code Quality**
   - All tests pass (unit, integration, e2e)
   - Code coverage > 80%
   - No linting errors
   - No security vulnerabilities

2. **Performance**
   - Response times within SLA
   - Memory usage within limits
   - Database query performance acceptable

3. **Documentation**
   - API documentation updated
   - User guides updated
   - Configuration examples updated

4. **Compatibility**
   - Backward compatibility maintained
   - Database migrations tested
   - Configuration migration tested

### Release Criteria

- [ ] All quality gates pass
- [ ] Security scan clean
- [ ] Performance benchmarks met
- [ ] Documentation complete
- [ ] Release notes prepared
- [ ] Packages built successfully
- [ ] Installation tested

## Rollback Plan

### Rollback Triggers

- Critical bugs discovered post-release
- Security vulnerabilities found
- Performance degradation
- Data corruption issues

### Rollback Process

1. **Immediate Actions**
   - Disable new deployments
   - Notify stakeholders
   - Assess impact

2. **Rollback Steps**
   - Revert to previous stable version
   - Update load balancer configuration
   - Verify system stability

3. **Post-Rollback**
   - Investigate root cause
   - Fix issues in development
   - Plan new release

### Rollback Communication

```markdown
# Rollback Notice

**Date**: 2024-01-15
**Version**: v1.1.0 → v1.0.1
**Reason**: Critical bug in payment processing
**Impact**: Payment processing temporarily unavailable
**ETA**: 2 hours

## Actions Taken
- Rolled back to v1.0.1
- Disabled payment processing
- Investigating root cause

## Next Steps
- Fix identified bug
- Test thoroughly
- Plan hotfix release

## Contact
- Technical: tech@bazaruto.com
- Business: business@bazaruto.com
```

## Release Metrics

### Track These Metrics

- Release frequency
- Time to release
- Bug discovery rate
- Rollback rate
- User adoption rate
- Performance impact

### Release Dashboard

| Metric | Target | Current | Trend |
|--------|--------|---------|-------|
| Release Frequency | 2 months | 1.8 months | 📈 |
| Time to Release | 2 weeks | 1.5 weeks | 📈 |
| Bug Discovery Rate | < 5% | 3% | 📈 |
| Rollback Rate | < 1% | 0.5% | 📈 |
| User Adoption (7d) | > 80% | 85% | 📈 |

## Communication

### Release Announcements

1. **Internal Communication**
   - Slack/Teams notification
   - Email to stakeholders
   - Documentation updates

2. **External Communication**
   - GitHub release notes
   - Blog post (if major release)
   - Social media announcement
   - Newsletter update

### Stakeholder Notifications

- **Engineering Team**: Technical details, breaking changes
- **Product Team**: New features, user impact
- **Support Team**: Known issues, workarounds
- **Customers**: New features, improvements

## Best Practices

### Do's

- ✅ Follow semantic versioning
- ✅ Maintain detailed changelog
- ✅ Test thoroughly before release
- ✅ Communicate changes clearly
- ✅ Monitor post-release metrics
- ✅ Have rollback plan ready

### Don'ts

- ❌ Skip testing
- ❌ Release without documentation
- ❌ Ignore security vulnerabilities
- ❌ Break backward compatibility without notice
- ❌ Release during business hours (for critical systems)
- ❌ Skip stakeholder communication

## Tools and Automation

### Release Tools

- **Goreleaser**: Automated packaging and releases
- **GitHub Actions**: CI/CD pipeline
- **Semantic Release**: Automated versioning
- **Conventional Commits**: Standardized commit messages

### Monitoring Tools

- **Prometheus**: Release metrics
- **Grafana**: Release dashboards
- **Sentry**: Error tracking
- **DataDog**: Performance monitoring

This release workflow ensures consistent, reliable, and well-documented releases while maintaining high quality and user satisfaction.


