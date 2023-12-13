# HTTP Skeleton

A template repo for making stable web services

## Background

When searching the internet for running an HTTP server with Go there are tons of examples out there. However, a vast majority of the examples I've come across, be it Medium articles, Reddit posts or other sources most of them basically comes down to very basic examples or the good old "use the standard libs". 

The standard libs are great, and if you just want to quickly hack something together on your local server, it might also be fine to follow those examples. But there are plenty of pitfalls, and a fair amount of work to make a proper HTTP serving service.

From a pure selfish point of view am I intending to use this project as a template to bootstrap my own projects.

## Caveat

I've been writing this code by frankensteining code I've written in the past with new code into something I feel makes sense. Don't shy away from creating an issue and ask if something doesn't, chances are you're right. I've had no code review on this and it's a bit more than 2000 LOC so there is no guarantee there are no brain farts in there. It does however work as intended, but some design decisions might be awkward.

## Running it

Clone or create a repository with this as the template. Then simply run
 - `go run main.go serve` to use the default settings, starting a server on port 3000 and prometheus on port 9090. 
 - `go run main.go serve -h` to list the available flags
 - `go run main.go serve --cfg-save --config config.yml` to dump the default settings as yaml to config.yml

When running you can test by using your favorite tool to access <http://localhost:3000/v1/myService/private/hello?input=world>

## What does it solve?

A lot of projects starts as a proof of concept, slowly features are added and without really knowing how all of a sudden your test hack is a fundamental part of your platform. Going back and linting, testing, restructuring, instrumenting and more is tedious, boring, risky and takes a lot of time.

While over engineering is always something to be mindful of, having a base skeleton with the usual suspects can be a lot of help when creating a new service. This skeleton is adding a lot of the boilerplate code that is needed to have a stable service. Of course this is highly subjective and based on my opinions.

### Linting

For linting [golangci-lint](https://github.com/golangci/golangci-lint) is used. The configuration is found in [.golangci.yml](.golangci.yml) and all the code is being checked. If you're using CI/CD this should of course be a step in the build to enforce the rules are followed, and [integrating](https://golangci-lint.run/usage/integrations/) with your favorite editor is possible.

### Dependencies

The lowest risk (assuming you have infinite time and knowledge) is always to write all code by yourself, or since you're trusting Go at least sticking to the standard libraries. However, that's rarely the reality and while one should always be very careful with dependencies, even small projects tend to have its fair share of dependencies and you then put your trust in other companies, organizations or individuals. There has been many [cases](https://qz.com/646467/how-one-programmer-broke-the-internet-by-deleting-a-tiny-piece-of-code) where developers goes [rogue](https://developers.slashdot.org/story/22/01/09/2336239/open-source-developer-intentionally-corrupts-his-own-widely-used-libraries) or [companies changing their license](https://www.env0.com/blog/hashicorp-license-change). A good way to protect yourself, apart from of course limiting the amount of external dependencies, is to use the vendor folder. It's annoying, it adds an extra step when updating dependencies and it adds a lot of files to your repository but you will know that your CI pipelines will be using the exact same versions of the dependencies, or when you run `go get` or `go mod tidy` three months later. For this particular project, at the time of writing this, the whole project is 628K, and `go mod vendor` adds 30M to that. It hurts a bit, but in my opinion the good out weights the bad.

### Testing

There are many ways to create tests in Go. I've added two unit tests in [handlers](server/handler/handler_test.go) and [service](server/service/service_test.go) and one integration test in [serve](cmd/serve/serve_test.go) as examples. To run test use the following commands:
 - All tests: `go test ./...`
 - Only unit tests: `go test -short ./...` or `go test -run "Unit" ./...`
 - Only integration tests: `go test -run "integration" ./...`

### Typing

All endpoints (one to be exact since it's a skeleton) use structs for return values. They are located in the [dto](server/dto/) library. It can be very tempting to use maps for input and output to web services, but input validation and output contracts are much harder then. A Go service consumer can simply import that library for the current version and will know that the right data will be sent. I'm using [go-playground/validator](https://github.com/go-playground/validator) to validate the input in [request.go](server/util/request/request.go) and non-complying input will never reach the handler. It's also good for the unit tests which easily will detect if a key is removed and it's also possible to add tests which will fail if the input/output is changed, helping to avoid hard-to-debug consumer errors.

### Help texts, flags and configuration

I'm using [Cobra](https://github.com/spf13/cobra) and [Viper](https://github.com/spf13/viper) to have a somewhat standardized way of handling the binary. This gives a lot "free" functionality for the CLI: structure that many recognize, help texts for flags and commands, configuration as files, flags and env variables and more. I'm not completely in love with how you need to do things with those libraries, but it's a compromise I think is well worth it.

There's also support of saving the current config to disk by using the flags `--cfg-save --config my-file.cfg`, which will save the config based on how you called the service with flags, configuration file and environment variables. 

### Docker

The provided [docker-compose](tools/docker-compose.yml) file contains a couple of usual suspects. You can simply run `docker-compose --project-directory tools up redis-skeleton` and you will be running a local redis you can use for development.

There's also a [Dockerfile](Dockerfile) with three steps in it: build, test and run. Run `docker build .` to build the image and then `docker run -p 3000:3000 <your_image>` to run. You can then make requests to localhost:3000 to test your build.

### Database switching

Switching database tends to be painful and a good way to ease that pain is to put an abastraction layer on top of your storage so that only your abastraction layer is aware of actual database used. There are plenty of ways of doing it and there are pros and coins to them. I've used [Badger](https://dgraph.io/docs/badger/) and [Redis](https://redis.io/) as examples in this repository.

Badger is default, but if you add the flags `--db-type redis --db-addr localhost:6379` the service will use your local Redis server (that you can run with [docker-compose](#docker)) instead. This is more an example of one way of abastracting away the database/kv-store as you'd hopefully not need two key-value stores in one service.

### Routing

Using [gorilla mux](https://github.com/gorilla/mux) for setting up the routes, an old but reliable package with many useful features. When you want to add and endpoint, you add the handler in [server/handler/](server/handler/), change the IHandler interface in[router.go](server/router/router.go) (there for unit tests) and the route in [routs.go](server/router/routs.go). 

### Graceful shutdown

Properly shutting down the HTTP listeners, closing the databases etc rather than just shutting down. This helps avoiding clients getting errors or the database ending up in an unhealthy state.

### K8s monitoring endpoints

Adds /readyz, /healthz and /livez which will let Kubernetes check the state of the service if running under it. The default functions always gives back 200 OK, which should be changed to make proper checks.

### Instrumentation

The service starts a second http listener on port 9090, which can be accessed with http://localhost:9090/metrics, It's set to expose the default Go metrics plus response size, response code and timing of all /private endpoints. To change the behaviour you can use the flags starting with `--mid-prom` and make changes to [prometheus.go](server/middleware/prometheus.go).

## What isn't it

While backend APIs tend to be fairly similiar, CRUD operations based on requests, there tends to be special cases in many of them. This layout doesn't promise to solve all your problems. 

## What's lacking / Project status

Chances are that if you're a seasoned developer, or working at a company with some history you already have your own boilerplate code. With that in mind, this repository is most likely mainly useful for myself and people exploring Go. For an explorer the repository can most likely be a bit overwhelming and more documentation is clearly needed. However, I'm a bit reluctant in spend the time on that before knowing if this is useful to anyone else. If it's only for myself it's not needed, so I'll leave it largely like this and if it gets any traction I will add documentation and answer questions.

## License 

This is all licensensed under GPLV3. The reason for picking this licensense, is if you find it useful but want to make tweaks you have to share those hanges.

## Directory layout
```
├── cmd                     - Command files, cobra style
│   ├── config              - Example if you want to be able to validate your config
│   └── serve               - The heavy lifting done when running 
├── model                   - Base DB struct, example of making it easier to switch between different databases 
│   ├── badger              - Example Badger DB struct
│   │   └── sillycounter
│   └── redis               - Example Redis DB struct
│       ├── common          - Convenience functions for Redis
│       └── sillycounter
├── server                  - HTTP related things, the server setup here
│   ├── ckeys               - My context keys
│   ├── dto                 - Input, output and errors from endpoints
│   ├── handler             - HTTP handlers
│   │   └── mocks           - Mocks for testing
│   ├── middleware          - Example middlewares 
│   ├── router              - HTTP router configuration
│   ├── service             - Services, called by the handlers to do the business logic
│   │   └── mocks           - Mocks for testing
│   └── util                - Common things related to the HTTP server
│       ├── myctx           - Context convenience functions, getting and setting the logger in the request context
│       ├── request         - Parsing and validation of input
│       └── response        - Formatting output to be uniform and error handling
├── instrumentation         - Place for different instrumentation implementations
│   └── otel                - Incomplete example of using OTEL 
├── tools                   - Random tools to aid with the local development env
└── util                    - Common libraries used by the project
    └── logging             - A couple of convenience functions
```
