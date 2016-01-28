# BitRun API

BitRun API is a project that provides ability to execute code snippets
written in various languages (Ruby, Pythong, Node, PHP, Go, ...). Under the hood
it used Docker to run the code and provides process and filesystem isolation. API
service is written in Go and could be easily installed on the server and only 
requires Docker to run.

## Reference

To execute the code you must provide filename and content. 

```
POST https://bit.run/api/v1/run
```

Parameters:

- `filename` - Name of the file to run. This is needed to determine the language. Required.
- `content` - Code to execute. Required.

Example:

```bash
curl \
  -i \
  -X POST "https://bit.run/api/v1/run" \
  -d "filename=test.rb&content=puts 'Hello World'"
```

Output:

```
HTTP/1.1 200 OK
Content-Type: text/plain
Content-Length: 13
X-Run-Command: ruby test.rb
X-Run-Duration: 261.752379ms
X-Run-Exitcode: 0

Hello World
```

If request is successful, API will respond with plaintext of the executed command.
Extra meta data will be included in the headers:

- `X-Run-Command`  - full command that was executed
- `X-Run-Duration` - how long it took to process the request (not to run the code)
- `X-Run-Exitcode` - exit code of executed command

Each run is limited by 10 seconds. If your code runs longer than 10s API will
respond with 400 and provide error message:

```json
{
  "error": "Operation timed out after 10s"
}
```

### Command override

By default, bitrun will execute code snippet with a default command. For example,
ruby snippet will use the following command: "ruby main.rb". To override the command,
you can specify `command` parameter when making an API call:

```bash
curl \
  -i \
  -X POST "https://bit.run/api/v1/run" \
  -d "filename=test.rb&content=puts 'Hello World'&command=ruby -v"
```

Response:

```
HTTP/1.1 200 OK
Content-Type: text/plain
Content-Length: 59
X-Run-Command: ruby test.rb
X-Run-Duration: 126.436286ms
X-Run-Exitcode: 0

ruby 2.2.3p173 (2015-08-18 revision 51636) [x86_64-linux]
```

### Supported languages

To check which languages are currently supported, make a call:

```
GET https://bit.run/api/v1/config
```

Response will include supported languages along with commands used to run scripts:

```json
{
  ".rb": {
    "image": "ruby:2.2",
    "command": "ruby %s",
    "format": "text/plain"
  },
  ".py": {
    "image": "python:2.7",
    "command": "python %s",
    "format": "text/plain"
  },
  ".php": {
    "image": "php:5.6",
    "command": "php %s",
    "format": "text/plain"
  }
}
```

## License

MIT