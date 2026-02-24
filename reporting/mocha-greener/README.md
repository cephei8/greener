# Greener Mocha Reporter

[![npm version](https://img.shields.io/npm/v/mocha-greener.svg)](https://www.npmjs.com/package/mocha-greener)

Mocha reporter for [Greener](https://codeberg.org/cephei8/greener/).

## Usage
```console
$ export GREENER_INGRESS_ENDPOINT=http://localhost:8080
$ export GREENER_INGRESS_API_KEY=<API key created in Greener UI>

$ npm install mocha-greener
$ mocha --reporter mocha-greener
```

Check out [Greener documentation](https://codeberg.org/cephei8/greener/src/branch/main/index.md) or [the main Greener repo](https://codeberg.org/cephei8/greener) for details on how to run the Greener server.

For all the reporter configuration options see [Plugin Configuration](https://codeberg.org/cephei8/greener/src/branch/main/index.md#plugin-configuration).

## License
This project is licensed under the terms of the [Apache License 2.0](./LICENSE).
