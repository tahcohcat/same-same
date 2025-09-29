# Contributing to Same-Same Vector Database

Thank you for your interest in contributing to Same-Same! 🎉 We welcome contributions from developers of all experience levels.

## Quick Start

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/same-same.git
   cd same-same
   ```
3. **Set up your development environment**:
   ```bash
   make dev-setup
   cp .env.example .env
   # Edit .env with your API keys
   ```
4. **Run tests** to ensure everything works:
   ```bash
   make test
   ```

## Development Workflow

### Setting Up
```bash
# Install dependencies
go mod download

# Run locally
make run

# Run with Docker
make docker-run
```

### Making Changes
1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following our coding standards

3. **Add tests** for new functionality

4. **Run the test suite**:
   ```bash
   make test
   make test-coverage
   make lint
   ```

5. **Commit your changes** with a clear commit message:
   ```bash
   git commit -m "feat: add support for custom embedders"
   ```

### Commit Message Format
We follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `test:` Adding tests
- `refactor:` Code refactoring
- `perf:` Performance improvements
- `chore:` Maintenance tasks

## Ways to Contribute

### 🐛 Bug Reports
Found a bug? Please [open an issue](https://github.com/tahcohcat/same-same/issues) with:
- **Clear description** of the problem
- **Steps to reproduce** the issue
- **Expected vs actual behavior**
- **Environment details** (OS, Go version, etc.)
- **Error messages or logs**

### Feature Requests
Have an idea? We'd love to hear it! Please [open an issue](https://github.com/tahcohcat/same-same/issues) with:
- **Problem description** - What problem does this solve?
- **Proposed solution** - How should it work?
- **Alternatives considered** - What other approaches did you think about?
- **Use case examples** - When would this be used?

### Documentation
- Fix typos or unclear explanations
- Add examples for common use cases
- Improve API documentation
- Create tutorials or guides

### Code Contributions

#### Priority Areas
- **New Embedders**: Add support for other embedding providers
- **Storage Backends**: Implement persistent storage options
- **Performance**: Optimize vector search algorithms
- **Monitoring**: Add metrics and observability
- **Security**: Improve authentication and authorization
- **Testing**: Increase test coverage

#### Architecture Guidelines
- **Embedder Interface**: All embedders must implement `embedders.Embedder`
- **Storage Interface**: Follow the storage contract in `internal/storage/`
- **Handler Pattern**: HTTP handlers should be thin and delegate to services
- **Error Handling**: Use structured errors with appropriate HTTP status codes
- **Logging**: Use structured logging with appropriate levels

## Code Standards

### Go Style Guide
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Run `golangci-lint` for linting
- Add godoc comments for public functions

### Testing
- **Unit tests** for all business logic
- **Integration tests** for API endpoints
- **Table-driven tests** for multiple test cases
- **Mocks** for external dependencies
- Target **80%+ code coverage**

### Example Test
```go
func TestVectorHandler_CreateVector(t *testing.T) {
    tests := []struct {
        name           string
        input          models.Vector
        expectedStatus int
        expectedError  string
    }{
        {
            name: "valid vector",
            input: models.Vector{
                ID: "test1",
                Embedding: []float64{0.1, 0.2, 0.3},
            },
            expectedStatus: http.StatusCreated,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Adding New Embedders

1. **Create embedder directory**:
   ```
   internal/embedders/quotes/your_provider/
   ```

2. **Implement the interface**:
   ```go
   type YourEmbedder struct {
       apiKey string
       // ... other fields
   }

   func (e *YourEmbedder) Embed(text string) ([]float64, error) {
       // Implementation
   }
   ```

3. **Add configuration** in `internal/server/server.go`

4. **Write tests** with mocked HTTP responses

5. **Update documentation** with setup instructions

## Testing Your Changes

### Local Testing
```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Test specific package
go test ./internal/handlers -v

# Integration tests
make api-test
```

### Docker Testing
```bash
# Build and test in Docker
make docker-build
make docker-run
make health-check
```

## Documentation

### API Documentation
- Update OpenAPI specs in `docs/api.yaml`
- Include request/response examples
- Document error codes and messages

### Code Documentation
- Add godoc comments for public APIs
- Include usage examples in comments
- Document complex algorithms or business logic

## Review Process

1. **Submit PR** with clear title and description
2. **Automated checks** must pass (tests, linting, etc.)
3. **Code review** by maintainers
4. **Address feedback** if requested
5. **Merge** once approved

### PR Checklist
- [ ] Tests pass locally
- [ ] New tests added for new functionality
- [ ] Documentation updated
- [ ] Commit messages follow convention
- [ ] No breaking changes (or clearly documented)
- [ ] Performance impact considered

## 🎖️ Recognition

Contributors will be:
- Listed in `CONTRIBUTORS.md`
- Mentioned in release notes
- Thanked in project documentation

## Questions?

- **General questions**: [Open a discussion](https://github.com/tahcohcat/same-same/discussions)
- **Bug reports**: [Open an issue](https://github.com/tahcohcat/same-same/issues)
- **Security issues**: Email security@same-same.dev

## Code of Conduct

We follow the [Contributor Covenant](https://www.contributor-covenant.org/) to ensure a welcoming environment for all contributors.

---

**Happy Contributing!** 

Your contributions help make Same-Same better for everyone. Thank you for being part of our community!