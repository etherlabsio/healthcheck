<img src="logo.png" width="500" align="top"/>

# Healthcheck 

[![Build Status](https://travis-ci.com/etherlabsio/healthcheck.svg)](https://travis-ci.com/etherlabsio/healthcheck) [![Go Report Card](https://goreportcard.com/badge/github.com/etherlabsio/healthcheck)](https://goreportcard.com/report/github.com/etherlabsio/healthcheck) [![GoDoc](https://godoc.org/github.com/etherlabsio/healthcheck?status.svg)](https://godoc.org/github.com/etherlabsio/healthcheck) [![codecov](https://codecov.io/gh/etherlabsio/healthcheck/branch/master/graph/badge.svg)](https://codecov.io/gh/etherlabsio/healthcheck) [![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fetherlabsio%2Fhealthcheck.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fetherlabsio%2Fhealthcheck?ref=badge_shield)

A simple and extensible RESTful Healthcheck API implementation for Go services.

Health provides an `http.Handlefunc` for use as a healthcheck endpoint used by external services or load balancers. The function is used to determine the health of the application and to remove unhealthy application hosts or containers from rotation.

Instead of blindly returning a `200` HTTP status code, a healthcheck endpoint should test all the mandatory dependencies that are essential for proper functioning of a web service.

Implementing the `Checker` interface and passing it on to healthcheck allows you to test the the dependencies such as a database connection, caches, files and even external services you rely on. You may choose to not fail the healthcheck on failure of certain dependencies such as external services that you are not always dependent on.

## Example

```GO
package main

import (
    "context"
    "database/sql"
    "net/http"
    "time"

    "github.com/etherlabsio/healthcheck"
    "github.com/etherlabsio/healthcheck/checkers"
    _ "github.com/go-sql-driver/mysql"
    "github.com/gorilla/mux"
)

func main() {
    // For brevity, error check is being omitted here.
    db, _ := sql.Open("mysql", "user:password@/dbname")
    defer db.Close()

    r := mux.NewRouter()
    r.Handle("/healthcheck", healthcheck.Handler(

        // WithTimeout allows you to set a max overall timeout.
        healthcheck.WithTimeout(5*time.Second),

        // Checkers fail the status in case of any error.
        healthcheck.WithChecker(
            "heartbeat", checkers.Heartbeat("$PROJECT_PATH/heartbeat"),
        ),

        healthcheck.WithChecker(
            "database", healthcheck.CheckerFunc(
                func(ctx context.Context) error {
                    return db.PingContext(ctx)
                },
            ),
        ),

        // Observers do not fail the status in case of error.
        healthcheck.WithObserver(
            "diskspace", checkers.DiskSpace("/var/log", 90),
        ),
    ))

    http.ListenAndServe(":8080", r)
}
```

Based on the example provided above, `curl localhost:8080/healthcheck | jq` should yield on error a response with an HTTP statusCode of `503`.

```JSON
{
  "status": "Service Unavailable",
  "errors": {
    "database": "dial tcp 127.0.0.1:3306: getsockopt: connection refused",
    "heartbeat": "heartbeat not found. application should be out of rotation"
  }
}
```

## License

This project is licensed under the terms of the MIT license. See the [LICENSE](LICENSE) file.


[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fetherlabsio%2Fhealthcheck.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fetherlabsio%2Fhealthcheck?ref=badge_large)
