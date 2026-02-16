# Contributing to playcheck

Thank you for your interest in contributing to playcheck! We welcome contributions from the community.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally
   ```bash
   git clone https://github.com/yourusername/playcheck.git
   cd playcheck
   ```
3. **Create a branch** for your changes
   ```bash
   git checkout -b feature/my-new-feature
   ```

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git

### Building from Source

```bash
go build -o playcheck ./cmd/scanner
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/manifest
```

## Making Changes

### Code Style

- Follow standard Go formatting (`gofmt`)
- Run `go vet` to catch common errors
- Write clear, descriptive commit messages
- Keep functions small and focused
- Add comments for exported functions and types

### Testing

- **All new features must include tests**
- Aim for 80%+ code coverage
- Write table-driven tests where appropriate
- Include both positive and negative test cases
- Test edge cases

Example test structure:
```go
func TestMyFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "valid input",
            input: "test",
            want:  "expected",
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFeature(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Adding New Rules

When adding a new compliance rule:

1. **Add the rule to `internal/policies/policies.json`**
   ```json
   {
     "id": "NEW001",
     "name": "Rule Name",
     "severity": "CRITICAL",
     "category": "category_name",
     "description": "Detailed description",
     "message": "User-facing message",
     "detection_patterns": [...],
     "remediation": "How to fix",
     "policy_link": "https://..."
   }
   ```

2. **Implement detection logic** in the appropriate scanner:
   - Manifest rules → `internal/manifest/validator.go`
   - Code patterns → `internal/codescan/rules.go`
   - Data safety → `internal/datasafety/checker.go`

3. **Add tests** for the new rule

4. **Update documentation** (README.md and README.ja.md)

### Documentation

- Update README.md for user-facing changes
- Update README.ja.md (Japanese version) as well
- Add godoc comments for exported functions
- Include examples where helpful

## Submitting Changes

1. **Commit your changes**
   ```bash
   git add .
   git commit -m "Add feature: description of change"
   ```

2. **Push to your fork**
   ```bash
   git push origin feature/my-new-feature
   ```

3. **Create a Pull Request** on GitHub
   - Provide a clear title and description
   - Reference any related issues
   - Ensure all tests pass
   - Wait for review

### Pull Request Guidelines

- **Keep PRs focused**: One feature or fix per PR
- **Write clear descriptions**: Explain what and why
- **Include tests**: All code changes should have tests
- **Update docs**: Keep documentation in sync
- **Follow conventions**: Match existing code style
- **Be responsive**: Address review feedback promptly

## Reporting Issues

### Bug Reports

When reporting bugs, please include:

- **Description**: Clear summary of the issue
- **Steps to reproduce**: How to trigger the bug
- **Expected behavior**: What should happen
- **Actual behavior**: What actually happens
- **Environment**: OS, Go version, etc.
- **Sample project**: If possible, a minimal test case

### Feature Requests

For new features, please describe:

- **Use case**: What problem does it solve?
- **Proposed solution**: How should it work?
- **Alternatives**: What other approaches did you consider?
- **Google Play policy**: Link to relevant policy documentation

## Code of Conduct

### Our Standards

- Be respectful and inclusive
- Welcome newcomers
- Focus on constructive feedback
- Accept responsibility for mistakes
- Prioritize community benefit

### Unacceptable Behavior

- Harassment or discriminatory language
- Trolling or insulting comments
- Publishing others' private information
- Other unprofessional conduct

## Questions?

Feel free to:
- Open an issue for discussion
- Ask questions in pull requests
- Reach out to maintainers

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to playcheck! 🎉
