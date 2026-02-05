# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Currently supported versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

The Forge team takes security bugs seriously. We appreciate your efforts to responsibly disclose your findings.

### Where to Report

**Please DO NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of these methods:

1. **GitHub Security Advisories** (Preferred)
   - Go to the [Security tab](https://github.com/hamidrabedi/foreit/security)
   - Click "Report a vulnerability"
   - Fill in the details

2. **Email**
   - Send details to: security@foreit.dev
   - Encrypt sensitive information using our PGP key (if available)

### What to Include

Please include the following information in your report:

- Type of vulnerability (e.g., SQL injection, XSS, authentication bypass)
- Full paths of source file(s) related to the vulnerability
- Location of the affected source code (tag/branch/commit or direct URL)
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit it

### Response Timeline

- **Acknowledgment**: Within 48 hours
- **Initial Assessment**: Within 5 business days
- **Status Update**: Every 7 days until resolved
- **Fix Timeline**: Depends on severity
  - Critical: Within 7 days
  - High: Within 14 days
  - Medium: Within 30 days
  - Low: Within 90 days

### Security Update Process

1. **Triage**: We validate and assess the severity
2. **Fix Development**: We develop a fix in a private repository
3. **Testing**: We test the fix thoroughly
4. **Disclosure**: We coordinate disclosure with the reporter
5. **Release**: We release the patched version
6. **Advisory**: We publish a security advisory

### Public Disclosure

- We prefer coordinated disclosure
- We will credit you in the security advisory (unless you prefer to remain anonymous)
- Please give us reasonable time to fix the issue before public disclosure
- We follow a 90-day disclosure timeline unless circumstances require otherwise

## Security Features

### Authentication & Authorization

- JWT-based authentication
- API key authentication
- Basic authentication
- Role-based access control (RBAC)
- Permission system

### Data Protection

- CSRF protection enabled by default
- XSS prevention through proper escaping
- SQL injection prevention via parameterized queries
- Input validation and sanitization
- Output encoding

### Security Headers

The framework sets secure HTTP headers by default:

```
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'
```

### Database Security

- Prepared statements for all queries
- Connection pooling with secure defaults
- Encrypted connections (TLS) support
- No sensitive data in logs

### Session Management

- Secure session cookies (HttpOnly, Secure, SameSite)
- Session timeout
- Session regeneration on authentication
- CSRF token validation

## Security Best Practices

### For Users

1. **Keep Updated**: Always use the latest version
2. **Environment Variables**: Store secrets in environment variables, never in code
3. **HTTPS**: Always use HTTPS in production
4. **Database**: Use strong passwords and restrict database access
5. **Logging**: Don't log sensitive information
6. **Dependencies**: Regularly update dependencies

### For Contributors

1. **Code Review**: All security-sensitive code requires review
2. **Testing**: Write security tests for new features
3. **Dependencies**: Vet third-party dependencies
4. **Secrets**: Never commit secrets or credentials
5. **Documentation**: Document security implications

## Security Scanning

We use multiple security scanning tools:

- **govulncheck**: Go vulnerability scanning
- **gosec**: Go security analyzer
- **npm audit**: NPM dependency scanning
- **Snyk**: Dependency and code scanning
- **Trivy**: Container and filesystem scanning
- **CodeQL**: Semantic code analysis
- **TruffleHog**: Secret scanning
- **OSV Scanner**: Open source vulnerability scanning
- **OpenSSF Scorecard**: Security health metrics

## Compliance

- **OWASP Top 10**: We follow OWASP security guidelines
- **CWE/SANS Top 25**: We address common weakness enumeration
- **OpenSSF Best Practices**: We follow OpenSSF guidelines

## Security Contacts

- **Security Team**: security@foreit.dev
- **Maintainer**: @hamidrabedi

## Bug Bounty

We currently don't have a paid bug bounty program, but we:

- Acknowledge security researchers in our security advisories
- List contributors in our Hall of Fame
- Provide detailed attribution for significant findings

## Security Hall of Fame

We thank the following security researchers for responsibly disclosing vulnerabilities:

<!-- This section will be updated as we receive reports -->

_No vulnerabilities have been reported yet._

## Additional Resources

- [OWASP Go Security Cheatsheet](https://cheatsheetseries.owasp.org/cheatsheets/Go_SG_Cheat_Sheet.html)
- [CWE Top 25](https://cwe.mitre.org/top25/)
- [GitHub Security Best Practices](https://docs.github.com/en/code-security)
- [OpenSSF Best Practices](https://bestpractices.coreinfrastructure.org/)

---

**Last Updated**: 2026-02-05
**Version**: 1.0
