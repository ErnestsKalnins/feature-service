# Setting up the project

If on a *nix system, run `make build` at the project root.
Otherwise, manually do the steps, there aren't many.

Afterwards, setup the database by running `go run cmd/makedb/main.go`. Make
sure to specify the DSN either via environment (`DSN=feature.sqlite go run ...`), or
by configuring a configuration env file and passing it via flag (`go run ... --env-file=dev.env`).
