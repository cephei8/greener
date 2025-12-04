from __future__ import annotations

import uuid
from dataclasses import dataclass
from typing import Any

from advanced_alchemy.exceptions import AdvancedAlchemyError, DuplicateKeyError
from litestar import Controller, post
from litestar.connection import ASGIConnection
from litestar.exceptions import ValidationException
from litestar.status_codes import HTTP_201_CREATED, HTTP_204_NO_CONTENT
from sqlalchemy.ext.asyncio import AsyncSession

from app.models import (
    Label,
    LabelRepository,
    Session_,
    SessionRepository,
    StatusRequest,
    Testcase,
    TestcaseRepository,
    TestcasesRequest,
    TestcasesRequestDTO,
    TestcaseStatus,
)
from app.util import authenticate_api_key


@dataclass
class LabelRequest:
    key: str
    value: str | None = None


@dataclass
class SessionRequest:
    id: str | None = None
    description: str | None = None
    baggage: dict[str, Any] | None = None
    labels: list[LabelRequest] | None = None


@dataclass
class SessionResponse:
    id: str


class IngressController(Controller):
    path = "/ingress"
    tags = ["Ingress"]

    def _status_to_enum(self, status: StatusRequest) -> TestcaseStatus:
        status_mapping = {
            StatusRequest.PASS: TestcaseStatus.PASS,
            StatusRequest.FAIL: TestcaseStatus.FAIL,
            StatusRequest.ERROR: TestcaseStatus.ERROR,
            StatusRequest.SKIP: TestcaseStatus.SKIP,
        }
        return status_mapping[status]

    @post("/sessions", status_code=HTTP_201_CREATED)
    async def create_session(
        self,
        data: SessionRequest,
        request: ASGIConnection,
        db_session: AsyncSession,
        session_repo: SessionRepository,
        label_repo: LabelRepository,
    ) -> SessionResponse:
        api_key = await authenticate_api_key(request, db_session)
        user_id = api_key.user_id

        if data.id:
            try:
                session_id = uuid.UUID(data.id)
            except ValueError as e:
                raise ValidationException("Cannot parse session ID") from e
        else:
            session_id = uuid.uuid4()

        session = Session_(
            id=session_id,
            description=data.description,
            baggage=data.baggage,
            user_id=user_id,
        )

        try:
            session = await session_repo.add(session)
        except DuplicateKeyError as e:
            raise ValidationException("Session with this ID already exists") from e

        if data.labels:
            labels = []
            for label_data in data.labels:
                label = Label(
                    session_id=session.id,
                    key=label_data.key,
                    value=label_data.value,
                    user_id=user_id,
                )
                labels.append(label)

            if labels:
                await label_repo.add_many(labels)

        return SessionResponse(id=str(session.id))

    @post("/testcases", dto=TestcasesRequestDTO, status_code=HTTP_201_CREATED)
    async def create_testcases(
        self,
        data: TestcasesRequest,
        request: ASGIConnection,
        db_session: AsyncSession,
        session_repo: SessionRepository,
        testcase_repo: TestcaseRepository,
    ) -> None:
        api_key = await authenticate_api_key(request, db_session)
        user_id = api_key.user_id

        if not data.testcases:
            return

        testcases = []
        for tc_data in data.testcases:
            try:
                session_id = uuid.UUID(tc_data.session_id)
            except ValueError as e:
                raise ValidationException("Cannot parse session ID") from e

            session = await session_repo.get_one_or_none(id=session_id)
            if session is None:
                raise ValidationException("Unknown session ID")

            if session.user_id != user_id:
                raise ValidationException("Session not found")

            testcase = Testcase(
                session_id=session.id,
                name=tc_data.testcase_name,
                classname=tc_data.testcase_classname,
                file=tc_data.testcase_file,
                testsuite=tc_data.testsuite,
                status=self._status_to_enum(tc_data.status),
                output=tc_data.output,
                baggage=tc_data.baggage,
                user_id=user_id,
            )
            testcases.append(testcase)

        await testcase_repo.add_many(testcases)
