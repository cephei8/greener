# Greener Jest Reporter

[![npm version](https://img.shields.io/npm/v/jest-greener.svg)](https://www.npmjs.com/package/jest-greener)

Jest reporter for [Greener](https://git.sr.ht/~cephei8/greener/).

## Usage
```console
$ export GREENER_INGRESS_ENDPOINT=http://localhost:8080
$ export GREENER_INGRESS_API_KEY=<API key created in Greener UI>

$ npm install jest-greener
$ jest --reporters="default" --reporters="jest-greener"
```

Check out [Greener repository](https://git.sr.ht/~cephei8/greener) for details on how to run the Greener server.

For all the reporter configuration options see [Plugin Configuration](https://git.sr.ht/~cephei8/greener#plugin-configuration).

## License
This project is licensed under the terms of the [Apache License 2.0](./LICENSE).
