# Coding Guidelines

## Server Structure and Design

- Create a single NewServer constructor function per service that:

    - Takes all dependencies as arguments
    - Returns http.Handler interface type
    - Configures its own muxer
    - Applies all middleware (logging, auth, CORS, etc.)


- Maintain all route definitions in a single routes.go file for better API surface visibility
- Use dependency injection via function arguments instead of embedding in structs
- Keep main() minimal - it should only call run()


## Handler Design

- Write handlers that return http.Handler instead of implementing the interface directly
- Use closure to create handler-specific environments for initialization
- Keep request/response types scoped to handler functions if they're only used there
- Use sync.Once for expensive handler setup operations to improve startup time

## Request Processing

- Create centralized encode/decode functions for consistent request/response handling
- Implement a simple validation interface for request data validation
- Use context for passing request-scoped values


## Middleware

- Use the adapter pattern for middleware: func(http.Handler) http.Handler
- Define middleware in routes.go for clear visibility
- For complex middleware with dependencies, use factory functions that return middleware functions


## Application Lifecycle

- Implement graceful shutdown with context cancellation
- Respect context cancellation at all levels of the application
- Create health check endpoints (/healthz or /readyz) for monitoring readiness
- Use environment variables and flags for configuration, passed through the run() function


## Testing Best Practices

- Write end-to-end tests that exercise the full API flow
- Use the run() function in tests to execute the program as it runs in production
- Implement waiting for readiness in tests using health check endpoints
- Pass system dependencies (stdin, stdout, env vars) as arguments for better testing
- Use context.WithCancel for managing test lifecycle


## Error Handling

- Return errors from all functions that can fail
- Handle errors appropriately at each level
- Use error wrapping for better context


## Code Organization

- Keep global space clear
- Group related functionality together
- Make dependencies explicit through function arguments
- Use interfaces sparingly and keep them simple (preferably single method)

## References
- [How I Write HTTP Services in Go After 13 Years](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/)
