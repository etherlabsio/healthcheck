# Healthcheck
[![CircleCI](https://circleci.com/gh/etherlabsio/healthcheck/tree/master.svg?style=shield)](https://circleci.com/gh/etherlabsio/healthcheck/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/etherlabsio/healthcheck)](https://goreportcard.com/report/github.com/etherlabsio/healthcheck)
[![GoDoc](https://godoc.org/github.com/etherlabsio/healthcheck?status.svg)](https://godoc.org/github.com/etherlabsio/healthcheck)
[![codecov](https://codecov.io/gh/etherlabsio/healthcheck/branch/master/graph/badge.svg)](https://codecov.io/gh/etherlabsio/healthcheck)

A simple and extensible RESTful Healthcheck API implementation for Go services.

Health provides an `http.Handlefunc` for use as a healthcheck endpoint used by external services or load balancers
for determining the health of the application and to remove the application host or container from rotation if it is found to be unhealthy.

Instead of blindly returning a `200` HTTP status code, a healthcheck endpoint should test all the mandatory dependencies that are essential for proper functioning of a web service.

Implementing the `Checker` interface and passing it on to healthcheck allows you to test the the dependencies such as a database connection, caches, files and even external services you rely on. You may choose to not fail the healthcheck on failure of certain dependencies such as external services that you are not always dependent on.

## Example

```GO
    package main

    import (
        "database/sql"
        "net/http"

        "github.com/etherlabsio/healthcheck"
        "github.com/etherlabsio/healthcheck/checkers"
        _ "github.com/go-sql-driver/mysql"
        "github.com/gorilla/mux"
    )

    func main() {
        db, err := sql.Open("mysql", "user:password@/dbname")
        if err != nil {
            panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
        }
        defer db.Close()

        r := mux.NewRouter()
        r.Handle("/healthcheck", healthcheck.Handler(
            healthcheck.WithChecker(
                "heartbeat", checkers.Heartbeat("$PROJECT_PATH/heartbeat"),
            ),
            healthcheck.WithChecker(
                "database", healthcheck.CheckerFunc(func() error {
                    return db.Ping()
                }),
            ),
            healthcheck.WithObserver( // It won't fail the whole status
                "metrics", checkers.Heartbeat("$PROJECT_PATH/metrics"),
            ),
        ))
        http.ListenAndServe(":8080", r)
    }
```

Based on the example provided above, `curl localhost:8080/healthcheck | jq` should yield on error a response with an HTTP statusCode of `503`.

``` JSON
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
