from __future__ import annotations

import base64
import hashlib
import json
import secrets
from uuid import UUID

from litestar.connection import ASGIConnection
from litestar.exceptions import NotAuthorizedException
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from app.models import APIKey


async def authenticate_api_key(
    request: ASGIConnection, db_session: AsyncSession
) -> APIKey:
    api_key_header = request.headers.get("x-api-key")
    if not api_key_header:
        raise NotAuthorizedException("Missing X-API-Key header")

    try:
        decoded_data = json.loads(base64.b64decode(api_key_header).decode())
        apikey_id = UUID(decoded_data["apiKeyId"])
        apikey_secret = decoded_data["apiKeySecret"]
    except (ValueError, KeyError, json.JSONDecodeError):
        raise NotAuthorizedException("Invalid API key format")

    stmt = (
        select(APIKey).where(APIKey.id == apikey_id).options(selectinload(APIKey.user))
    )
    result = await db_session.execute(stmt)

    if not (apikey := result.scalars().first()):
        raise NotAuthorizedException("Invalid API key")

    secret_hash = hash_secret(apikey_secret, apikey.secret_salt)
    if not secrets.compare_digest(secret_hash, apikey.secret_hash):
        raise NotAuthorizedException("Invalid API key")

    return apikey


def hash_secret(secret: str, salt: bytes) -> bytes:
    return hashlib.pbkdf2_hmac("sha256", secret.encode(), salt, 100000)
