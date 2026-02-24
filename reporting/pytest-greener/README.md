# Greener Pytest Plugin

[![PyPI version](https://img.shields.io/pypi/v/pytest-greener.svg)](https://pypi.org/project/pytest-greener/)

Pytest plugin for [Greener](https://codeberg.org/cephei8/greener/).

## Usage
```console
$ export GREENER_INGRESS_ENDPOINT=http://localhost:8080
$ export GREENER_INGRESS_API_KEY=<API key created in Greener UI>

$ pip install pytest-greener
$ pytest --greener
```

Check out [Greener documentation](https://codeberg.org/cephei8/greener/src/branch/main/index.md) or [the main Greener repo](https://codeberg.org/cephei8/greener) for details on how to run the Greener server.

For all the plugin configuration options see [Plugin Configuration](https://codeberg.org/cephei8/greener/src/branch/main/index.md#plugin-configuration).

## License
This project is licensed under the terms of the [Apache License 2.0](./LICENSE).
