from __future__ import annotations

import enum
from dataclasses import dataclass
from datetime import datetime
from typing import Annotated, Any
from uuid import UUID, uuid4

from advanced_alchemy.mixins import AuditColumns
from litestar.dto import DataclassDTO, Mark, dto_field
from litestar.dto.config import DTOConfig
from litestar.pagination import OffsetPagination
from litestar.plugins.sqlalchemy import (
    SQLAlchemyDTO,
    SQLAlchemyDTOConfig,
    base,
    repository,
)
from pydantic import Field
from pydantic.dataclasses import dataclass as pydantic_dataclass
from sqlalchemy import ForeignKey, Integer, MetaData, String, Text
from sqlalchemy.ext.asyncio import AsyncAttrs
from sqlalchemy.orm import Mapped, mapped_column, relationship


class UUIDAuditBase(
    base.CommonTableAttributes, AuditColumns, base.AdvancedDeclarativeBase, AsyncAttrs
):
    __abstract__ = True

    id: Mapped[UUID] = mapped_column(default=uuid4, primary_key=True)


class User(UUIDAuditBase):
    __tablename__ = "users"

    username: Mapped[str] = mapped_column(
        String(128), unique=True, info=dto_field(Mark.PRIVATE)
    )
    password_salt: Mapped[bytes] = mapped_column(info=dto_field(Mark.PRIVATE))
    password_hash: Mapped[bytes] = mapped_column(info=dto_field(Mark.PRIVATE))


class APIKey(UUIDAuditBase):
    __tablename__ = "apikeys"

    description: Mapped[str | None] = mapped_column(Text)
    secret_salt: Mapped[bytes] = mapped_column(info=dto_field(Mark.PRIVATE))
    secret_hash: Mapped[bytes] = mapped_column(info=dto_field(Mark.PRIVATE))

    user_id: Mapped[UUID] = mapped_column(
        ForeignKey("users.id"), info=dto_field(Mark.PRIVATE)
    )
    user: Mapped[User] = relationship(lazy="noload", info=dto_field(Mark.PRIVATE))


@dataclass
class APIKeyUnencrypted:
    id: UUID
    key: str
    description: str | None
    created_at: datetime


class APIKeyWriteDTO(SQLAlchemyDTO[APIKey]):
    config = SQLAlchemyDTOConfig(
        exclude={"id", "created_at", "updated_at"},
        forbid_unknown_fields=True,
        rename_strategy="camel",
    )


class APIKeyReadDTO(SQLAlchemyDTO[APIKey]):
    config = SQLAlchemyDTOConfig(
        exclude={"updated_at"},
        forbid_unknown_fields=True,
        rename_strategy="camel",
    )


class APIKeyUnencryptedReadDTO(DataclassDTO[APIKeyUnencrypted]):
    config = DTOConfig(
        forbid_unknown_fields=True,
        rename_strategy="camel",
    )


class Label(base.BigIntAuditBase):
    __tablename__ = "labels"

    key: Mapped[str] = mapped_column(Text)
    value: Mapped[str | None] = mapped_column(Text)

    user_id: Mapped[UUID] = mapped_column(
        ForeignKey("users.id"), info=dto_field(Mark.PRIVATE)
    )
    user: Mapped[User] = relationship(lazy="noload", info=dto_field(Mark.PRIVATE))
    session_id: Mapped[UUID] = mapped_column(
        ForeignKey("sessions.id", info=dto_field(Mark.PRIVATE))
    )
    session: Mapped[Session_] = relationship(
        lazy="noload", info=dto_field(Mark.PRIVATE)
    )


class TestcaseStatus(enum.IntEnum):
    __test__ = False

    ERROR = 0
    FAIL = 1
    PASS = 2
    SKIP = 3


class Testcase(UUIDAuditBase):
    __tablename__ = "testcases"
    __test__ = False

    status: Mapped[TestcaseStatus] = mapped_column(Integer)
    name: Mapped[str] = mapped_column(Text)
    classname: Mapped[str | None] = mapped_column(Text)
    file: Mapped[str | None] = mapped_column(Text)
    testsuite: Mapped[str | None] = mapped_column(Text)
    output: Mapped[str | None] = mapped_column(Text)
    baggage: Mapped[dict | None]

    user_id: Mapped[UUID] = mapped_column(
        ForeignKey("users.id"), info=dto_field(Mark.PRIVATE)
    )
    user: Mapped[User] = relationship(lazy="noload", info=dto_field(Mark.PRIVATE))
    session_id: Mapped[UUID] = mapped_column(ForeignKey("sessions.id"))
    session: Mapped[Session_] = relationship(
        lazy="noload", info=dto_field(Mark.PRIVATE)
    )


class Session_(UUIDAuditBase):
    __tablename__ = "sessions"

    description: Mapped[str | None] = mapped_column(
        Text, info=dto_field(Mark.READ_ONLY)
    )
    baggage: Mapped[dict | None] = mapped_column(info=dto_field(Mark.READ_ONLY))

    labels: Mapped[list[Label]] = relationship(
        back_populates="session", lazy="noload", info=dto_field(Mark.PRIVATE)
    )
    testcases: Mapped[list[Testcase]] = relationship(
        back_populates="session", lazy="noload", info=dto_field(Mark.PRIVATE)
    )
    user_id: Mapped[UUID] = mapped_column(
        ForeignKey("users.id"), info=dto_field(Mark.PRIVATE)
    )
    user: Mapped[User] = relationship(lazy="noload", info=dto_field(Mark.PRIVATE))


class SessionReadDTO(SQLAlchemyDTO[Session_]):
    config = SQLAlchemyDTOConfig(
        exclude={"updated_at"},
        forbid_unknown_fields=True,
        rename_strategy="camel",
    )


class TestcaseReadDTO(SQLAlchemyDTO[Testcase]):
    config = SQLAlchemyDTOConfig(
        exclude={"updated_at"},
        forbid_unknown_fields=True,
        rename_strategy="camel",
    )


class LabelReadDTO(SQLAlchemyDTO[Label]):
    config = SQLAlchemyDTOConfig(
        exclude={"id", "updated_at"},
        forbid_unknown_fields=True,
        rename_strategy="camel",
    )


@dataclass
class LoginData:
    username: str
    password: str


@dataclass
class RefreshData:
    refresh_token: str


@pydantic_dataclass
class ChangePasswordData:
    password_old: str
    password_new: Annotated[
        str, Field(min_length=6, max_length=32, pattern=r"^[a-zA-Z0-9@_.!\-]*$")
    ]


@dataclass
class TokenData:
    access_token: str
    access_token_expires_at: datetime
    refresh_token: str
    refresh_token_expires_at: datetime


class LoginDataDTO(DataclassDTO[LoginData]):
    config = DTOConfig(
        forbid_unknown_fields=True,
        rename_strategy="camel",
    )


class RefreshDataDTO(DataclassDTO[RefreshData]):
    config = DTOConfig(
        forbid_unknown_fields=True,
        rename_strategy="camel",
    )


class ChangePasswordDataDTO(DataclassDTO[ChangePasswordData]):
    config = DTOConfig(
        forbid_unknown_fields=True,
        rename_strategy="camel",
    )


class TokenDataDTO(DataclassDTO[TokenData]):
    config = DTOConfig(
        forbid_unknown_fields=True,
        rename_strategy="camel",
    )


class StatusRequest(str, enum.Enum):
    PASS = "pass"
    FAIL = "fail"
    ERROR = "error"
    SKIP = "skip"


@dataclass
class TestcaseRequest:
    session_id: str
    testcase_name: str
    status: StatusRequest
    testcase_classname: str | None = None
    testcase_file: str | None = None
    testsuite: str | None = None
    output: str | None = None
    baggage: dict[str, Any] | None = None


@dataclass
class TestcasesRequest:
    testcases: list[TestcaseRequest]


class TestcasesRequestDTO(DataclassDTO[TestcasesRequest]):
    config = DTOConfig(
        forbid_unknown_fields=True,
        rename_strategy="camel",
    )


@dataclass
class TestcaseOffsetPagination(OffsetPagination[Testcase]):
    aggregatedStatus: TestcaseStatus | None


@dataclass
class QueryValidationResult:
    is_grouping: bool


class QueryValidationResultDTO(DataclassDTO[QueryValidationResult]):
    config = DTOConfig(
        forbid_unknown_fields=True,
        rename_strategy="camel",
    )


@dataclass
class GroupItem:
    columns: list[str | None]
    status: TestcaseStatus


@dataclass
class GroupOffsetPagination(OffsetPagination[GroupItem]):
    header: list[str] | None
    aggregated_status: TestcaseStatus | None


class GroupOffsetPaginationDTO(DataclassDTO[GroupOffsetPagination]):
    config = DTOConfig(
        forbid_unknown_fields=True,
        rename_strategy="camel",
    )


################################################################################

METADATA = MetaData()
for model in [
    User,
    APIKey,
    Label,
    Testcase,
    Session_,
]:
    model.__table__.to_metadata(METADATA)


class UserRepository(repository.SQLAlchemyAsyncRepository[User]):
    model_type = User


class APIKeyRepository(repository.SQLAlchemyAsyncRepository[APIKey]):
    model_type = APIKey


class LabelRepository(repository.SQLAlchemyAsyncRepository[Label]):
    model_type = Label


class TestcaseRepository(repository.SQLAlchemyAsyncRepository[Testcase]):
    __test__ = False

    model_type = Testcase


class SessionRepository(repository.SQLAlchemyAsyncRepository[Session_]):
    model_type = Session_
