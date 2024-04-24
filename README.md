[![Review Assignment Due Date](https://classroom.github.com/assets/deadline-readme-button-24ddc0f5d75046c5622901739e7c5dd533143b0c8e959d652212380cedb1ea36.svg)](https://classroom.github.com/a/y-k-1497)

# TritonHTTP

## Spec Summary

Here we provide a concise summary of the TritonHTTP spec. You should read the spec doc for more details and clarifications.

### HTTP Messages

TritonHTTP follows the [general HTTP message format](https://developer.mozilla.org/en-US/docs/Web/HTTP/Messages). And it has some further specifications:

- HTTP version supported: `HTTP/1.1`
- Request method supported: `GET`
- Response status supported:
  - `200 OK`
  - `400 Bad Request`
  - `404 Not Found`
- Request headers:
  - `Host` (required)
  - `Connection` (optional, `Connection: close` has special meaning influencing server logic)
  - Other headers are allowed, but won't have any effect on the server logic
- Response headers:
  - `Date` (required)
  - `Last-Modified` (required for a `200` response)
  - `Content-Type` (required for a `200` response)
  - `Content-Length` (required for a `200` response)
  - `Connection: close` (required in response for a `Connection: close` request, or for a `400` response)
  - Response headers should be written in sorted order for the ease of testing
  - Response headers should be returned in 'canonical form', meaning that the first letter and any letter following a hyphen should be upper-case. All other letters in the header string should be lower-case.

### Server Logic

When to send a `200` response?
- When a valid request is received, and the requested file can be found.

When to send a `404` response?
- When a valid request is received, and the requested file cannot be found or is not under the doc root.

When to send a `400` response?
- When an invalid request is received.
- When timeout occurs and a partial request is received.

When to close the connection?
- When timeout occurs and no partial request is received.
- When EOF occurs.
- After sending a `400` response.
- After handling a valid request with a `Connection: close` header.

When to update the timeout?
- When trying to read a new request.

What is the timeout value?
- 5 seconds.

## Implementation

Please limit your implimentation to the following files, because we'll only copy over these files for grading:
- `tritonhttp/`
  - `request.go`
  - `response.go`
  - `server.go`

There are some utility functions defined in `tritonhttp/util.go` that you might find useful.

## Usage

The source code for tools needed to interact with TritonHTTP can be found in `cmd`. The following commands can be used to launch these tools:

1) `make fetch` - A tool that allows you to construct custom responses and send them to your web server. Please refer to the README in `fetch`'s directory for more information.

2) `make gohttpd` - Starts up Go's inbuilt web-server.

3) `make tritonhttpd`  - Starts up your implementation of TritonHTTP

## Submission

Please submit on gradescope through GitHub.
