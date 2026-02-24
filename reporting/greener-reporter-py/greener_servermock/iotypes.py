from enum import Enum
from typing import Literal, List, Union, Optional, Annotated

from pydantic import BaseModel as PydanticBaseModel, ConfigDict, UUID4, Field
from pydantic.alias_generators import to_camel


class BaseModel(PydanticBaseModel):
    model_config = ConfigDict(extra="forbid", alias_generator=to_camel)


class CallFunc(str, Enum):
    CREATE_SESSION = "createSession"
    REPORT = "report"


class CallCreateSessionPayload(BaseModel):
    id: Optional[UUID4]
    description: Optional[str]
    baggage: Optional[dict]
    labels: Optional[str]


class CallCreateSession(BaseModel):
    func: Literal[CallFunc.CREATE_SESSION]
    payload: CallCreateSessionPayload


class CallReportStatus(str, Enum):
    PASS = "pass"
    FAIL = "fail"
    ERR = "error"
    SKIP = "skip"


class CallReportItem(BaseModel):
    session_id: str
    testcase_name: str
    testcase_classname: Optional[str]
    testcase_file: Optional[str]
    testsuite: Optional[str]
    status: CallReportStatus
    output: Optional[str]
    baggage: Optional[dict]

class CallReportPayload(BaseModel):
    testcases: List[CallReportItem]

class CallReport(BaseModel):
    func: Literal[CallFunc.REPORT]
    payload: CallReportPayload


CallPayload = Annotated[
    Union[CallCreateSession, CallReport], Field(discriminator="func")
]


class Calls(BaseModel):
    calls: List[CallPayload]


class ResponseStatus(str, Enum):
    SUCCESS = "success"
    ERROR = "error"


class ResponseErrorPayload(BaseModel):
    code: int
    ingress_code: int
    message: str


class ResponseError(BaseModel):
    status: Literal[ResponseStatus.ERROR]
    payload: ResponseErrorPayload


class ResponseCreateSessionPayload(BaseModel):
    id: UUID4


class ResponseCreateSession(BaseModel):
    status: Literal[ResponseStatus.SUCCESS]
    payload: ResponseCreateSessionPayload


class ResponseReport(BaseModel):
    status: Literal[ResponseStatus.SUCCESS]
    payload: None


class Responses(BaseModel):
    create_session_response: Union[ResponseCreateSession, ResponseError] = Field(
        discriminator="status"
    )
    report_response: Union[ResponseReport, ResponseError] = Field(
        discriminator="status"
    )
