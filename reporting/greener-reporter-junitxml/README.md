# Greener Reporter - JUnit XML

Tool to report JUnit XML results to [Greener](https://git.sr.ht/~cephei8/greener/).

## Usage
```console
$ export GREENER_INGRESS_ENDPOINT=http://localhost:8080
$ export GREENER_INGRESS_API_KEY=<API key created in Greener UI>

$ go install git.sr.ht/~cephei8/greener/reporting/greener-reporter-junitxml@latest

$ greener-reporter-junitxml -f junit-report.xml
```

Check out [Greener repository](https://git.sr.ht/~cephei8/greener) for details on how to run the Greener server.

For all the reporter configuration options see `greener-reporter-junitxml --help` or [Plugin Configuration](https://git.sr.ht/~cephei8/greener#plugin-configuration).
