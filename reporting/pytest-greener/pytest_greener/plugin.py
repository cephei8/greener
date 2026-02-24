import asyncio
import logging
import os
import threading
from typing import Optional

import pytest
from greener_reporter import Reporter, TestcaseStatus

REPORTER_PLUGIN_NAME = "greener_reporter"

logger = logging.getLogger(__name__)


def pytest_addoption(parser):
    group = parser.getgroup("greener", "Greener Reporter options")
    group.addoption(
        "--greener",
        dest="greener",
        action="store_true",
        help="Enable Greener Reporter",
    )


def pytest_configure(config):
    if not config.getoption("greener"):
        return

    reporter = GreenerReporter()
    config.pluginmanager.register(reporter, REPORTER_PLUGIN_NAME)


def pytest_unconfigure(config):
    if not config.pluginmanager.hasplugin(REPORTER_PLUGIN_NAME):
        return

    reporter = config.pluginmanager.getplugin(REPORTER_PLUGIN_NAME)
    reporter.stop()


class ReportOutcome:
    PASSED = "passed"
    FAILED = "failed"
    SKIPPED = "skipped"


class ReportWhen:
    SETUP = "setup"
    CALL = "call"
    TEARDOWN = "teardown"


class GreenerReporter:
    def __init__(self) -> None:
        ingress_endpoint = os.environ.get("GREENER_INGRESS_ENDPOINT")
        if ingress_endpoint is None:
            raise ValueError("GREENER_INGRESS_ENDPOINT is not set")

        ingress_api_key = os.environ.get("GREENER_INGRESS_API_KEY")
        if ingress_api_key is None:
            raise ValueError("GREENER_INGRESS_API_KEY is not set")

        self._loop = asyncio.new_event_loop()
        self._thread = threading.Thread(
            target=self._run_event_loop,
            daemon=True,
            name="greener-reporter-loop"
        )
        self._stopped = False
        self._thread.start()

        try:
            future = asyncio.run_coroutine_threadsafe(
                self._create_reporter(ingress_endpoint, ingress_api_key),
                self._loop
            )
            future.result(timeout=5.0)
        except Exception as e:
            logger.error(f"Failed to initialize greener reporter: {e}")
            self.stop()
            raise

        self._session_id = None
        self._testsuite = None

    def _run_event_loop(self) -> None:
        asyncio.set_event_loop(self._loop)
        self._loop.run_forever()

    async def _create_reporter(self, endpoint: str, api_key: str) -> None:
        self.reporter = Reporter(endpoint, api_key)

    def _handle_testcase_error(self, future: asyncio.Future) -> None:
        try:
            future.result()
        except Exception as e:
            logger.warning(f"Failed to report test case: {e}")

    def stop(self) -> None:
        if self._stopped:
            logger.debug("Reporter already stopped, ignoring")
            return
        self._stopped = True

        if not self._loop:
            return

        try:
            future = asyncio.run_coroutine_threadsafe(
                self.reporter.shutdown(),
                self._loop
            )
            future.result(timeout=10.0)
        except TimeoutError:
            logger.error("Reporter shutdown timed out after 10 seconds")
        except RuntimeError as e:
            logger.debug(f"Event loop already stopped: {e}")
        except Exception as e:
            logger.error(f"Error during reporter shutdown: {e}")

        try:
            self._loop.call_soon_threadsafe(self._loop.stop)
        except RuntimeError:
            logger.debug("Event loop already stopped")

        self._thread.join(timeout=5.0)
        if self._thread.is_alive():
            logger.warning("Reporter thread did not stop cleanly within 5 seconds")

        self._loop = None

    @pytest.hookimpl(tryfirst=True)
    def pytest_configure(self, config):
        self._testsuite = config.getini("junit_suite_name")

    @pytest.hookimpl(wrapper=True)
    def pytest_sessionstart(self, session):
        session_id = os.environ.get("GREENER_SESSION_ID")
        description = os.environ.get("GREENER_SESSION_DESCRIPTION")
        baggage = os.environ.get("GREENER_SESSION_BAGGAGE")
        labels = os.environ.get("GREENER_SESSION_LABELS")

        try:
            future = asyncio.run_coroutine_threadsafe(
                self.reporter.create_session(
                    session_id,
                    description,
                    baggage,
                    labels,
                ),
                self._loop
            )
            greener_session = future.result(timeout=30.0)
            self._session_id = greener_session.id
        except TimeoutError:
            logger.error("Session creation timed out after 30 seconds")
            raise
        except Exception as e:
            logger.error(f"Failed to create greener session: {e}")
            raise

        yield

    @pytest.hookimpl(wrapper=True)
    def pytest_runtest_logreport(self, report):
        status = None
        if report.outcome == ReportOutcome.FAILED and report.when == ReportWhen.SETUP:
            status = TestcaseStatus.ERR
        elif (
                report.outcome == ReportOutcome.SKIPPED and report.when == ReportWhen.SETUP
        ) or report.when == ReportWhen.CALL:
            status = {
                ReportOutcome.PASSED: TestcaseStatus.PASS,
                ReportOutcome.FAILED: TestcaseStatus.FAIL,
                ReportOutcome.SKIPPED: TestcaseStatus.SKIP,
            }[report.outcome]

        if status:
            tc_file, tc_classname, tc_name = _parse_nodeid(report.nodeid)
            try:
                future = asyncio.run_coroutine_threadsafe(
                    self.reporter.create_testcase(
                        self._session_id,
                        tc_name,
                        tc_classname,
                        tc_file,
                        self._testsuite,
                        status,
                        report.longreprtext,
                        None,
                    ),
                    self._loop
                )
                future.add_done_callback(self._handle_testcase_error)
            except RuntimeError as e:
                logger.error(f"Failed to submit test case (event loop stopped): {e}")

        yield


def _parse_nodeid(address: str) -> tuple[str, Optional[str], str]:
    path, possible_open_bracket, params = address.partition("[")
    names = path.split("::")

    tc_file = names[0]
    tc_classname = '.'.join(names[1:-1]) if names[1:-1] else None
    tc_name = names[-1] + possible_open_bracket + params

    return tc_file, tc_classname, tc_name
