# Contributing to Forge

Thank you for your interest in contributing to Forge! This document provides guidelines and instructions for contributing.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Documentation](#documentation)

## 📜 Code of Conduct

### Our Pledge

We are committed to providing a welcoming and inspiring community for all. Please be respectful and constructive in your interactions.

### Our Standards

- **Be respectful** of differing viewpoints and experiences
- **Give and accept constructive feedback** gracefully
- **Focus on what is best** for the community
- **Show empathy** towards other community members

### Unacceptable Behavior

- Harassment, trolling, or derogatory comments
- Publishing others' private information
- Any conduct that could reasonably be considered inappropriate

## 🚀 Getting Started

### Prerequisites

- **Go 1.25+** - [Install Go](https://go.dev/dl/)
- **PostgreSQL 15+** - [Install PostgreSQL](https://www.postgresql.org/download/)
- **Node.js 22+** - [Install Node](https://nodejs.org/) (for admin UI)
- **Git** - [Install Git](https://git-scm.com/)

### Fork and Clone

```bash
# Fork the repository on GitHub
# Clone your fork
git clone https://github.com/YOUR_USERNAME/foreit.git
cd foreit

# Add upstream remote
git remote add upstream https://github.com/hamidrabedi/foreit.git
```

## 🛠️ Development Setup

### Backend Setup

```bash
cd forge

# Download dependencies
go mod download

# Build all packages
go build ./...

# Run tests
go test ./...
```

### Frontend Setup (Admin UI)

```bash
cd forge/admin/ui/web

# Install dependencies
npm install

# Run development server
npm run dev

# Build for production
npm run build

# Run tests
npm test
```

### Documentation Setup

```bash
cd docs-site

# Install dependencies
npm install

# Run development server
npm start

# Build
npm run build
```

## 🔨 Making Changes

### 1. Create a Branch

```bash
# Update your fork
git checkout main
git pull upstream main

# Create a feature branch
git checkout -b feature/your-feature-name
```

### 2. Make Your Changes

- Write clean, readable code
- Follow the [Coding Standards](#coding-standards)
- Add tests for new features
- Update documentation as needed

### 3. Commit Your Changes

```bash
# Stage your changes
git add .

# Commit with a descriptive message
git commit -m "feat: add awesome feature

- Detailed description of changes
- Why this change is needed
- Any breaking changes"
```

#### Commit Message Format

We use conventional commits:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `perf`: Performance improvements
- `ci`: CI/CD changes

**Examples:**
```
feat(orm): add support for composite primary keys
fix(admin): correct filter query generation
docs(readme): update installation instructions
test(migration): add tests for column renaming
```

## 🔍 Pull Request Process

### Before Submitting

- [ ] Code builds successfully
- [ ] All tests pass
- [ ] New tests added for new features
- [ ] Documentation updated
- [ ] Code follows style guidelines
- [ ] Commit messages are clear
- [ ] No merge conflicts

### Submitting a PR

1. Push your branch to your fork
   ```bash
   git push origin feature/your-feature-name
   ```

2. Open a Pull Request on GitHub

3. Fill in the PR template with:
   - Description of changes
   - Related issues
   - Testing performed
   - Screenshots (if UI changes)

4. Wait for review

### PR Review Process

- **Automated Checks**: CI tests, security scans, linting
- **Code Review**: At least one maintainer review required
- **Security Review**: Automatic security scan on all PRs
- **Testing**: All tests must pass
- **Documentation**: Docs must be updated

### After Approval

- Maintainers will merge your PR
- Your contribution will be acknowledged
- Branch can be deleted

## 💻 Coding Standards

### Go Code

```go
// Good: Clear function names and comments
// CreateUser creates a new user with validation
func CreateUser(ctx context.Context, email string) (*User, error) {
    if email == "" {
        return nil, errors.New("email is required")
    }
    // Implementation...
}

// Bad: Unclear names, no validation
func create(e string) *User {
    return &User{Email: e}
}
```

**Guidelines:**
- Use `gofmt` for formatting
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Write godoc comments for exported functions
- Handle all errors explicitly
- Use context for cancellation
- Avoid global variables
- Keep functions small and focused

### TypeScript/React Code

```typescript
// Good: Type-safe with proper typing
interface User {
  id: number;
  name: string;
  email: string;
}

const UserCard: React.FC<{ user: User }> = ({ user }) => {
  return <div>{user.name}</div>;
};

// Bad: Using any, unclear types
const UserCard = ({ user }: any) => <div>{user.name}</div>;
```

**Guidelines:**
- Use TypeScript strict mode
- Define proper interfaces
- Avoid `any` type
- Use functional components with hooks
- Keep components small
- Use meaningful variable names

### Code Review Checklist

- [ ] Code is clean and readable
- [ ] No commented-out code
- [ ] No debugging statements (console.log, fmt.Println)
- [ ] Error handling is comprehensive
- [ ] Security best practices followed
- [ ] Performance is acceptable
- [ ] No hardcoded values (use config/env)

## 🧪 Testing

### Writing Tests

**Go Tests:**

```go
func TestUserCreation(t *testing.T) {
    // Setup
    ctx := context.Background()
    db := setupTestDB(t)
    defer db.Close()
    
    // Execute
    user, err := CreateUser(ctx, "test@example.com")
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, "test@example.com", user.Email)
}
```

**Frontend Tests:**

```typescript
describe('UserCard', () => {
  it('renders user name', () => {
    const user = { id: 1, name: 'John', email: 'john@example.com' };
    render(<UserCard user={user} />);
    expect(screen.getByText('John')).toBeInTheDocument();
  });
});
```

### Running Tests

```bash
# Run all Go tests
cd forge
go test ./...

# Run integration tests (requires PostgreSQL)
cd ../tests
go test ./integration/...

# Run frontend tests
cd ../forge/admin/ui/web
npm test

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Requirements

- All new features must have tests
- Bug fixes should include regression tests
- Aim for >70% code coverage
- Tests should be fast and isolated
- Use meaningful test names

## 📚 Documentation

### Code Documentation

- Add godoc comments for all exported functions/types
- Include usage examples in comments
- Document edge cases and limitations

```go
// CreateUser creates a new user with the given email.
// It validates the email format and checks for duplicates.
// Returns an error if validation fails or user already exists.
//
// Example:
//   user, err := CreateUser(ctx, "user@example.com")
//   if err != nil {
//       log.Fatal(err)
//   }
func CreateUser(ctx context.Context, email string) (*User, error) {
    // Implementation...
}
```

### User Documentation

- Update relevant documentation in `docs-site/docs/`
- Add examples for new features
- Update API reference if needed
- Include screenshots for UI changes

## 🏗️ Project Structure

```
forge/
├── admin/          # Admin interface
├── api/            # REST API framework
├── cli/            # CLI tools
├── db/             # Database layer
├── orm/            # ORM system
├── schema/         # Schema definition
└── server/         # HTTP server

examples/
└── ecommerce/      # E-commerce example

tests/
├── integration/    # Integration tests
└── testdata/       # Test fixtures

docs-site/
└── docs/           # Documentation
```

## 🐛 Reporting Bugs

### Before Reporting

- Search existing issues
- Try the latest version
- Collect debug information

### Bug Report Template

```markdown
**Describe the bug**
A clear description of the bug.

**To Reproduce**
Steps to reproduce:
1. Go to '...'
2. Click on '...'
3. See error

**Expected behavior**
What you expected to happen.

**Environment:**
- OS: [e.g., Ubuntu 22.04]
- Go version: [e.g., 1.25.0]
- Forge version: [e.g., 1.0.0]

**Additional context**
Any other relevant information.
```

## 💡 Feature Requests

We welcome feature requests! Please:

1. Check if it already exists
2. Describe the use case
3. Explain why it's needed
4. Suggest implementation approach (optional)

## 🔐 Security

**Do not report security vulnerabilities as GitHub issues.**

See [SECURITY.md](SECURITY.md) for security reporting.

## 📞 Getting Help

- **Documentation**: https://hamidrabedi.github.io/foreit/
- **Discussions**: [GitHub Discussions](https://github.com/hamidrabedi/foreit/discussions)
- **Issues**: [GitHub Issues](https://github.com/hamidrabedi/foreit/issues)

## 🎖️ Recognition

Contributors are recognized in:
- GitHub contributors page
- Release notes
- Documentation acknowledgments

## 📄 License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

## Quick Reference

### Common Commands

```bash
# Build
cd forge && go build ./...

# Test
cd forge && go test ./...
cd tests && go test ./integration/...

# Lint
cd forge && golangci-lint run

# Frontend
cd forge/admin/ui/web && npm run dev

# Docs
cd docs-site && npm start
```

### Useful Links

- [Go Documentation](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)
- [React Documentation](https://react.dev/)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)

---

<div align="center">

**Thank you for contributing to Forge! 🙏**

[⭐ Star us on GitHub](https://github.com/hamidrabedi/foreit) • [📖 Documentation](https://hamidrabedi.github.io/foreit/)

</div>
