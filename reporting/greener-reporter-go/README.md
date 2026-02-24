# Greener Reporter - Go

Tool to report Go test results to [Greener](https://sr.ht/~cephei8/greener/).

## Usage
```console
$ export GREENER_INGRESS_ENDPOINT=http://localhost:8080
$ export GREENER_INGRESS_API_KEY=<API key created in Greener UI>

$ go install git.sr.ht/~cephei8/greener/reporting/greener-reporter-go@latest

$ go test -json ./... 2>&1 | greener-reporter-go -f -
```

Check out [Greener documentation](https://man.sr.ht/~cephei8/greener-docs/) or [the main Greener repo](https://git.sr.ht/~cephei8/greener) for details on how to run the Greener server.

For all the reporter configuration options see `greener-reporter-go --help` or [Plugin Configuration](https://man.sr.ht/~cephei8/greener-docs/#plugin-configuration).
