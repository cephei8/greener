import json
from ctypes import (
    POINTER,
    byref,
    cdll,
    Structure,
    c_char_p,
    c_uint32,
    c_int,
)
import pathlib
import platform
from typing import List

_PREFIX = {"Darwin": "lib", "Linux": "lib", "Windows": ""}[platform.system()]
_EXT = {"Darwin": ".dylib", "Linux": ".so", "Windows": ".dll"}[platform.system()]
_LIB_NAME = f"{_PREFIX}greener_servermock{_EXT}"
_LIB_PATH = pathlib.Path(__file__).parent.resolve() / _LIB_NAME


class _Servermock(Structure):
    pass


class _Error(Structure):
    _fields_ = [("message", c_char_p)]


class Error(Exception):
    def __init__(self, message: str):
        self.message = message


class Servermock:
    def __init__(self):
        self._lib = cdll.LoadLibrary(str(_LIB_PATH))

        self._lib.greener_servermock_new.restype = POINTER(_Servermock)
        self._lib.greener_servermock_new.argtypes = []

        self._lib.greener_servermock_delete.restype = None
        self._lib.greener_servermock_delete.argtypes = [
            POINTER(_Servermock),
            POINTER(POINTER(_Error)),
        ]

        self._lib.greener_servermock_serve.restype = None
        self._lib.greener_servermock_serve.argtypes = [
            POINTER(_Servermock),
            c_char_p,
            POINTER(POINTER(_Error)),
        ]

        self._lib.greener_servermock_get_port.restype = c_int
        self._lib.greener_servermock_get_port.argtypes = [
            POINTER(_Servermock),
            POINTER(POINTER(_Error)),
        ]

        self._lib.greener_servermock_assert.restype = None
        self._lib.greener_servermock_assert.argtypes = [
            POINTER(_Servermock),
            c_char_p,
            POINTER(POINTER(_Error)),
        ]

        self._lib.greener_servermock_fixture_names.restype = None
        self._lib.greener_servermock_fixture_names.argtypes = [
            POINTER(_Servermock),
            POINTER(POINTER(c_char_p)),
            POINTER(c_uint32),
            POINTER(POINTER(_Error)),
        ]

        self._lib.greener_servermock_fixture_calls.restype = None
        self._lib.greener_servermock_fixture_calls.argtypes = [
            POINTER(_Servermock),
            c_char_p,
            POINTER(c_char_p),
            POINTER(POINTER(_Error)),
        ]

        self._lib.greener_servermock_fixture_responses.restype = None
        self._lib.greener_servermock_fixture_responses.argtypes = [
            POINTER(_Servermock),
            c_char_p,
            POINTER(c_char_p),
            POINTER(POINTER(_Error)),
        ]

        # ----------------------------------------------------------------------

        self._handle = self._lib.greener_servermock_new()

    @staticmethod
    def _verify_call(err) -> None:
        if err:
            raise Error(err.contents.message.decode())

    def serve(self, responses: dict):
        err = POINTER(_Error)()
        self._lib.greener_servermock_serve(
            self._handle,
            c_char_p(json.dumps(responses).encode()),
            byref(err),
        )
        self._verify_call(err)

    @property
    def port(self):
        err = POINTER(_Error)()
        port = self._lib.greener_servermock_get_port(
            self._handle,
            byref(err),
        )
        self._verify_call(err)
        return port

    def assert_calls(self, calls: dict):
        err = POINTER(_Error)()
        self._lib.greener_servermock_assert(
            self._handle,
            c_char_p(json.dumps(calls).encode()),
            byref(err),
        )
        self._verify_call(err)

    def fixture_names(self) -> List[str]:
        fixture_names = POINTER(c_char_p)()
        fixture_names_num = c_uint32()
        err = POINTER(_Error)()
        self._lib.greener_servermock_fixture_names(
            self._handle,
            byref(fixture_names),
            byref(fixture_names_num),
            byref(err),
        )
        self._verify_call(err)

        names = [fixture_names[i].decode() for i in range(fixture_names_num.value)]

        return names

    def fixture_calls(self, name: str) -> dict:
        calls_json_ptr = c_char_p()
        err = POINTER(_Error)()
        self._lib.greener_servermock_fixture_calls(
            self._handle,
            c_char_p(name.encode()),
            byref(calls_json_ptr),
            byref(err),
        )
        self._verify_call(err)

        calls_json: bytes = calls_json_ptr.value  # type: ignore
        return json.loads(calls_json.decode())

    def fixture_responses(self, name: str) -> dict:
        responses_json_ptr = c_char_p()
        err = POINTER(_Error)()
        self._lib.greener_servermock_fixture_responses(
            self._handle,
            c_char_p(name.encode()),
            byref(responses_json_ptr),
            byref(err),
        )
        self._verify_call(err)

        responses_json: bytes = responses_json_ptr.value  # type: ignore
        return json.loads(responses_json.decode())

    def shutdown(self):
        err = POINTER(_Error)()
        self._lib.greener_servermock_delete(
            self._handle,
            byref(err),
        )
        try:
            self._verify_call(err)
        finally:
            self._handle = None
