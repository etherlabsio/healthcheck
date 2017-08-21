# Health

A simple and extensible RESTful Healthcheck API implementation for Go services.

Health provides an `http.Handlefunc` for use as a healthcheck endpoint used bu external services or loadbalancers
for determining the health of the application and to remove the application host or container out of rotation in case it is found to be unhealthy.

Instead of blindly return a `200` HTTP status code, a healthcheck endpoint should test all the mandatory dependencies that are essential for proper functioning of a web service.

By implementing the `Checker` interface and passing it on to healthcheck allows you to test the the dependencies such as a database connection, caches, files and even external services you rely on. You may choose to not fail the healthcheck on failure of certain dependencies such as external services that you are not always dependent on.
