"""Initial

Revision ID: ebee5d3831a2
Revises:
Create Date: 2025-10-03 23:00:47.441935

"""

import warnings
from typing import TYPE_CHECKING

import sqlalchemy as sa
from advanced_alchemy.types import (
    GUID,
    ORA_JSONB,
    DateTimeUTC,
    EncryptedString,
    EncryptedText,
    PasswordHash,
    StoredObject,
)
from sqlalchemy import Text  # noqa: F401
from sqlalchemy.dialects import postgresql

from alembic import op

if TYPE_CHECKING:
    from collections.abc import Sequence

__all__ = [
    "downgrade",
    "upgrade",
    "schema_upgrades",
    "schema_downgrades",
    "data_upgrades",
    "data_downgrades",
]

sa.GUID = GUID
sa.DateTimeUTC = DateTimeUTC
sa.ORA_JSONB = ORA_JSONB
sa.EncryptedString = EncryptedString
sa.EncryptedText = EncryptedText
sa.StoredObject = StoredObject

revision = "ebee5d3831a2"
down_revision = None
branch_labels = None
depends_on = None


def upgrade() -> None:
    with warnings.catch_warnings():
        warnings.filterwarnings("ignore", category=UserWarning)
        with op.get_context().autocommit_block():
            schema_upgrades()
            data_upgrades()


def downgrade() -> None:
    with warnings.catch_warnings():
        warnings.filterwarnings("ignore", category=UserWarning)
        with op.get_context().autocommit_block():
            data_downgrades()
            schema_downgrades()


def schema_upgrades() -> None:
    """schema upgrade migrations go here."""
    op.create_table(
        "users",
        sa.Column("username", sa.String(), nullable=False),
        sa.Column("password_salt", sa.LargeBinary(), nullable=False),
        sa.Column("password_hash", sa.LargeBinary(), nullable=False),
        sa.Column("id", sa.GUID(length=16), nullable=False),
        sa.Column("sa_orm_sentinel", sa.Integer(), nullable=True),
        sa.Column("created_at", sa.DateTimeUTC(timezone=True), nullable=False),
        sa.Column("updated_at", sa.DateTimeUTC(timezone=True), nullable=False),
        sa.PrimaryKeyConstraint("id", name=op.f("pk_users")),
        sa.UniqueConstraint("username", name=op.f("uq_users_username")),
    )
    op.create_table(
        "apikeys",
        sa.Column("description", sa.String(), nullable=True),
        sa.Column("secret_salt", sa.LargeBinary(), nullable=False),
        sa.Column("secret_hash", sa.LargeBinary(), nullable=False),
        sa.Column("user_id", sa.GUID(length=16), nullable=False),
        sa.Column("id", sa.GUID(length=16), nullable=False),
        sa.Column("sa_orm_sentinel", sa.Integer(), nullable=True),
        sa.Column("created_at", sa.DateTimeUTC(timezone=True), nullable=False),
        sa.Column("updated_at", sa.DateTimeUTC(timezone=True), nullable=False),
        sa.ForeignKeyConstraint(
            ["user_id"], ["users.id"], name=op.f("fk_apikeys_user_id_users")
        ),
        sa.PrimaryKeyConstraint("id", name=op.f("pk_apikeys")),
    )
    op.create_table(
        "sessions",
        sa.Column("description", sa.String(), nullable=True),
        sa.Column(
            "baggage",
            sa.JSON()
            .with_variant(postgresql.JSONB(astext_type=Text()), "cockroachdb")
            .with_variant(sa.ORA_JSONB(), "oracle")
            .with_variant(postgresql.JSONB(astext_type=Text()), "postgresql"),
            nullable=True,
        ),
        sa.Column("user_id", sa.GUID(length=16), nullable=False),
        sa.Column("id", sa.GUID(length=16), nullable=False),
        sa.Column("sa_orm_sentinel", sa.Integer(), nullable=True),
        sa.Column("created_at", sa.DateTimeUTC(timezone=True), nullable=False),
        sa.Column("updated_at", sa.DateTimeUTC(timezone=True), nullable=False),
        sa.ForeignKeyConstraint(
            ["user_id"], ["users.id"], name=op.f("fk_sessions_user_id_users")
        ),
        sa.PrimaryKeyConstraint("id", name=op.f("pk_sessions")),
    )
    op.create_table(
        "labels",
        sa.Column("key", sa.String(), nullable=False),
        sa.Column("value", sa.String(), nullable=True),
        sa.Column("user_id", sa.GUID(length=16), nullable=False),
        sa.Column("session_id", sa.GUID(length=16), nullable=False),
        sa.Column(
            "id", sa.BigInteger().with_variant(sa.Integer(), "sqlite"), nullable=False
        ),
        sa.Column("created_at", sa.DateTimeUTC(timezone=True), nullable=False),
        sa.Column("updated_at", sa.DateTimeUTC(timezone=True), nullable=False),
        sa.ForeignKeyConstraint(
            ["session_id"], ["sessions.id"], name=op.f("fk_labels_session_id_sessions")
        ),
        sa.ForeignKeyConstraint(
            ["user_id"], ["users.id"], name=op.f("fk_labels_user_id_users")
        ),
        sa.PrimaryKeyConstraint("id", name=op.f("pk_labels")),
    )
    op.create_table(
        "testcases",
        sa.Column("status", sa.Integer(), nullable=False),
        sa.Column("name", sa.String(), nullable=False),
        sa.Column("classname", sa.String(), nullable=True),
        sa.Column("file", sa.String(), nullable=True),
        sa.Column("testsuite", sa.String(), nullable=True),
        sa.Column("output", sa.String(), nullable=True),
        sa.Column(
            "baggage",
            sa.JSON()
            .with_variant(postgresql.JSONB(astext_type=Text()), "cockroachdb")
            .with_variant(sa.ORA_JSONB(), "oracle")
            .with_variant(postgresql.JSONB(astext_type=Text()), "postgresql"),
            nullable=True,
        ),
        sa.Column("user_id", sa.GUID(length=16), nullable=False),
        sa.Column("session_id", sa.GUID(length=16), nullable=False),
        sa.Column("id", sa.GUID(length=16), nullable=False),
        sa.Column("sa_orm_sentinel", sa.Integer(), nullable=True),
        sa.Column("created_at", sa.DateTimeUTC(timezone=True), nullable=False),
        sa.Column("updated_at", sa.DateTimeUTC(timezone=True), nullable=False),
        sa.ForeignKeyConstraint(
            ["session_id"],
            ["sessions.id"],
            name=op.f("fk_testcases_session_id_sessions"),
        ),
        sa.ForeignKeyConstraint(
            ["user_id"], ["users.id"], name=op.f("fk_testcases_user_id_users")
        ),
        sa.PrimaryKeyConstraint("id", name=op.f("pk_testcases")),
    )


def schema_downgrades() -> None:
    """schema downgrade migrations go here."""
    op.drop_table("testcases")
    op.drop_table("labels")
    op.drop_table("sessions")
    op.drop_table("apikeys")
    op.drop_table("users")


def data_upgrades() -> None:
    """Add any optional data upgrade migrations here!"""


def data_downgrades() -> None:
    """Add any optional data downgrade migrations here!"""
