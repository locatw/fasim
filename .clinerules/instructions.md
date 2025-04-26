# Factory Automation Simulator (Fasim) Instructions

## Application Overview

Please refer to README.md for the application overview.

## Detailed Design

- Backend API
  - All backend APIs must follow RESTful principles
  - All API requests and responses must use JSON format for data exchange

## Coding

- Source code comments
  - Comments should contain information that cannot be inferred from the code
  - Avoid self-evident comments
  - All comments must be written in English
- Models
  - Do not create setter methods for model fields
  - The `New{Name}FromParams()` function for model object creation should only be used from repository-related code. Use this function only when creating objects from persisted data, and use `New{Name}()` for other purposes.
- Testing
  - All code changes must be verified by running unit tests
  - Run `go test ./...` in the backend directory to execute all tests
  - All tests must pass before considering a task complete
  - When testing multiple scenarios, implement tests as parameterized tests
