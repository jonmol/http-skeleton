# Serve

This is where all the setup, start and eventually graceful shutdown is happening. There's a fair bit of code, and it's likely to grow a bit as you add more middleware and databases, but it shouldn't become a lot bigger.

## Config.go

This is one large flag file. All flag names are constants, this is to avoid magic strings and getting compilation errors if a flag is expected but not setup. If you need other types of flags than the ones provided, you'll also need to add them in the two functions addFlags and setDefaults in [serve.go](../serve.go).

If you're adding a flag of an existing type, it's just to add the FieldX constant and inside ConfigStructure. 

## Serve.go

Intimidating but the main entry point is Run(). What it does is:
 - Check if telemetry should be on, and if so start it (separate Go routine)
 - Connect to the database, could be multiple as things grow
 - Start the HTTP listener (separate Go routine)
 - Start listening to HUP signal for restart (separate Go routine)
 - Start a blocking listen for SIGINT and SIGTERM, and if received gracefully shut down and exit.
 
### Functions you're likely to need to edit

Only two functions are likely to need changing here. Of course, if you add new flags you might need to tweak existing code, but the main suspects are:

#### addSecMiddlewares

These are the middlewares added to the secure path, a likely thing to add is a JWT parser so that you can verify a user, and potentially add user information to the handler context being passed around to all handlers and service functions. 

#### addPublicMiddlewares

Currently only has the context middleware, you're likely to want at least a CORS middleware here as well, but it of course depends on your use case.


### Router

The router is done in setupRouter, the actual magic happens in [the router](../../server/router/README.md)
