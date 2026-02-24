use serde_json::Value;

#[derive(Clone, serde::Serialize)]
pub struct Fixture {
    pub calls: Vec<Call>,
    pub responses: Responses,
}

#[derive(Clone, serde::Serialize, serde::Deserialize)]
pub struct Call {
    pub func: String,
    pub payload: Value,
}

#[derive(Clone, serde::Serialize, serde::Deserialize)]
pub struct Responses {
    #[serde(rename = "createSessionResponse")]
    pub create_session_response: Response,
    #[serde(rename = "reportResponse")]
    pub report_response: Response,
}

#[derive(Clone, serde::Serialize, serde::Deserialize)]
pub struct Response {
    pub status: String,
    pub payload: Option<Value>,
}

pub fn create_fixtures() -> Vec<(String, Fixture)> {
    vec![
        (
            "createSessionWithId".to_string(),
            Fixture {
                calls: vec![Call {
                    func: "createSession".to_string(),
                    payload: serde_json::json!({
                        "id": "c209c477-d186-49a7-ab83-2ba6dcb409b4",
                        "description": "some description",
                        "baggage": {"a": "b"},
                        "labels": "ab=2,cd"
                    }),
                }],
                responses: Responses {
                    create_session_response: Response {
                        status: "success".to_string(),
                        payload: Some(serde_json::json!({
                            "id": "16af52dc-3296-4249-be93-3aaef3a85845"
                        })),
                    },
                    report_response: Response {
                        status: "success".to_string(),
                        payload: None,
                    },
                },
            },
        ),
        (
            "createSessionWithoutId".to_string(),
            Fixture {
                calls: vec![Call {
                    func: "createSession".to_string(),
                    payload: serde_json::json!({
                        "id": null,
                        "description": null,
                        "baggage": null,
                        "labels": null
                    }),
                }],
                responses: Responses {
                    create_session_response: Response {
                        status: "success".to_string(),
                        payload: Some(serde_json::json!({
                            "id": "16af52dc-3296-4249-be93-3aaef3a85845"
                        })),
                    },
                    report_response: Response {
                        status: "success".to_string(),
                        payload: None,
                    },
                },
            },
        ),
        (
            "createSessionResponseError".to_string(),
            Fixture {
                calls: vec![Call {
                    func: "createSession".to_string(),
                    payload: serde_json::json!({
                        "id": null,
                        "description": null,
                        "baggage": null,
                        "labels": null
                    }),
                }],
                responses: Responses {
                    create_session_response: Response {
                        status: "error".to_string(),
                        payload: Some(serde_json::json!({
                            "code": 3,
                            "ingressCode": 400,
                            "message": "error message"
                        })),
                    },
                    report_response: Response {
                        status: "success".to_string(),
                        payload: None,
                    },
                },
            },
        ),
        (
            "report".to_string(),
            Fixture {
                calls: vec![Call {
                    func: "report".to_string(),
                    payload: serde_json::json!({
                        "testcases": [
                            {
                                "sessionId": "16af52dc-3296-4249-be93-3aaef3a85111",
                                "testcaseName": "test_some_logic",
                                "testcaseClassname": "my_class",
                                "testcaseFile": "my_file.py",
                                "testsuite": "some test suite",
                                "status": "pass",
                                "output": null,
                                "baggage": null
                            }
                        ]
                    }),
                }],
                responses: Responses {
                    create_session_response: Response {
                        status: "success".to_string(),
                        payload: Some(serde_json::json!({
                            "id": "16af52dc-3296-4249-be93-3aaef3a85845"
                        })),
                    },
                    report_response: Response {
                        status: "success".to_string(),
                        payload: None,
                    },
                },
            },
        ),
        (
            "reportNameOnly".to_string(),
            Fixture {
                calls: vec![Call {
                    func: "report".to_string(),
                    payload: serde_json::json!({
                        "testcases": [
                            {
                                "sessionId": "16af52dc-3296-4249-be93-3aaef3a85878",
                                "testcaseName": "test_some_logic",
                                "testcaseClassname": null,
                                "testcaseFile": null,
                                "testsuite": null,
                                "status": "skip",
                                "output": null,
                                "baggage": null
                            }
                        ]
                    }),
                }],
                responses: Responses {
                    create_session_response: Response {
                        status: "success".to_string(),
                        payload: Some(serde_json::json!({
                            "id": "16af52dc-3296-4249-be93-3aaef3a85845"
                        })),
                    },
                    report_response: Response {
                        status: "success".to_string(),
                        payload: None,
                    },
                },
            },
        ),
    ]
}
