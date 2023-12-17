# Middleware

Some middleware can be added as third party dependency, but some you want to add yourself. There are two examples here of useful middlewares

## Prometheus

Prometheus does [provide their own middleware](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/promhttp#InstrumentHandlerDuration) but they way they do it is that you have to wrap your handler with their call. Middlewares like that are fine, but if you need five of them you'll have a chain like `mid1(mid2(mid3(mid4(mid5(yourHandlerFunc)))))`, you can of course have a function that does 'AddFiveMiddlewares(handler)' and hides it. It's clearly a case of taste and there's no clear best way of doing it. This middleware has the problem that it has to have access to the routes. On the other hand, it only needs to be called for initialization once, and then it will group them togeteher.

## Context

This is a very small middleware that does three things

It always adds a logger to the request handler. That means that a handler will have a configured logger in the context, which it can also pass on to the service function(s).

If headerName is set, it checks if the request has the header and it's a valid UUID, it uses that. If the header hasn't got the header it creates a UUID, adds it to the logger and to the response headers. All logs printed with the context logger will thus have the traceID present so it's easier to follow a request. A client that reads the header and adds it back for follow-up requests will have the same traceID in the logs. Meaning that a longer session can be connected.

It also has the option to add the request path to the logger, meaning that all logs with the logger will contain the request path making it easier to follow if for instance a service function is called by multiple handler functions. An example of both turned on looks like:

This request:
```
you@puter:~/projects/http-skeleton$ curl -v -X GET http://localhost:3000/v1/myService/private/hello?input=rude
Note: Unnecessary use of -X or --request, GET is already inferred.
*   Trying [::1]:3000...
* Connected to localhost (::1) port 3000
> GET /v1/myService/private/hello?input=rude HTTP/1.1
> Host: localhost:3000
> User-Agent: curl/8.4.0
> Accept: */*
> 
< HTTP/1.1 400 Bad Request
< Access-Control-Allow-Origin: https://example.com
< Content-Type: application/json
< X-Trace: a943bac0-6f72-4c71-8be0-7c505e9194a9
< Date: Sun, 17 Dec 2023 16:45:42 GMT
< Content-Length: 72
< 
* Connection #0 to host localhost left intact
{"error":{"code":"invalid_argument","msg":"no response to rude people"}}

you@puter:~/projects/http-skeleton$ curl  -X GET -H "X-trace: a943bac0-6f72-4c71-8be0-7c505e9194a9" http://localhost:3000/v1/myService/private/hello?input=veryRude
...
< X-Trace: a943bac0-6f72-4c71-8be0-7c505e9194a9
...

```
Gives the following log output:
```
you@puter:~/projects/http-skeleton$ go run main.go serve --mid-trace-id-header X-trace --mid-url-path --log-format json
...
{"time":"2023-12-17T17:39:33.805018654+00:00","level":"ERROR","source":{"function":"github.com/jonmol/http-skeleton/server/handler.(*Handler).Hello.func1","file":"/home/you/projects/http-skeleton/server/handler/api.go","line":29},"msg":"service call failed","serviceUID":"b1b9ce1d-6185-4b57-8134-ca4d7a412ded","pid":25473,"traceID":"a943bac0-6f72-4c71-8be0-7c505e9194a9","requestPath":"/v1/myService/private/hello","pkg":"handler","err":"no response to rude people"}
{"time":"2023-12-17T17:50:26.114639759+00:00","level":"ERROR","source":{"function":"github.com/jonmol/http-skeleton/server/handler.(*Handler).Hello.func1","file":"/home/you/projects/http-skeleton/server/handler/api.go","line":29},"msg":"service call failed","serviceUID":"26fea922-9cb6-4971-9e92-05760036c57c","pid":25659,"traceID":"a943bac0-6f72-4c71-8be0-7c505e9194a9","requestPath":"/v1/myService/private/hello","pkg":"handler","err":"outrageous input"}

```
In the first request a new traceID was created and it was used in the second request. Both have the same UUID in the log so we can infer it was made by the same client. This is obviously not about security, and it's easy for the client to send a valid traceID, this is purely for debugging well behaved clients.
