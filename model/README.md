# Databases and external sources

Switching databases is painful, and not something you should have to do often. But if you have 20 different files directly using the special features of your database it's a lot more painful to do that switch than if you have a common function you call (in this case Counter.IncGlobal(...) and Counter.IncWord(...)) which does all the things that are specific to your storage. There's an example of adding a new database at the bottom of this readme.

Having an abstraction like this also allows you to have any type of storage under the hood. You could have a general storage engine, like a cloud buckeet, or plain file under there. Your services don't need to know and don't care.

It can feel a bit tedius writing the code like this since you'll have to update the model package when changing in the service, but in the case of having to switch storage it's a pretty big relief. If you realise you can't bake everything together in a large NoSQL document and need to have relations, or vice versa that your data isn't structured and you end up using large JSON blobs which a NoSQL DB tends to handle better, you're having a very small painful road ahead instead of a long and boring one.


## Looking inside model.go
Model.go isn't really doing much. It's just there as the glue which helps with the abstraction and calling on to the DB selected. In this version the DB struct has a Counter directly set on it, it's set on the two OpenX functions. It works when you only have a couple of tables/distinctions but if it grows, I'd make the private DB.db field exported by renaming it to DB.DB and directly call myDBInstance.DB.Counter.IncGlobal(...). This I'd say is large a matter of taste though.

### The db interface
The interface is there to avoid making special code for a specific database too high up in the hierarchy. It comes with 5 functions:
 - Open - Connect to the DB
 - EnsureDB - Make sure all is ready
 - TearDown - Destroy the DB
 - Close - Disconnect from the DB
 - Healthy - Check our connection / health

#### Open
Simply connect to the database. Some client code needs more arguments, some needs none. Having a context as a parameter is there as a minimum so that if supported you can provide your application level context and use that for a graceful shutdown.

### EnsureDB
This function should be idempotent. It should make sure the tables exists, the indices are there and things are ready to run. It makes local development (and production deployment) a lot smoother, since anyone can just run it and all will be setup. It's also crucial for integration tests that needs to have a database running.

For production code it can be iffy to give the code the right to set indices and create tables. Since the code anyway can destroy all the data if it has delete/update rights I'm not too worried about it. But if you're in a very strict environment it could be good to wrap the call in a feature flag, devmode or similar, making it only being called when you're developing or running tests.

The middleground is to run it in production as well, but that it only checks if the tables/indices are there and if not returns an error and the service fails to start.

### Teardown
This is a destructive and scary function. It should reset the database to the state before EnsureDB is run. It should only be run during integration tests to clean up afterwards so that the next test isn't poluted with data from previous tests.

### Close
Simply disconnect from the database

### Healthy
Simple health check, can be to ping the database, execute a test query, check files exists or whatever makes sense for your database. 

## The DB struct
The flow for creating a new connection is:
 - Call NewModel - returns a pointer to a DB with a logger
 - Call OpenRedis/OpenBadger/OpenYourDatabase with the specific DB parameters. This is the only call where the caller needs to be aware of which database is used
 - Call EnsureDB - makes sure everything is setup properly
If you feel the three calls are too much, EnsureDB can simply be called from OpenX. I do however think the function deserves existing since it gives you a chance to heal a database while running if there's a need. It's a corner case but nice to have covered

## Adding a new database 
Say you realize that a simple key-value store doesn't cover your needs. Instead you need a relational database and you opt for MariaDB.

### Adding DB client
Your first step would be to create a mariadb folder. In there create a mariadb.go file with a DB struct that covers what you need. Create the 5 functions on the struct so that it passes as the db interface in model.go:
```go
import (
	"context"
	"errors"
	"log/slog"

	_ "github.com/go-sql-driver/mysql"
	"database/sql"

    "github.com/jonmol/http-skeleton/model/mariadb/sillycounter"
	"github.com/jonmol/http-skeleton/util/logging"
	"github.com/redis/go-redis/v9"
)

type DB struct {
	db      *sql.DB
	l       *slog.Logger
	dsn    string
	Counter *sillycounter.SillyCounter
}

func (db *DB) Open(ctx context.Context) error {
    sql.Open("mysql", db.dsn)
    ...
}

func (db *DB) Close(ctx context.Context) error {...}
func (db *DB) Healthy(ctx context.Context) bool {...}
func (db *DB) EnsureDB(ctx context.Context) error {...}
func (db *DB) TearDown(ctx context.Context) error {...}

func New(ctx context.Context, dsn string) *DB {...}

```

### Adding silly counter
Then create your sillycounter under mariadb/sillycounter/silly.go:
```go
type SillyCounter struct {
	db *sql.DB
	l  *slog.Logger
}
func (s *SillyCounter) TearDown(ctx context.Context) error {}
func (s *SillyCounter) EnsureDB(_ context.Context) error {}
func (s *SillyCounter) Close(_ context.Context) error {}

func (s *SillyCounter) IncGlobal(ctx context.Context) (uint64, error) {...}

func (s *SillyCounter) IncWord(ctx context.Context, w string) (uint64, error) {
   query := "INSERT INTO word_counter (word, count) values (?, 1) ON CONFLICT (word) DO UPDATE SET count = word_counter.count + 1"
   ...
}

func New(db *sql.DB) *SillyCounter {...}

```

### Add connect function
Now the model and silly counter are in place. Open [model.go](model.go) and add a ConnectMariaDB function:
```go
func (db *DB) OpenMariaDB(ctx context.Context, connStr string) error { ... }

``` 

### Add app level support
Last two steps and it's all done. Edit [serve.go](../cmd/serve/serve.go) to add MariaDB in connectDB:
```go
	case "maria":
		if err := db.OpenMariaDB(ctx, viper.GetString(FieldDBAddr), viper.GetString(FieldDBPass)); err != nil {
			panic(fmt.Sprintf("Failed to connect to %s at %s", viper.GetString(FieldDBType), viper.GetString(FieldDBAddr)))
		} else {
			slog.Info("Connected to MariaDB", slog.String("path", viper.GetString(FieldDBAddr)))
		}

```
And in [config.go](../cmd/serve/config.go) just write that maria is a valid argument:
```go
		{Name: FieldDBType, Desc: "What key value store to use. badger|redis|maria", Def: "badger"},
```

## Summary
Editing five files can feel like a lot, but in general you tend to need to do it once, and once in place it will just work. As your service grows (until it's time to start thinking about splitting it up) you can keep adding new functions.
