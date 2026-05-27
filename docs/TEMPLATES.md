# ailinter Documentation Templates

Templates for consistent documentation across the ailinter project.

## README Section Template

```markdown
## Section Title

Brief 1-2 line description of what this section covers.

### Subsection

Actionable content: code blocks, tables, or bullet points.

| Column | Description |
|--------|-------------|
| Row    | Value       |

> Callout box for important notes, tips, or warnings.

**Bold** for emphasis, `code` for commands and filenames, [links](url) for references.
```

## Feature Announcement Template

```markdown
### Feature Name (NEW in vX.Y.Z)

Brief description of what was added and why it matters.

**X patterns/features** covering Y categories:

| Category | Count | Examples |
|----------|:---:|----------|
| Category 1 | N | `example1`, `example2` |

Every finding includes [action taken]. Results appear in [output formats].

> **Benchmark:** X result on Y dataset with Z metrics. [Source →](link)
```

## Release Notes Template

```markdown
## vX.Y.Z (YYYY-MM-DD)

### Added
- Feature description (brief)

### Changed
- What changed and why

### Fixed
- Bug description and resolution

### Deprecated
- What's being removed and migration path

### Security
- Security fixes included in this release
```

## GitHub Issue Template

```markdown
### Description
Clear, concise description of the issue.

### Steps to Reproduce
1. Run `command`
2. See error

### Expected Behavior
What should happen.

### Actual Behavior
What happens instead.

### Environment
- OS: [macOS / Linux / Windows]
- ailinter version: [`ailinter --version`]
- Go version: [`go version`]

### Additional Context
Screenshots, logs, or related issues.
```

## PR Description Template

```markdown
### Summary
Brief description of changes.

### Motivation
Why this change is needed.

### Changes
- Itemized list of what was done

### Testing
- [ ] All tests pass (`make test`)
- [ ] Lint passes (`make lint`)
- [ ] Manual verification steps

### Screenshots / Output
Before/after if relevant.

### Related Issues
Closes #X
```

## Blog Post / Article Template

```markdown
# Title: Action-Oriented, Specific, Under 60 Characters

*X min read · YYYY-MM-DD*

## Hook (2-3 sentences)
The problem statement. Why should the reader care?

## Context (2-3 paragraphs)
Background. Current state of things.

## Solution / How We Built It (3-5 paragraphs)
The meat. What you did, why, and how.

## Results / Numbers (1-2 paragraphs + table)
Data. Benchmarks. Outcomes.

## Lessons Learned (2-3 bullet points)
What you'd do differently.

## What's Next (1 paragraph)
Call to action or roadmap preview.

*Links: GitHub · Twitter · Newsletter*
```
