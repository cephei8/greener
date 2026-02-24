# Greener Reporter - JUnit XML

Tool to report JUnit XML results to [Greener](https://sr.ht/~cephei8/greener/).

## Usage
```console
$ export GREENER_INGRESS_ENDPOINT=http://localhost:8080
$ export GREENER_INGRESS_API_KEY=<API key created in Greener UI>

$ go install git.sr.ht/~cephei8/reporting/greener-reporter-junitxml@latest

$ greener-reporter-junitxml -f junit-report.xml
```

Check out [Greener documentation](https://man.sr.ht/~cephei8/greener-docs/) or [the main Greener repo](https://git.sr.ht/~cephei8/greener) for details on how to run the Greener server.

For all the reporter configuration options see `greener-reporter-junitxml --help` or [Plugin Configuration](https://man.sr.ht/~cephei8/greener-docs/#plugin-configuration).
