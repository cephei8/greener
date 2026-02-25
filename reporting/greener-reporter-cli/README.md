# Greener Reporter - CLI

CLI tool to report test results to [Greener](https://github.com/cephei8/greener/).

## Usage
```console
$ export GREENER_INGRESS_ENDPOINT=http://localhost:8080
$ export GREENER_INGRESS_API_KEY=<API key created in Greener UI>

$ go install github.com/cephei8/greener/reporting/greener-reporter-cli@latest

$ greener-reporter-cli create session --label rc --label version=1.0.0-rc
$ greener-reporter-cli create testcase --session-id=289f7f7b-5e60-434b-bb93-6fa91be513d1 --name s1
```

Check out [Greener repository](https://github.com/cephei8/greener) for details on how to run the Greener server.
