# Routing

The router package consists of two files, and it sets up all the HTTP routes for the HTTP service

## Routs.go

Here it's the configuration. The endpoints have three different arrays of endpoints:
 - health - for kubernetes or what ever you are using to check the health status
 - private - paths that needs more protection, if it's login protected or rate limiting doesn't matter but the separation exists
 - public - paths that should be globally accessible by anyone
 
A thing to note is the extra `{verb: http.MethodOptions, handler: h.OK}` added to the private and public endpoints. This is so that if a web client calls, the [preflight request](https://developer.mozilla.org/en-US/docs/Glossary/Preflight_request) is returned with OK. Without it CORS will not work.

## Router.go
 
The base path for the private and public endpoints are /v1/serviceName/private resp /public. The reason to have that is to be able to have a lot of different services under the same domain. If you prefer serviceName.example.com it's easy to change so that private and public are direct subrouters from r inside the function BuildRouter. However there are some arguments to keep it as is:

### CORS

If you have a web frontend calling your API you won't have to dabble with CORS in the frontend if you keep them all in the same domain. If you are using app.example.com for the FE and api.example.com that point is moot.

### Versioning

If you need to do breaking changes, you can let your old version serve /v1... for as long as it's needed and spin up the new version on /v2... and have your load balancer route trafic to the new one. This might save you from a lot of headache down the road.

### Service name in path

There's no real difference between having service.example.com/... vs example.com/service/... when it comes to routing. You do however want to have something that groups all your endpoints so that the routing can happen semi-globally rather than on individual endpoints. You might feel at this point that "we will only ever need this service", but chances are you won't and if you go for a general api.example.com/endpoint you'll have to setup routing rules for every endpoint in the load balancer.


### Prometheus middleware

The middleware is a bit annoying, we typically don't want metrics on the health checks since they can be very frequent and squew the stats. The way it's done is that it loops over all endpoints given to it and creates a map of all routes we want to check. If you want to do it for all, it could be initiated like the other middlewares in [serve.go](../../cmd/serve/serve.go) that are passed in, but the way it's used in this project it needs to be told which paths to look for and pass on stats for to Prometheus.
