# What is this?

BitRun API is a project that provides ability to execute code snippets
written in various languages (Ruby, Pythong, Node, PHP, Go, ...). Under the hood
it used Docker to run the code and provides process and filesystem isolation. API
service is written in Go and could be easily installed on the server and only 
requires Docker to run.

## API

To execute the code you must provide filename and content. 

```
POST https://bit.run/run
```

Parameters:

- `filename` - Name of the file to run. This is needed to determine the language. Required.
- `content` - Code to execute. Required.

Example:

```
curl -i -X POST "https://bit.run/run" -d "filename=test.rb&content=puts 'Hello World'"
```

Output:

```
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Content-Length: 13
Connection: keep-alive
X-Duration: 216.185471ms
X-Exit-Code: 0

Hello World
```

If request is successful, API will respond with plaintext of the executed command.
Extra meta data will be included in the headers:

- `X-Duration` - how long it took to process the request (not to run the code)
- `X-Exit-Code` - exit code of executed command

## License

MIT