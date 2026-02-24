import asyncio
import json
import aiohttp
from typing import Optional, Dict, Any
from enum import Enum
from collections import deque


class Error(Exception):
    def __init__(self, code: int, ingress_code: int, message: str):
        super().__init__(message)
        self.code = code
        self.ingress_code = ingress_code
        self.message = message

    def __str__(self):
        return f"Error(code={self.code}, ingress_code={self.ingress_code}, message={self.message!r})"


class Session:
    def __init__(self, session_id: str):
        self.id = session_id


class TestcaseStatus(str, Enum):
    __test__ = False

    PASS = "pass"
    FAIL = "fail"
    ERR = "error"
    SKIP = "skip"


class Reporter:
    BATCH_SIZE = 100
    BATCH_TIMEOUT = 5.0

    def __init__(self, endpoint: str, api_key: str):
        self._endpoint = endpoint.rstrip('/')
        self._api_key = api_key
        self._client_session = None
        self._client_session_lock = asyncio.Lock()
        self._testcase_batch = []
        self._errors = deque()
        self._flush_task = asyncio.create_task(self._periodic_flush())
        self._closed = False

    async def _get_client_session(self):
        if self._client_session is None:
            async with self._client_session_lock:
                if self._client_session is None:
                    self._client_session = aiohttp.ClientSession()
        return self._client_session

    async def _make_request(
        self,
        method: str,
        path: str,
        json_data = None
    ) -> Dict[str, Any]:
        client_session = await self._get_client_session()

        url = f"{self._endpoint}{path}"
        headers = {
            "Content-Type": "application/json",
            "X-API-Key": self._api_key
        }

        try:
            async with client_session.request(
                method, url, json=json_data, headers=headers
            ) as resp:
                body = await resp.text()

                if resp.status in (200, 201):
                    if body:
                        return json.loads(body)
                    return {}
                else:
                    try:
                        error_data = json.loads(body)
                        detail = error_data.get("detail", body) if error_data else body
                    except (json.JSONDecodeError, AttributeError):
                        detail = body

                    raise Error(
                        code=-1,
                        ingress_code=resp.status,
                        message=f"Request failed ({resp.status}): {detail}"
                    )
        except aiohttp.ClientError as e:
            raise Error(
                code=-1,
                ingress_code=0,
                message=f"Network error: {str(e)}"
            )

    async def create_session(
        self,
        session_id: Optional[str],
        description: Optional[str],
        baggage: Optional[str],
        labels: Optional[str],
    ) -> Session:
        baggage_data = None
        if baggage is not None:
            try:
                baggage_data = json.loads(baggage)
            except json.JSONDecodeError as e:
                raise Error(
                    code=-1,
                    ingress_code=0,
                    message=f"Invalid baggage JSON: {str(e)}"
                )

        labels_data = None
        if labels is not None:
            parsed_labels = []
            for label_str in labels.split(','):
                label_str = label_str.strip()
                if '=' in label_str:
                    key, value = label_str.split('=', 1)
                    parsed_labels.append({"key": key, "value": value})
                else:
                    parsed_labels.append({"key": label_str})
            labels_data = parsed_labels

        request_data = {
            "id": session_id,
            "description": description,
            "baggage": baggage_data,
            "labels": labels_data,
        }

        response = await self._make_request(
            "POST",
            "/api/v1/ingress/sessions",
            request_data
        )

        return Session(response["id"])

    async def create_testcase(
        self,
        session_id: str,
        testcase_name: str,
        testcase_classname: Optional[str],
        testcase_file: Optional[str],
        testsuite: Optional[str],
        status: TestcaseStatus,
        output: Optional[str],
        baggage: Optional[dict],
    ):
        if self._closed:
            raise Error(
                code=-1,
                ingress_code=0,
                message="Reporter is closed"
            )

        testcase_data = {
            "sessionId": session_id,
            "testcaseName": testcase_name,
            "testcaseClassname": testcase_classname,
            "testcaseFile": testcase_file,
            "testsuite": testsuite,
            "status": status.value,
            "output": output,
            "baggage": baggage,
        }

        self._testcase_batch.append(testcase_data)

        if len(self._testcase_batch) >= self.BATCH_SIZE:
            await self._flush_batch()

    async def _periodic_flush(self):
        try:
            while not self._closed:
                await asyncio.sleep(self.BATCH_TIMEOUT)
                await self._flush_batch()
        except asyncio.CancelledError:
            pass

    async def _flush_batch(self):
        if not self._testcase_batch:
            return

        batch = self._testcase_batch[:]
        self._testcase_batch.clear()

        try:
            await self._make_request(
                "POST",
                "/api/v1/ingress/testcases",
                {"testcases": batch}
            )
        except Error as e:
            self._errors.append(e)

    async def shutdown(self):
        self._closed = True

        if self._flush_task:
            self._flush_task.cancel()
            try:
                await self._flush_task
            except asyncio.CancelledError:
                pass

        await self._flush_batch()

        if self._client_session:
            await self._client_session.close()

    def pop_error(self) -> Optional[Error]:
        if self._errors:
            return self._errors.popleft()
        return None
