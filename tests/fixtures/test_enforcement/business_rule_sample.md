# Business Rule: User Authentication

## Rule Title
Users must authenticate before accessing protected resources

## Description
All API endpoints that require authentication must verify the user's identity using a valid JWT token. The token must be validated against the secret key and must not be expired.

## Code Function
`authenticateUser(token string) (User, error)`

## Priority
Critical

## Test Requirements
- Happy Path: Valid token returns user object
- Edge Case: Expired token returns error
- Error Case: Invalid token returns error
- Error Case: Missing token returns error












