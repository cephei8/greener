use serde::{Deserialize, Serialize};
use serde_json::Value as JsonValue;
use std::fmt;

#[derive(Serialize, Deserialize, Debug, Clone, PartialEq)]
#[serde(rename_all = "lowercase")]
pub enum TestcaseStatus {
    Pass,
    Fail,
    Error,
    Skip,
}

impl fmt::Display for TestcaseStatus {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            TestcaseStatus::Pass => write!(f, "pass"),
            TestcaseStatus::Fail => write!(f, "fail"),
            TestcaseStatus::Error => write!(f, "error"),
            TestcaseStatus::Skip => write!(f, "skip"),
        }
    }
}

#[derive(Serialize, Deserialize, Debug, Clone)]
#[serde(rename_all = "camelCase")]
pub struct Label {
    pub key: String,
    pub value: Option<String>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
#[serde(rename_all = "camelCase")]
pub struct SessionRequest {
    pub id: Option<String>,
    pub description: Option<String>,
    pub baggage: Option<JsonValue>,
    pub labels: Option<Vec<Label>>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
#[serde(rename_all = "camelCase")]
pub struct SessionResponse {
    pub id: String,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
#[serde(rename_all = "camelCase")]
pub struct ErrorResponse {
    pub message: String,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
#[serde(rename_all = "camelCase")]
pub struct TestcaseRequest {
    pub session_id: String,
    pub testcase_name: String,
    pub testcase_classname: Option<String>,
    pub testcase_file: Option<String>,
    pub testsuite: Option<String>,
    pub status: TestcaseStatus,
    pub output: Option<String>,
    pub baggage: Option<JsonValue>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
#[serde(rename_all = "camelCase")]
pub struct TestcasesRequest {
    pub testcases: Vec<TestcaseRequest>,
}
