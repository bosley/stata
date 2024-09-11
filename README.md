# stata
Static file hosted with https support for spinning up quick https servers

## build

`go build .`

## use

There isn't much more to it than this:

```
 -bind string
        Address to bind to (default "127.0.0.1")
  -directory string
        Directory to serve (default ".")
  -port int
        Port to serve on (default 8080)
  -secure
        Enable HTTPS with a self-signed certificate
```

./stata --port 80 --secure --directory /tmp/myfiles
