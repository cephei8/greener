from litestar import Controller, get
from litestar.status_codes import HTTP_200_OK


class ReadyController(Controller):
    path = "/ready"
    tags = ["Ready"]

    @get("/", status_code=HTTP_200_OK, exclude_from_auth=True)
    async def ready_check(self) -> dict[str, str]:
        return {"status": "ok"}
