use greener_reporter::{
    Label, Reporter, ReporterError, SessionRequest, TestcaseRequest, TestcaseStatus,
    GREENER_REPORTER_ERROR, GREENER_REPORTER_ERROR_INGRESS,
    GREENER_REPORTER_ERROR_INVALID_ARGUMENT,
};
use greener_servermock::GreenerServermock;
use serde_json::Value;

fn get_fixture_names() -> Vec<String> {
    let mut servermock = GreenerServermock::new();
    let names = servermock.fixture_names();
    assert!(!names.is_empty(), "fixture names are empty");
    names
}

fn process_fixture(fixture_name: &str) {
    let mut servermock = GreenerServermock::new();
    let calls = servermock
        .fixture_calls(fixture_name)
        .expect("failed to get fixture calls")
        .to_string();
    let responses = servermock
        .fixture_responses(fixture_name)
        .expect("failed to get fixture responses")
        .to_string();
    servermock
        .serve(&responses)
        .expect("failed to serve responses");
    let port = servermock.port();

    let endpoint = format!("http://127.0.0.1:{}", port);
    let api_key = "some-api-token".to_string();

    let reporter = Reporter::new(endpoint, api_key).expect("failed to create reporter");
    let calls_json: serde_json::Value =
        serde_json::from_str(&calls).expect("failed to parse calls JSON");
    let responses_json: serde_json::Value =
        serde_json::from_str(&responses).expect("failed to parse responses JSON");

    for call in calls_json["calls"]
        .as_array()
        .expect("calls is not an array")
    {
        make_call(&reporter, call, &responses_json);
    }

    reporter.shutdown().expect("failed to shutdown reporter");
    servermock
        .assert(&calls)
        .expect("calls did not match expected pattern");
}

fn make_call(reporter: &Reporter, call: &Value, responses: &Value) {
    let c_func = match call["func"].as_str() {
        Some(f) => f,
        None => {
            eprintln!("call value: {:#?}", call);
            panic!("call.func is not a string");
        }
    };
    let c_payload = &call["payload"];

    match c_func {
        "createSession" => {
            let r = &responses["createSessionResponse"];
            let r_status = r["status"].as_str().unwrap_or("");

            let session_id = if !c_payload["id"].is_null() {
                c_payload["id"].as_str().map(|s| s.to_string())
            } else {
                None
            };

            let description = if !c_payload["description"].is_null() {
                c_payload["description"].as_str().map(|s| s.to_string())
            } else {
                None
            };

            let baggage = if !c_payload["baggage"].is_null() {
                Some(c_payload["baggage"].clone())
            } else {
                None
            };

            let labels = if !c_payload["labels"].is_null() {
                if let Some(labels_str) = c_payload["labels"].as_str() {
                    let parsed_labels: Vec<Label> = labels_str
                        .split(',')
                        .filter(|s| !s.is_empty())
                        .map(|s| {
                            let parts: Vec<&str> = s.split('=').collect();
                            Label {
                                key: parts[0].to_string(),
                                value: parts.get(1).map(|v| v.to_string()),
                            }
                        })
                        .collect();
                    Some(parsed_labels)
                } else {
                    None
                }
            } else {
                None
            };

            let session_request = SessionRequest {
                id: session_id,
                description,
                baggage,
                labels,
            };

            let result = reporter.create_session(session_request);
            if r_status == "success" {
                match result {
                    Ok(session_id) => {
                        let expected_id = r["payload"]["id"].as_str().unwrap();
                        assert_eq!(session_id, expected_id, "incorrect created session id");
                    }
                    Err(e) => panic!("session creation failed: {}", e),
                }
            } else if r_status == "error" {
                match result {
                    Ok(_) => panic!("session creation succeeded, should've failed"),
                    Err(e) => match e {
                        ReporterError::Unknown(msg) => {
                            assert_eq!(
                                GREENER_REPORTER_ERROR as i64,
                                r["payload"]["code"].as_i64().unwrap()
                            );
                            assert_eq!(0, r["payload"]["ingressCode"].as_i64().unwrap());
                            assert_eq!(msg, r["payload"]["message"].as_str().unwrap())
                        }
                        ReporterError::InvalidArgument(msg) => {
                            assert_eq!(
                                GREENER_REPORTER_ERROR_INVALID_ARGUMENT as i64,
                                r["payload"]["code"].as_i64().unwrap()
                            );
                            assert_eq!(0, r["payload"]["ingressCode"].as_i64().unwrap());
                            assert_eq!(msg, r["payload"]["message"].as_str().unwrap())
                        }
                        ReporterError::Ingress(msg, ingress_code) => {
                            assert_eq!(
                                GREENER_REPORTER_ERROR_INGRESS as i64,
                                r["payload"]["code"].as_i64().unwrap()
                            );
                            assert_eq!(
                                ingress_code as i64,
                                r["payload"]["ingressCode"].as_i64().unwrap()
                            );
                            assert_eq!(
                                msg,
                                format!(
                                    "failed session request: {}",
                                    r["payload"]["message"].as_str().unwrap()
                                )
                            )
                        }
                    },
                }
            } else {
                panic!("unknown resp 'status': {}", r_status);
            }
        }
        "report" => {
            let r = &responses["reportResponse"];
            let r_status = r["status"].as_str().unwrap_or("");

            let mut results = Vec::new();
            for tc in c_payload["testcases"].as_array().unwrap() {
                let testcase = TestcaseRequest {
                    session_id: tc["sessionId"].as_str().unwrap().to_string(),
                    testcase_name: tc["testcaseName"].as_str().unwrap().to_string(),
                    testcase_classname: tc["testcaseClassname"].as_str().map(|s| s.to_string()),
                    testcase_file: tc["testcaseFile"].as_str().map(|s| s.to_string()),
                    testsuite: tc["testsuite"].as_str().map(|s| s.to_string()),
                    status: match tc["status"].as_str().unwrap_or("") {
                        "pass" => TestcaseStatus::Pass,
                        "fail" => TestcaseStatus::Fail,
                        "error" => TestcaseStatus::Error,
                        "skip" => TestcaseStatus::Skip,
                        _ => TestcaseStatus::Pass,
                    },
                    output: None,
                    baggage: None,
                };
                let result = reporter.add_testcase(testcase);
                results.push(result);
            }

            let result = results.into_iter().collect::<Result<Vec<_>, _>>();

            if r_status == "success" {
                assert!(
                    result.is_ok(),
                    "testcase creation failed: {:?}",
                    result.err()
                );
            } else if r_status == "error" {
                match result {
                    Ok(_) => panic!("session creation succeeded, should've failed"),
                    Err(e) => match e {
                        ReporterError::Unknown(msg) => {
                            assert_eq!(
                                GREENER_REPORTER_ERROR as i64,
                                r["payload"]["code"].as_i64().unwrap()
                            );
                            assert_eq!(0, r["payload"]["ingressCode"].as_i64().unwrap());
                            assert_eq!(msg, r["payload"]["message"].as_str().unwrap())
                        }
                        ReporterError::InvalidArgument(msg) => {
                            assert_eq!(
                                GREENER_REPORTER_ERROR_INVALID_ARGUMENT as i64,
                                r["payload"]["code"].as_i64().unwrap()
                            );
                            assert_eq!(0, r["payload"]["ingressCode"].as_i64().unwrap());
                            assert_eq!(msg, r["payload"]["message"].as_str().unwrap())
                        }
                        ReporterError::Ingress(msg, ingress_code) => {
                            assert_eq!(
                                GREENER_REPORTER_ERROR_INGRESS as i64,
                                r["payload"]["code"].as_i64().unwrap()
                            );
                            assert_eq!(
                                ingress_code as i64,
                                r["payload"]["ingressCode"].as_i64().unwrap()
                            );
                            assert_eq!(
                                msg,
                                format!(
                                    "failed session request: {}",
                                    r["payload"]["message"].as_str().unwrap()
                                )
                            )
                        }
                    },
                }
            } else {
                panic!("unknown resp 'status': {}", r_status);
            }
        }
        _ => panic!("unknown call 'func': {}", c_func),
    }
}

#[test]
fn test_integration() {
    let fixture_names = get_fixture_names();
    for fixture_name in fixture_names {
        println!("processing fixture: {}", fixture_name);
        process_fixture(&fixture_name);
    }
}
