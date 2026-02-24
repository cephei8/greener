mod batcher;
mod errors;
mod ingress;
mod models;
mod reporter;

pub use errors::ReporterError;
pub use models::Label;
pub use models::SessionRequest;
pub use models::TestcaseRequest;
pub use models::TestcaseStatus;
pub use reporter::Reporter;
use std::ffi::{c_char, CStr, CString};
use std::ptr;

#[repr(C)]
pub struct GreenerReporterSession {
    pub id: *const c_char,
}

#[repr(C)]
pub struct GreenerReporterError {
    pub code: i32,
    pub ingress_code: i32,
    pub message: *const c_char,
}

pub const GREENER_REPORTER_ERROR: i32 = 1;
pub const GREENER_REPORTER_ERROR_INVALID_ARGUMENT: i32 = 2;
pub const GREENER_REPORTER_ERROR_INGRESS: i32 = 3;

fn set_error(err: ReporterError, err_result: *mut *const GreenerReporterError) {
    if err_result.is_null() {
        eprintln!("cannot return error details because greener_reporter_error** is null");
        return;
    }

    let err_result_boxed = match err {
        ReporterError::Unknown(msg) => Box::new(GreenerReporterError {
            code: GREENER_REPORTER_ERROR,
            ingress_code: 0,
            message: CString::new(msg).unwrap().into_raw(),
        }),
        ReporterError::InvalidArgument(msg) => Box::new(GreenerReporterError {
            code: GREENER_REPORTER_ERROR_INVALID_ARGUMENT,
            ingress_code: 0,
            message: CString::new(msg).unwrap().into_raw(),
        }),
        ReporterError::Ingress(msg, ingress_code) => Box::new(GreenerReporterError {
            code: GREENER_REPORTER_ERROR_INGRESS,
            ingress_code: ingress_code.into(),
            message: CString::new(msg).unwrap().into_raw(),
        }),
    };

    unsafe {
        *err_result = Box::into_raw(err_result_boxed);
    }
}

/// Creates a new Reporter instance.
///
/// # Safety
/// The caller must ensure that all pointers are valid if not null.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn greener_reporter_new(
    endpoint: *const c_char,
    api_key: *const c_char,
    error: *mut *const GreenerReporterError,
) -> *mut Reporter {
    unsafe {
        *error = std::ptr::null_mut();
    }

    if endpoint.is_null() {
        set_error(
            ReporterError::InvalidArgument("endpoint pointer is null".into()),
            error,
        );
        return ptr::null_mut();
    }

    if api_key.is_null() {
        set_error(
            ReporterError::InvalidArgument("api_key pointer is null".into()),
            error,
        );
        return ptr::null_mut();
    }

    let endpoint_str = unsafe { CStr::from_ptr(endpoint) }
        .to_string_lossy()
        .to_string();
    let api_key_str = unsafe { CStr::from_ptr(api_key) }
        .to_string_lossy()
        .to_string();

    match Reporter::new(endpoint_str, api_key_str) {
        Ok(reporter) => Box::into_raw(Box::new(reporter)),
        Err(e) => {
            set_error(e, error);
            ptr::null_mut()
        }
    }
}

/// Deletes an Reporter instance.
///
/// # Safety
/// The caller must ensure that all pointers are valid if not null.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn greener_reporter_delete(
    reporter: *mut Reporter,
    error: *mut *const GreenerReporterError,
) {
    unsafe {
        *error = std::ptr::null_mut();
    }
    if !reporter.is_null() {
        unsafe {
            let reporter = Box::from_raw(reporter);
            if let Err(e) = reporter.shutdown() {
                set_error(e, error);
            }
        }
    }
}

/// Creates a new session.
///
/// # Safety
/// The caller must ensure that all pointers are valid if not null.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn greener_reporter_session_create(
    reporter: *mut Reporter,
    session_id: *const c_char,
    description: *const c_char,
    baggage: *const c_char,
    labels: *const c_char,
    error: *mut *const GreenerReporterError,
) -> *const GreenerReporterSession {
    unsafe {
        *error = std::ptr::null_mut();
    }
    if reporter.is_null() {
        set_error(
            ReporterError::InvalidArgument("reporter pointer is null".into()),
            error,
        );
        return ptr::null();
    }

    let reporter = unsafe { &*reporter };

    let session_id_opt = if !session_id.is_null() {
        Some(
            unsafe { CStr::from_ptr(session_id) }
                .to_string_lossy()
                .to_string(),
        )
    } else {
        None
    };

    let description_opt = if !description.is_null() {
        Some(
            unsafe { CStr::from_ptr(description) }
                .to_string_lossy()
                .to_string(),
        )
    } else {
        None
    };

    let baggage_opt = if !baggage.is_null() {
        let baggage_str = unsafe { CStr::from_ptr(baggage) }
            .to_string_lossy()
            .to_string();
        match serde_json::from_str(&baggage_str) {
            Ok(v) => Some(v),
            Err(e) => {
                set_error(
                    ReporterError::InvalidArgument(format!("cannot parse baggage: {}", e)),
                    error,
                );
                return ptr::null();
            }
        }
    } else {
        None
    };

    let labels_opt = if !labels.is_null() {
        let labels_str = unsafe { CStr::from_ptr(labels) }
            .to_string_lossy()
            .to_string();
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
        if parsed_labels.is_empty() {
            None
        } else {
            Some(parsed_labels)
        }
    } else {
        None
    };

    let session = SessionRequest {
        id: session_id_opt,
        description: description_opt,
        baggage: baggage_opt,
        labels: labels_opt,
    };

    match reporter.create_session(session) {
        Ok(created_session_id) => {
            let session = Box::new(GreenerReporterSession {
                id: CString::new(created_session_id).unwrap().into_raw(),
            });
            Box::into_raw(session)
        }
        Err(e) => {
            set_error(e, error);
            ptr::null()
        }
    }
}

/// Creates a new testcase.
///
/// # Safety
/// The caller must ensure that all pointers are valid if not null.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn greener_reporter_testcase_create(
    reporter: *mut Reporter,
    session_id: *const c_char,
    testcase_name: *const c_char,
    testcase_classname: *const c_char,
    testcase_file: *const c_char,
    testsuite: *const c_char,
    status: *const c_char,
    output: *const c_char,
    baggage: *const c_char,
    error: *mut *const GreenerReporterError,
) {
    unsafe {
        *error = std::ptr::null_mut();
    }
    if reporter.is_null() {
        set_error(
            ReporterError::InvalidArgument("reporter pointer is null".into()),
            error,
        );
        return;
    }

    let reporter = unsafe { &*reporter };

    let session_id = unsafe { CStr::from_ptr(session_id) }
        .to_string_lossy()
        .to_string();
    let testcase_name = unsafe { CStr::from_ptr(testcase_name) }
        .to_string_lossy()
        .to_string();

    let testcase_classname = if !testcase_classname.is_null() {
        Some(
            unsafe { CStr::from_ptr(testcase_classname) }
                .to_string_lossy()
                .to_string(),
        )
    } else {
        None
    };
    let testcase_file = if !testcase_file.is_null() {
        Some(
            unsafe { CStr::from_ptr(testcase_file) }
                .to_string_lossy()
                .to_string(),
        )
    } else {
        None
    };
    let testsuite = if !testsuite.is_null() {
        Some(
            unsafe { CStr::from_ptr(testsuite) }
                .to_string_lossy()
                .to_string(),
        )
    } else {
        None
    };
    let output = if !output.is_null() {
        Some(
            unsafe { CStr::from_ptr(output) }
                .to_string_lossy()
                .to_string(),
        )
    } else {
        None
    };
    let baggage = if !baggage.is_null() {
        let json_str = unsafe { CStr::from_ptr(baggage) }
            .to_string_lossy()
            .to_string();
        Some(match serde_json::from_str(json_str.as_str()) {
            Ok(x) => x,
            Err(e) => {
                set_error(
                    ReporterError::InvalidArgument(format!("cannot parse baggage: {}", e)),
                    error,
                );
                return;
            }
        })
    } else {
        None
    };

    let status = match unsafe { CStr::from_ptr(status) }
        .to_string_lossy()
        .to_string()
        .as_str()
    {
        "pass" => TestcaseStatus::Pass,
        "fail" => TestcaseStatus::Fail,
        "error" => TestcaseStatus::Error,
        "skip" => TestcaseStatus::Skip,
        x => {
            set_error(
                ReporterError::InvalidArgument(format!("invalid testcase status: {}", x)),
                error,
            );
            return;
        }
    };

    let testcase = TestcaseRequest {
        session_id,
        testcase_name,
        testcase_classname,
        testcase_file,
        testsuite,
        status,
        output,
        baggage,
    };

    if let Err(e) = reporter.add_testcase(testcase) {
        set_error(e, error);
    }
}

/// Deletes an error instance.
///
/// # Safety
/// The caller must ensure that all pointers are valid if not null.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn greener_reporter_error_delete(error: *const GreenerReporterError) {
    if !error.is_null() {
        return;
    }

    let error = unsafe { Box::from_raw(error as *mut GreenerReporterError) };
    if !error.message.is_null() {
        unsafe {
            let _ = CString::from_raw(error.message as *mut c_char);
        }
    }
}

/// Deletes a session instance.
///
/// # Safety
/// The caller must ensure that all pointers are valid if not null.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn greener_reporter_session_delete(session: *const GreenerReporterSession) {
    if session.is_null() {
        return;
    }

    let session = unsafe { Box::from_raw(session as *mut GreenerReporterSession) };
    if !session.id.is_null() {
        unsafe {
            let _ = CString::from_raw(session.id as *mut c_char);
        }
    }
}

/// Deletes a session instance.
///
/// # Safety
/// The caller must ensure that all pointers are valid if not null.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn greener_reporter_report_error_pop(
    reporter: *mut Reporter,
    error: *mut *const GreenerReporterError,
) {
    unsafe {
        *error = std::ptr::null_mut();
    }
    if reporter.is_null() {
        set_error(
            ReporterError::InvalidArgument("reporter pointer is null".into()),
            error,
        );
        return;
    }

    let reporter = unsafe { &*reporter };
    if let Some(e) = reporter.pop_error() {
        set_error(e, error);
    }
}
