# Security Rules

## Secrets Management

- Never commit secrets to git
- Use environment variables
- Store sensitive data in `.env` (gitignored)

## Authentication

- Use secure session management
- Implement proper password hashing (bcrypt/argon2)
- Use HTTPS only

## Input Validation

- Validate on both client and server
- Sanitize all user input
- Use allowlists over denylists

## Dependencies

- Keep dependencies updated
- Review security advisories
- Run `npm audit` regularly

## Logging

- Never log sensitive data (passwords, tokens)
- Use structured logging
- Include request IDs for tracing
