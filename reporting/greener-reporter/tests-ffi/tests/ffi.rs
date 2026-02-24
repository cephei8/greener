use serde_json::Value;
use std::ffi::{CStr, CString};
use std::ptr;
use tests_ffi::*;

fn get_fixture_names_list() -> Vec<String> {
    unsafe {
        let servermock = greener_servermock_new();
        let mut names: *mut *const ::std::os::raw::c_char = ptr::null_mut();
        let mut num_names: u32 = 0;
        let mut error: *const greener_servermock_error = ptr::null();
        greener_servermock_fixture_names(servermock, &mut names, &mut num_names, &mut error);
        if !error.is_null() {
            let msg = CStr::from_ptr((*error).message).to_string_lossy();
            panic!("failed to get fixture names: {}", msg);
        }
        if num_names == 0 {
            panic!("no fixtures found");
        }
        let mut result = Vec::with_capacity(num_names as usize);
        for i in 0..num_names {
            let name_ptr = *names.offset(i as isize);
            let name = CStr::from_ptr(name_ptr).to_string_lossy().into_owned();
            result.push(name);
        }
        result
    }
}

fn run_integration_test() {
    let fixture_names = get_fixture_names_list();
    for fixture_name in fixture_names {
        println!("processing fixture {}", fixture_name);
        process_fixture(&fixture_name);
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn test_integration() {
        run_integration_test();
    }
}

fn process_fixture(fixture_name: &str) {
    unsafe {
        let servermock = greener_servermock_new();
        let name_c = CString::new(fixture_name).unwrap();
        let mut calls: *const ::std::os::raw::c_char = ptr::null();
        let mut responses: *const ::std::os::raw::c_char = ptr::null();
        let mut error: *const greener_servermock_error = ptr::null();
        greener_servermock_fixture_calls(servermock, name_c.as_ptr(), &mut calls, &mut error);
        if !error.is_null() {
            let msg = CStr::from_ptr((*error).message).to_string_lossy();
            panic!("failed to get fixture calls: {}", msg);
        }
        greener_servermock_fixture_responses(
            servermock,
            name_c.as_ptr(),
            &mut responses,
            &mut error,
        );
        if !error.is_null() {
            let msg = CStr::from_ptr((*error).message).to_string_lossy();
            panic!("failed to get fixture responses: {}", msg);
        }
        greener_servermock_serve(servermock, responses, &mut error);
        if !error.is_null() {
            let msg = CStr::from_ptr((*error).message).to_string_lossy();
            panic!("failed to serve: {}", msg);
        }
        let port = greener_servermock_get_port(servermock, &mut error);
        if !error.is_null() {
            let msg = CStr::from_ptr((*error).message).to_string_lossy();
            panic!("failed to get port: {}", msg);
        }
        let endpoint = format!("http://127.0.0.1:{}", port);
        let endpoint_c = CString::new(endpoint).unwrap();
        let api_key_c = CString::new("some-api-token").unwrap();

        let mut reporter_error: *const greener_reporter_error = ptr::null();
        let reporter = greener_reporter_new(
            endpoint_c.as_ptr(),
            api_key_c.as_ptr(),
            &mut reporter_error as *mut _,
        );
        if !reporter_error.is_null() {
            let msg = CStr::from_ptr((*reporter_error).message).to_string_lossy();
            panic!("failed to create reporter: {}", msg);
        }
        let calls_str = CStr::from_ptr(calls).to_string_lossy();
        let responses_str = CStr::from_ptr(responses).to_string_lossy();
        let calls_json: Value = serde_json::from_str(&calls_str).unwrap();
        let calls_array = calls_json["calls"]
            .as_array()
            .expect("invalid calls format");
        for call in calls_array {
            make_call(reporter, call, &responses_str);
        }
        let mut del_error: *const greener_reporter_error = ptr::null();
        greener_reporter_delete(reporter, &mut del_error as *mut _);
        if !del_error.is_null() {
            let msg = CStr::from_ptr((*del_error).message).to_string_lossy();
            panic!("failed to delete reporter: {}", msg);
        }
        let mut assert_error: *const greener_servermock_error = ptr::null();
        greener_servermock_assert(
            servermock,
            calls,
            &mut assert_error as *mut *const greener_servermock_error,
        );
        if !assert_error.is_null() {
            let msg = CStr::from_ptr((*assert_error).message).to_string_lossy();
            panic!("assert failed: {}", msg);
        }
    }
}

fn make_call(reporter: *mut greener_reporter, call: &Value, responses: &str) {
    unsafe {
        let func = call["func"].as_str().expect("missing 'func' in call");
        let payload = &call["payload"];
        match func {
            "createSession" => {
                let responses_json: Value = serde_json::from_str(responses).unwrap();
                let response = &responses_json["createSessionResponse"];
                let status = response["status"]
                    .as_str()
                    .expect("missing status in response");

                let session_id_c = payload["id"].as_str().map(|s| CString::new(s).unwrap());
                let description_c = payload["description"]
                    .as_str()
                    .map(|s| CString::new(s).unwrap());
                let baggage_c = if !payload["baggage"].is_null() {
                    let baggage_json = serde_json::to_string(&payload["baggage"]).unwrap();
                    Some(CString::new(baggage_json).unwrap())
                } else {
                    None
                };
                let labels_c = payload["labels"].as_str().map(|s| CString::new(s).unwrap());

                let session_id_ptr = session_id_c.as_ref().map_or(ptr::null(), |c| c.as_ptr());
                let description_ptr = description_c
                    .as_ref()
                    .map_or(ptr::null(), |c| c.as_ptr());
                let baggage_ptr = baggage_c.as_ref().map_or(ptr::null(), |c| c.as_ptr());
                let labels_ptr = labels_c.as_ref().map_or(ptr::null(), |c| c.as_ptr());

                match status {
                    "success" => {
                        let mut error: *const greener_reporter_error = ptr::null();
                        let session = greener_reporter_session_create(
                            reporter,
                            session_id_ptr,
                            description_ptr,
                            baggage_ptr,
                            labels_ptr,
                            &mut error as *mut _,
                        );
                        if !error.is_null() {
                            let msg = CStr::from_ptr((*error).message).to_string_lossy();
                            panic!("failed to create session: {}", msg);
                        }
                        let expected_id = response["payload"]["id"]
                            .as_str()
                            .expect("missing expected session id");
                        let actual_id = CStr::from_ptr((*session).id).to_string_lossy();
                        if actual_id != expected_id {
                            panic!(
                                "incorrect created session id: actual {}, expected {}",
                                actual_id, expected_id
                            );
                        }
                        greener_reporter_session_delete(session);
                    }
                    "error" => {
                        let mut error: *const greener_reporter_error = ptr::null();
                        let _session = greener_reporter_session_create(
                            reporter,
                            session_id_ptr,
                            description_ptr,
                            baggage_ptr,
                            labels_ptr,
                            &mut error as *mut _,
                        );
                        if error.is_null() {
                            panic!("session creation succeeded, should've failed");
                        }
                        let expected_message = response["payload"]["message"]
                            .as_str()
                            .expect("missing expected error message");
                        let error_msg = CStr::from_ptr((*error).message).to_string_lossy();
                        if !error_msg
                            .contains(&format!("failed session request: {}", expected_message))
                        {
                            panic!("incorrect error message: actual '{}', expected to contain 'failed session request: {}'", error_msg, expected_message);
                        }
                    }
                    _ => panic!("unknown response status: {}", status),
                }
            }
            "report" => {
                let responses_json: Value = serde_json::from_str(responses).unwrap();
                let response = &responses_json["reportResponse"];
                let status = response["status"]
                    .as_str()
                    .expect("missing status in response");

                let mut errors = Vec::new();
                for p in payload["testcases"].as_array().unwrap() {
                    let session_id = p["sessionId"]
                        .as_str()
                        .expect("missing sessionId in payload");
                    let testcase_name = p["testcaseName"]
                        .as_str()
                        .expect("missing testcaseName in payload");
                    let test_status = p["status"].as_str().expect("missing status in payload");
                    let testcase_classname = p["testcaseClassname"].as_str();
                    let testcase_file = p["testcaseFile"].as_str();
                    let testsuite = p["testsuite"].as_str();

                    let session_id_c = CString::new(session_id).unwrap();
                    let testcase_name_c = CString::new(testcase_name).unwrap();
                    let status_c = CString::new(test_status).unwrap();
                    let testcase_classname_c = testcase_classname.map(|s| CString::new(s).unwrap());
                    let testcase_file_c = testcase_file.map(|s| CString::new(s).unwrap());
                    let testsuite_c = testsuite.map(|s| CString::new(s).unwrap());
                    let mut testcase_classname_ptr = ptr::null();
                    let mut testcase_file_ptr = ptr::null();
                    let mut testsuite_ptr = ptr::null();
                    if let Some(ref cstr) = testcase_classname_c {
                        testcase_classname_ptr = cstr.as_ptr();
                    }
                    if let Some(ref cstr) = testcase_file_c {
                        testcase_file_ptr = cstr.as_ptr();
                    }
                    if let Some(ref cstr) = testsuite_c {
                        testsuite_ptr = cstr.as_ptr();
                    }

                    let mut error: *const greener_reporter_error = ptr::null();
                    greener_reporter_testcase_create(
                        reporter,
                        session_id_c.as_ptr(),
                        testcase_name_c.as_ptr(),
                        testcase_classname_ptr,
                        testcase_file_ptr,
                        testsuite_ptr,
                        status_c.as_ptr(),
                        ptr::null(),
                        ptr::null(),
                        &mut error as *mut _,
                    );

                    errors.push(error);
                }

                let error: *const greener_reporter_error = if errors.len() > 0 {
                    // server responses are stubbed, so take the first one
                    errors[0]
                } else {
                    ptr::null()
                };

                match status {
                    "success" => {
                        if !error.is_null() {
                            let msg = CStr::from_ptr((*error).message).to_string_lossy();
                            panic!("failed to create testcase: {}", msg);
                        }
                    }
                    "error" => {
                        if error.is_null() {
                            panic!("testcase creation succeeded, should've failed");
                        }
                        let expected_message = response["payload"]["message"]
                            .as_str()
                            .expect("missing expected error message");
                        let error_msg = CStr::from_ptr((*error).message).to_string_lossy();
                        if !error_msg
                            .contains(&format!("failed testcase request: {}", expected_message))
                        {
                            panic!("incorrect error message: actual '{}', expected to contain 'failed testcase request: {}'", error_msg, expected_message);
                        }
                    }
                    _ => panic!("unknown response status: {}", status),
                }
            }
            _ => panic!("unknown function: {}", func),
        }
    }
}
