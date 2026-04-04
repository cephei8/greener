# Greener Pytest Plugin

[![PyPI version](https://img.shields.io/pypi/v/pytest-greener.svg)](https://pypi.org/project/pytest-greener/)

Pytest plugin for [Greener](https://git.sr.ht/~cephei8/greener/).

## Usage
```console
$ export GREENER_INGRESS_ENDPOINT=http://localhost:8080
$ export GREENER_INGRESS_API_KEY=<API key created in Greener UI>

$ pip install pytest-greener
$ pytest --greener
```

Check out [Greener repository](https://git.sr.ht/~cephei8/greener) for details on how to run the Greener server.

For all the plugin configuration options see [Plugin Configuration](https://git.sr.ht/~cephei8/greener#plugin-configuration).

## License
This project is licensed under the terms of the [Apache License 2.0](./LICENSE).
