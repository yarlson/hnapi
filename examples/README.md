# Hacker News API SDK Examples

This directory contains example applications that demonstrate how to use the Hacker News API SDK.

## Running the Examples

To run the main example:

```bash
go run examples/main.go
```

This will demonstrate:
1. Initializing the client with custom options
2. Fetching top stories
3. Getting details for a specific story
4. Retrieving user profiles
5. Batch retrieval of multiple items
6. Real-time updates using the streaming API

The example application will run until you press Ctrl+C to exit. It will gracefully shut down by canceling the context, which stops all ongoing API operations.

## Integration Test

The repository also includes an integration test (`example_test.go` in the root directory) that demonstrates similar functionality in a test context. This can be run with:

```bash
go test -v -run TestIntegration
```

Note: The integration test will be skipped if you run tests in "short" mode (`go test -short`).

## Learn More

For more details on the API methods and configuration options, see the main [README.md](../README.md) file in the repository root. 