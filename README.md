# Acai Technical Challenge

This technical challenge is part of the interview process for a Software Engineer position at [Acai Travel](https://acaitravel.com). 
If you weren't sent here by one of our engineers, you can [get started here](https://www.acaitravel.com/about/careers).

We know you're eager to get to the code, but please read the instructions carefully before you begin.

The challenge might seem tricky at first, but once you get into it, we hope you'll enjoy the process and have fun 
working with AI and Go.

## Introduction

In this challenge, you'll work on an existing application from this repository, written in [Go](https://go.dev). You can 
make changes, add features, refactor existing code, etc. Think of it as if you've just joined a team and received a task 
to improve an existing codebase.

You will be given a few specific [tasks to complete](#Tasks), but feel free to do some housekeeping if you see something that 
could be improved.

The application is a personal assistant service, which provides an API for conversations with an AI assistant. You could 
say it's an API for an interface similar to ChatGPT: you have an endpoint to start a new conversation, an endpoint to 
send a message to an existing conversation, a way to list conversations, and an endpoint to fetch a conversation by ID.

The assistant is built on top of [OpenAI's model](https://openai.com/), but it leverages 
[additional tools](https://platform.openai.com/docs/guides/function-calling) and potentially some clever prompting to 
provide a more useful experience.

Currently, the assistant can:
- Answer questions about the current date and time.
- Provide weather information (though it seems broken).
- Provide information about holidays in Barcelona.
- Provide general AI assistance.

## About the codebase

We expect you to be able to navigate and figure out the codebase on your own, but here are some key takeaways to give 
you a boost:

- There is a `Makefile` with a few handy commands like `make up` and `make run`.
- The entry point to the application is in `cmd/server/main.go`, but the main logic lives in `internal/chat/server.go`.
- The application stores conversations in a [MongoDB](https://www.mongodb.com/) database. There's a docker compose file 
  to start a local MongoDB instance.
- The application uses [Twirp](https://twitchtv.github.io/twirp/docs/intro.html) and [protobuf](https://protobuf.dev/)
  as a framework for the API. **You do NOT need to dig deep into Twirp and protobuf**. It's easy to use, provides JSON
  via HTTP endpoints, and "automagically" wires HTTP handlers and server implementation.
- The project uses code generation, but you should be able to complete the challenge without needing to run or 
  understand it. In any case, do **not** make manual changes to the `internal/pb` package, maybe consider it a blackbox.

## General guidelines

1. **Do not fork this repository.** Instead, create a new repository in your own GitHub account and copy the contents of 
   this repository into it. Forks are linked to the original repository, and we'd like to avoid candidates discovering 
   each other's solutions. Keep your repository **public** so we can see your solution.
2. **Make use of git history.** It's easier for us to review your code if you commit your changes in meaningful chunks 
   with clear descriptions.
3. **Use standard Go tools.** Use the tools shipped with the Go compiler, such as `go fmt`, `go test`, etc. Avoid 
   unnecessary dependencies or tools. Keep it simple.
4. **Use Go conventions.** Follow Go conventions for naming, formatting, and structuring your code. Check the 
   [Effective Go](https://go.dev/doc/effective_go) and [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments).
5. **Leave comments** where it makes sense. It helps whoever reads the code after you.
6. **You may use AI assistance/co-pilots**, but remember we are looking for a meaningful and maintainable codebase, not 
   something slapped together quickly.

## Setting things up

You'll need:
- [Go](https://go.dev/doc/install) (use whatever version you have, or install the latest).
- [Docker](https://docs.docker.com/get-docker/) (to run the MongoDB container).
- The usual developer tools: git, make, etc.

Set up a repository:
1. Create a new repository in your GitHub account. Clone this repository, then copy everything except the `.git` folder 
   into your own repo.
2. Commit the changes as **"Initial commit"** to set your starting point.

Start the application:
1. Set your OpenAI API key in the environment variable `OPENAI_API_KEY`.
   ```bash
   export OPENAI_API_KEY=your_openai_api_key
   ```
2. Use make to start MongoDB and the application. Make sure docker daemon is running.
   ```bash
   make up run
   ```
3. You should see `Starting the server...`, indicating the HTTP server is running at [localhost:8080](http://localhost:8080).
4. Use `command+C` to stop the server when you're done.
5. Use `make down` to stop the MongoDB container.

## Usage

> Before you interact with the application, make sure it's running, follow steps in the **Setting things up** section.

The application provides a simple HTTP-based API, you can interact with it using any HTTP client (like Postman, curl, 
etc.) or use the [CLI tool](cmd/cli/README.md) provided in this repository.

### CLI tool

You can find [CLI tool](cmd/cli/README.md) in `cmd/cli` to interact with the application.

### HTTP API

We have created a [postman collection](https://documenter.getpostman.com/view/40257649/2sB3BKFo8S) for you to explore 
the API. You can use [postman](https://www.postman.com/) or any other HTTP client.

## Testing

The codebase includes tests for the server and the assistant. The tests require mongoDB to be running, so make sure
to start it with `make up` before running the tests.

Run the tests using:
```bash
go test ./...
```

## Tasks

**You can complete as many tasks as you like**, you can skip tasks that do not appeal to you.
The more tasks you complete, the better we can assess your skills.

We would like you to spend at least 1 hour on the challenge.

# Solution Summary

## Task 1: Fix Conversation Title ✅

**Problem:** Conversation titles were answering questions instead of summarizing topics.

**Solution:** 
- Fixed the `Title()` method in `internal/chat/assistant/assistant.go` to use only the first user message
- Updated prompt to work with O1 model (removed system messages, added clear instructions)
- Titles now properly summarize topics (e.g., "Weather in Barcelona" instead of attempting to answer)

**Bonus:** Implemented parallel execution of title and reply generation in `internal/chat/server.go` using goroutines, reducing response time.

**Note:** Further optimization possible by generating titles asynchronously after response, but current approach ensures title availability in initial response.

---

## Task 2: Fix the Weather ✅

**Problem:** Weather tool returned hardcoded "weather is fine" instead of real data.

**Solution:**
- Integrated WeatherAPI.com for real-time weather data
- Modified `internal/chat/assistant/assistant.go` to fetch and parse actual weather information including forecast for next 3 days as well
- Returns temperature, conditions, wind speed, and humidity

**Environment Variable Required:**
```bash
export WEATHER_API_KEY=your_weatherapi_key
```

---

## Task 3: Refactor Tools ✅

**Problem:** Tools were monolithic, hard to maintain and extend.

**Solution:**
- Created modular tool architecture with interface-based design
- New files: `internal/chat/assistant/tools/` directory containing:
  - `tool.go` - Core tool interface and registry
  - `weather.go` - Weather tool
  - `date.go` - Date/time tool  
  - `holidays.go` - Holidays tool
- Each tool is self-contained and independently testable
- Updated `internal/chat/assistant/assistant.go` to use tool registry

**Bonus:** Added calculator tool for basic mathematical operations.

---

## Task 4: Create Test for StartConversation API ✅

**Problem:** No automated tests for StartConversation endpoint.

**Solution:**
- Added comprehensive tests in `internal/chat/server_test.go`
- Tests verify: conversation creation, title population, reply generation, error handling
- Mock assistant for isolated testing

**Bonus:** Created `internal/chat/assistant/assistant_test.go` with tests for Title() method.

---

## Task 5: Instrument Web Server ✅

**Problem:** No observability or metrics for monitoring application performance.

**Solution:**
- Implemented OpenTelemetry metrics in `internal/httpx/metrics.go`
- Metrics tracked:
  - Request count (by method, path, status)
  - Request duration histogram
  - Active requests gauge
  - Error count by type
- Exposed metrics via `/metrics` endpoint for Prometheus scraping
- Updated `cmd/server/main.go` to initialize metrics

**Bonus:** Added distributed tracing in `internal/httpx/tracing.go` for request flow visibility.

---

## Summary

All 5 tasks completed with bonuses ✅

**Key Improvements:**
- Faster response times with parallel execution of title and reply generation
- Real weather data integration
- Extensible tool architecture
- Comprehensive test coverage
- Full observability with metrics and tracing

**Running the Application:**
```bash
make up    # Start MongoDB
make run   # Start server
make test  # Run tests
```

**Environment Variables:**
```bash
export OPENAI_API_KEY=your_openai_api_key
export WEATHER_API_KEY=your_weatherapi_key
```

