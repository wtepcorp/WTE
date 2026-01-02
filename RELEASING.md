# Release Guide

## Process Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     RELEASE PROCESS                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   1. Make changes → 2. Update version → 3. Create tag          │
│                                │                                │
│                                ▼                                │
│   4. Push tag → 5. GitHub Actions builds → 6. Release ready    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Semantic Versioning

Use format `vMAJOR.MINOR.PATCH`:

| Change Type | When to Increment | Example |
|-------------|-------------------|---------|
| **MAJOR** | Breaking API changes | v1.0.0 → v2.0.0 |
| **MINOR** | New features (backward compatible) | v1.0.0 → v1.1.0 |
| **PATCH** | Bug fixes | v1.0.0 → v1.0.1 |

**Examples:**
- Added new `wte backup` command → **MINOR** (v1.1.0)
- Fixed bug in `wte install` → **PATCH** (v1.0.1)
- Changed config format (old configs don't work) → **MAJOR** (v2.0.0)

---

## Step-by-Step Release Instructions

### Step 1: Ensure Code is Ready

```bash
# Make sure you're on main branch
git checkout main
git pull origin main

# Run tests
make test

# Verify it builds
make build

# Run linter
make lint
```

### Step 2: Update Version in Code

Edit file `internal/cli/root.go`:

```go
var (
    Version   = "1.1.0"  // ← Update version
    BuildTime = "unknown"
    GitCommit = "unknown"
)
```

### Step 3: Update CHANGELOG (Optional)

If you maintain CHANGELOG.md, add a section:

```markdown
## [1.1.0] - 2024-01-15

### Added
- New `wte backup` command for backups

### Fixed
- Fixed ARM64 installation error

### Changed
- Improved `wte status` output
```

### Step 4: Commit Changes

```bash
git add -A
git commit -m "chore: bump version to v1.1.0"
```

### Step 5: Create Tag

```bash
# Create annotated tag
git tag -a v1.1.0 -m "Release v1.1.0"

# Or with change description
git tag -a v1.1.0 -m "Release v1.1.0

- Added wte backup command
- Fixed ARM64 installation
- Improved status output"
```

### Step 6: Push Commit and Tag

```bash
# Push commit
git push origin main

# Push tag (triggers GitHub Actions)
git push origin v1.1.0

# Or push all tags
git push --tags
```

### Step 7: Verify Release

1. Go to GitHub → Releases
2. Verify workflow started (Actions → Release)
3. Wait for completion (usually 2-3 minutes)
4. Verify binaries are uploaded

---

## Alternative: Using GoReleaser

If you prefer GoReleaser (more powerful tool):

### Install GoReleaser

```bash
# macOS
brew install goreleaser

# Linux
go install github.com/goreleaser/goreleaser@latest

# Or download binary
curl -sfL https://goreleaser.com/static/run | bash
```

### Local Build (No Release)

```bash
# Build without publishing
goreleaser build --snapshot --clean
```

### Release via GoReleaser

```bash
# Create tag
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0

# GoReleaser will run automatically via GitHub Actions
# Or manually:
export GITHUB_TOKEN="your_github_token"
goreleaser release --clean
```

---

## Repository Setup on GitHub

### 1. Enable GitHub Actions

1. GitHub → Settings → Actions → General
2. Select "Allow all actions"
3. In "Workflow permissions" section select "Read and write permissions"

### 2. Create First Release

```bash
git tag -a v1.0.0 -m "Initial release"
git push origin v1.0.0
```

---

## Release Structure

After successful release, GitHub will have:

```
Release v1.1.0
├── wte-linux-amd64.tar.gz      # For standard servers
├── wte-linux-arm64.tar.gz      # For ARM servers (AWS Graviton, etc.)
├── wte-linux-armv7.tar.gz      # For Raspberry Pi
└── checksums.txt                # SHA256 checksums
```

---

## Update Mechanism for Users

### Automatic Update

Users with installed WTE can update with:

```bash
# Check for updates
sudo wte update --check

# Update to latest version
sudo wte update

# Force reinstall current version
sudo wte update --force
```

### Manual Update

```bash
# Download new version
wget https://github.com/wtepcorp/WTE/releases/latest/download/wte-linux-amd64.tar.gz

# Extract
tar -xzf wte-linux-amd64.tar.gz

# Replace binary
sudo mv wte-linux-amd64 /usr/local/bin/wte
sudo chmod +x /usr/local/bin/wte

# Verify version
wte version
```

### Update via Script

Provide users with script:

```bash
curl -sfL https://raw.githubusercontent.com/wtepcorp/WTE/main/install.sh | sudo bash
```

---

## Pre-Release Checklist

- [ ] All tests pass (`make test`)
- [ ] Code compiles (`make build`)
- [ ] Linter passes (`make lint`)
- [ ] Version updated in `internal/cli/root.go`
- [ ] CHANGELOG updated (if maintained)
- [ ] Commit pushed to main
- [ ] Tag created and pushed
- [ ] GitHub Actions completed successfully
- [ ] Binaries available on Releases page

---

## Rolling Back a Release

If something goes wrong:

```bash
# Delete tag locally
git tag -d v1.1.0

# Delete tag on GitHub
git push origin :refs/tags/v1.1.0

# Delete release manually on GitHub
# Settings → Releases → Delete
```

---

## FAQ

**Q: How often to release?**
A: As needed. Accumulate minor fixes, release critical bugs immediately.

**Q: Do I need to test on all platforms?**
A: Preferably test on primary platform (linux/amd64). ARM can be verified in CI.

**Q: What if I forgot to update version in code?**
A: GitHub Actions will substitute version from tag via ldflags. But better to update for consistency.

**Q: Can I release a pre-release?**
A: Yes, use tags like `v1.1.0-beta.1` or `v1.1.0-rc.1`.
