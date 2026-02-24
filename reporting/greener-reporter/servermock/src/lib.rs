mod fixtures;
mod server;

use std::ffi::{CStr, CString, c_char};
use std::ptr;
use std::sync::Arc;
use tokio::sync::Mutex;
use serde_json::json;
use tokio::runtime::Runtime;
use std::alloc::{Layout, alloc, dealloc};

#[derive(Clone, Debug, serde::Serialize, serde::Deserialize)]
pub struct ApiCalls {
    pub calls: Vec<ApiCall>,
}

#[derive(Clone, Debug, serde::Serialize, serde::Deserialize)]
pub struct ApiCall {
    pub func: String,
    pub payload: serde_json::Value,
}

#[derive(Default)]
struct Fixture {
    calls: String,
    responses: String,
}

#[repr(C)]
pub struct GreenerServermock {
    runtime: Arc<Runtime>,
    port: i32,
    responses: String,
    calls: Vec<String>,
    fixtures: Vec<(String, Fixture)>,
    recorded_calls: Arc<Mutex<Vec<ApiCall>>>,

    fixture_calls_cache: Vec<(*const c_char, *const c_char)>,
    fixture_responses_cache: Vec<(*const c_char, *const c_char)>,
    fixture_names_cache: Option<(*const *const c_char, u32)>,
}

impl Default for GreenerServermock {
    fn default() -> Self {
        Self::new()
    }
}

impl GreenerServermock {
    pub fn new() -> Self {
        let runtime = Arc::new(Runtime::new().unwrap());
        let mut servermock = Self {
            runtime,
            port: -1,
            responses: String::new(),
            calls: Vec::new(),
            fixtures: Vec::new(),
            recorded_calls: Arc::new(Mutex::new(Vec::new())),
            fixture_calls_cache: Vec::new(),
            fixture_responses_cache: Vec::new(),
            fixture_names_cache: None,
        };

        initialize_test_fixtures(&mut servermock);
        servermock
    }

    pub fn port(&self) -> i32 {
        self.port
    }

    pub fn serve(&mut self, responses: &str) -> Result<(), String> {
        self.responses = responses.to_string();
        let assigned_port = server::start_server(
            self.runtime.clone(),
            responses.to_string(),
            self.recorded_calls.clone(),
        )
        .map_err(|e| format!("failed to start server: {}", e))?;
        self.port = assigned_port as i32;
        Ok(())
    }

    pub fn assert(&mut self, expected_calls: &str) -> Result<(), String> {
        let recorded = self
            .runtime
            .block_on(async { self.recorded_calls.lock().await.clone() });

        let expected: ApiCalls = serde_json::from_str(expected_calls)
            .map_err(|e| format!("failed to parse expected calls: {}", e))?;

        if recorded.len() != expected.calls.len() {
            return Err(format!(
                "call count mismatch. expected {} calls but got {}.\nexpected: {:#?}\nactual: {:#?}",
                expected.calls.len(),
                recorded.len(),
                expected,
                recorded
            ));
        }

        for (i, (expected, actual)) in expected.calls.iter().zip(recorded.iter()).enumerate() {
            if expected.func != actual.func {
                return Err(format!(
                    "call {} function mismatch. Expected '{}' but got '{}'",
                    i, expected.func, actual.func
                ));
            }

            // Compare payloads ignoring whitespace and formatting
            let expected_payload = serde_json::to_value(&expected.payload)
                .map_err(|e| format!("failed to serialize expected payload: {}", e))?;
            let actual_payload = serde_json::to_value(&actual.payload)
                .map_err(|e| format!("failed to serialize actual payload: {}", e))?;

            if expected_payload != actual_payload {
                return Err(format!(
                    "call {} payload mismatch.\nexpected: {}\nactual: {}",
                    i,
                    serde_json::to_string_pretty(&expected_payload).unwrap(),
                    serde_json::to_string_pretty(&actual_payload).unwrap()
                ));
            }
        }

        Ok(())
    }

    pub fn fixture_names(&mut self) -> Vec<String> {
        self.fixtures.iter().map(|(name, _)| name.clone()).collect()
    }

    pub fn fixture_calls(&mut self, name: &str) -> Option<String> {
        self.fixtures
            .iter()
            .find(|(fname, _)| fname == name)
            .map(|(_, fixture)| fixture.calls.clone())
    }

    pub fn fixture_responses(&mut self, name: &str) -> Option<String> {
        self.fixtures
            .iter()
            .find(|(fname, _)| fname == name)
            .map(|(_, fixture)| fixture.responses.clone())
    }
}

#[repr(C)]
pub struct GreenerServermockError {
    pub message: *const c_char,
}

fn string_to_c_char(s: &str) -> *const c_char {
    CString::new(s).unwrap().into_raw()
}

fn initialize_test_fixtures(servermock: &mut GreenerServermock) {
    let fixtures = crate::fixtures::create_fixtures();
    for (name, fixture) in fixtures {
        let calls = json!({ "calls": fixture.calls }).to_string();
        let responses = json!({
            "createSessionResponse": {
                "status": fixture.responses.create_session_response.status,
                "payload": fixture.responses.create_session_response.payload
            },
            "reportResponse": {
                "status": fixture.responses.report_response.status,
                "payload": fixture.responses.report_response.payload
            }
        }).to_string();

        servermock.fixtures.push((
            name,
            Fixture {
                calls,
                responses,
            },
        ));
    }
}

/// Creates a new servermock instance.
///
/// # Safety
/// The caller must ensure that all pointers are valid if not null.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn greener_servermock_new() -> *mut GreenerServermock {
    let port = 8080
        + (std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_millis()
            % 1000) as i32;

    let runtime = Arc::new(Runtime::new().unwrap());
    let mut servermock = Box::new(GreenerServermock {
        runtime,
        port,
        responses: String::new(),
        calls: Vec::new(),
        fixtures: Vec::new(),
        recorded_calls: Arc::new(Mutex::new(Vec::new())),
        fixture_calls_cache: Vec::new(),
        fixture_responses_cache: Vec::new(),
        fixture_names_cache: None,
    });

    initialize_test_fixtures(&mut servermock);

    Box::into_raw(servermock)
}

/// Deletes a servermock instance.
///
/// # Safety
/// The caller must ensure that all pointers are valid if not null.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn greener_servermock_delete(
    ctx: *mut GreenerServermock,
    error: *mut *const GreenerServermockError,
) {
    unsafe{ *error = std::ptr::null_mut(); }
    if !ctx.is_null() {
        let servermock = unsafe { &mut *ctx };

        for (ptr, _) in servermock.fixture_calls_cache.drain(..) {
            if !ptr.is_null() {
                let _ = unsafe { CString::from_raw(ptr as *mut c_char) };
            }
        }

        for (ptr, _) in servermock.fixture_responses_cache.drain(..) {
            if !ptr.is_null() {
                let _ = unsafe { CString::from_raw(ptr as *mut c_char) };
            }
        }

        if let Some((names_ptr, num_names)) = servermock.fixture_names_cache.take() {
            if !names_ptr.is_null() {
                for i in 0..num_names {
                    let string_ptr = unsafe { *names_ptr.add(i as usize) };
                    if !string_ptr.is_null() {
                        let _ = unsafe { CString::from_raw(string_ptr as *mut c_char) };
                    }
                }
                let layout = Layout::array::<*const c_char>(num_names as usize).unwrap();
                unsafe { dealloc(names_ptr as *mut u8, layout) };
            }
        }

        let _ = unsafe { Box::from_raw(ctx) };
    }
}

/// Serves responses from the servermock.
///
/// # Safety
/// The caller must ensure that all pointers are valid if not null.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn greener_servermock_serve(
    ctx: *mut GreenerServermock,
    responses: *const c_char,
    error: *mut *const GreenerServermockError,
) {
    unsafe{ *error = std::ptr::null_mut(); }
    if ctx.is_null() || responses.is_null() {
        if !error.is_null() {
            let err = Box::new(GreenerServermockError {
                message: string_to_c_char("context or responses pointer is null"),
            });
            unsafe {
                *error = Box::into_raw(err);
            }
        }
        return;
    }

    let responses_str = unsafe { CStr::from_ptr(responses).to_string_lossy().into_owned() };
    let ctx_ref = unsafe { &mut *ctx };

    if let Err(e) = ctx_ref.serve(&responses_str) {
        if !error.is_null() {
            let err = Box::new(GreenerServermockError {
                message: string_to_c_char(&format!("failed to start server: {}", e)),
            });
            unsafe {
                *error = Box::into_raw(err);
            }
        }
    }
}

/// Gets the port number of the servermock.
///
/// # Safety
/// The caller must ensure that all pointers are valid if not null.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn greener_servermock_get_port(
    ctx: *mut GreenerServermock,
    error: *mut *const GreenerServermockError,
) -> i32 {
    unsafe{ *error = std::ptr::null_mut(); }
    if ctx.is_null() {
        if !error.is_null() {
            let err = Box::new(GreenerServermockError {
                message: string_to_c_char("context pointer is null"),
            });
            unsafe {
                *error = Box::into_raw(err);
            }
        }
        return -1;
    }

    let ctx_ref = unsafe { &*ctx };
    ctx_ref.port
}

/// Asserts that the servermock received the expected calls.
///
/// # Safety
/// The caller must ensure that all pointers are valid if not null.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn greener_servermock_assert(
    ctx: *mut GreenerServermock,
    calls: *const c_char,
    error: *mut *const GreenerServermockError,
) -> bool {
    unsafe{ *error = std::ptr::null_mut(); }
    if ctx.is_null() || calls.is_null() {
        if !error.is_null() {
            let err = Box::new(GreenerServermockError {
                message: string_to_c_char("context or calls pointer is null"),
            });
            unsafe {
                *error = Box::into_raw(err);
            }
        }
        return false;
    }

    let calls_str = unsafe { CStr::from_ptr(calls).to_string_lossy().into_owned() };
    let ctx_ref = unsafe { &mut *ctx };

    match ctx_ref.assert(&calls_str) {
        Ok(_) => true,
        Err(e) => {
            if !error.is_null() {
                let err = Box::new(GreenerServermockError {
                    message: string_to_c_char(&e),
                });
                unsafe {
                    *error = Box::into_raw(err);
                }
            }
            false
        }
    }
}

/// Gets the names of all fixtures in the servermock.
///
/// # Safety
/// The caller must ensure that all pointers are valid if not null.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn greener_servermock_fixture_names(
    ctx: *mut GreenerServermock,
    names: *mut *const *const c_char,
    num_names: *mut u32,
    error: *mut *const GreenerServermockError,
) {
    unsafe{ *error = std::ptr::null_mut(); }
    if ctx.is_null() || names.is_null() || num_names.is_null() {
        if !error.is_null() {
            let err = Box::new(GreenerServermockError {
                message: string_to_c_char("context, names, or num_names pointer is null"),
            });
            unsafe {
                *error = Box::into_raw(err);
            }
        }
        return;
    }

    let ctx_mut = unsafe { &mut *ctx };

    if let Some((cached_names, cached_num_names)) = ctx_mut.fixture_names_cache {
        unsafe {
            *names = cached_names;
            *num_names = cached_num_names;
        }
        return;
    }

    let fixture_names: Vec<String> = ctx_mut.fixtures.iter().map(|(name, _)| name.clone()).collect();
    let num_fixtures = fixture_names.len() as u32;

    if num_fixtures == 0 {
        unsafe {
            *names = ptr::null();
            *num_names = 0;
        }
        return;
    }

    let layout = Layout::array::<*const c_char>(num_fixtures as usize).unwrap();
    let names_ptr = unsafe { alloc(layout) as *mut *const c_char };

    if names_ptr.is_null() {
        if !error.is_null() {
            let err = Box::new(GreenerServermockError {
                message: string_to_c_char("failed to allocate memory for fixture names"),
            });
            unsafe {
                *error = Box::into_raw(err);
            }
        }
        return;
    }

    for (i, name) in fixture_names.iter().enumerate() {
        match CString::new(name.as_str()) {
            Ok(c_str) => unsafe {
                *names_ptr.add(i) = c_str.into_raw();
            },
            Err(_) => {
                if !error.is_null() {
                    let err = Box::new(GreenerServermockError {
                        message: string_to_c_char("failed to create C string for fixture name"),
                    });
                    unsafe {
                        *error = Box::into_raw(err);
                    }
                }
                unsafe { dealloc(names_ptr as *mut u8, layout) };
                return;
            }
        }
    }

    ctx_mut.fixture_names_cache = Some((names_ptr, num_fixtures));

    unsafe {
        *names = names_ptr;
        *num_names = num_fixtures;
    }
}

/// Gets the fixture calls for a specific fixture.
///
/// # Safety
/// The caller must ensure that all pointers are valid if not null.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn greener_servermock_fixture_calls(
    ctx: *mut GreenerServermock,
    fixture_name: *const c_char,
    calls: *mut *const c_char,
    error: *mut *const GreenerServermockError,
) {
    unsafe{ *error = std::ptr::null_mut(); }
    if ctx.is_null() || fixture_name.is_null() || calls.is_null() {
        if !error.is_null() {
            let err = Box::new(GreenerServermockError {
                message: string_to_c_char("context, fixture_name, or calls pointer is null"),
            });
            unsafe {
                *error = Box::into_raw(err);
            }
        }
        return;
    }

    let ctx_mut = unsafe { &mut *ctx };
    let fixture_name_str = unsafe { CStr::from_ptr(fixture_name).to_string_lossy() };
    let fixture_name_str = fixture_name_str.to_string();

    if let Some((calls_ptr, _)) = ctx_mut
        .fixture_calls_cache
        .iter()
        .find(|(name_ptr, _)| unsafe {
            CStr::from_ptr(*name_ptr).to_string_lossy() == fixture_name_str
        })
    {
        unsafe {
            *calls = *calls_ptr;
        }
        return;
    }

    if let Some((_, fixture)) = ctx_mut
        .fixtures
        .iter()
        .find(|(name, _)| name == &fixture_name_str)
    {
        match CString::new(fixture.calls.as_str()) {
            Ok(c_str) => {
                let calls_ptr = c_str.into_raw();
                let name_cstring = CString::new(fixture_name_str).unwrap();
                let name_ptr = name_cstring.into_raw();
                
                ctx_mut.fixture_calls_cache.push((name_ptr, calls_ptr));
                unsafe {
                    *calls = calls_ptr;
                }
            }
            Err(_) => {
                if !error.is_null() {
                    let err = Box::new(GreenerServermockError {
                        message: string_to_c_char("failed to create C string for fixture calls"),
                    });
                    unsafe {
                        *error = Box::into_raw(err);
                    }
                }
            }
        }
    } else if !error.is_null() {
        let err = Box::new(GreenerServermockError {
            message: string_to_c_char("fixture not found"),
        });
        unsafe {
            *error = Box::into_raw(err);
        }
    }
}

/// Gets the fixture responses for a specific fixture.
///
/// # Safety
/// The caller must ensure that all pointers are valid if not null.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn greener_servermock_fixture_responses(
    ctx: *mut GreenerServermock,
    fixture_name: *const c_char,
    responses: *mut *const c_char,
    error: *mut *const GreenerServermockError,
) {
    unsafe{ *error = std::ptr::null_mut(); }
    if ctx.is_null() || fixture_name.is_null() || responses.is_null() {
        if !error.is_null() {
            let err = Box::new(GreenerServermockError {
                message: string_to_c_char("context, fixture_name, or responses pointer is null"),
            });
            unsafe {
                *error = Box::into_raw(err);
            }
        }
        return;
    }

    let ctx_mut = unsafe { &mut *ctx };
    let fixture_name_str = unsafe { CStr::from_ptr(fixture_name).to_string_lossy() };
    let fixture_name_str = fixture_name_str.to_string();

    if let Some((responses_ptr, _)) = ctx_mut
        .fixture_responses_cache
        .iter()
        .find(|(name_ptr, _)| unsafe {
            CStr::from_ptr(*name_ptr).to_string_lossy() == fixture_name_str
        })
    {
        unsafe {
            *responses = *responses_ptr;
        }
        return;
    }

    if let Some((_, fixture)) = ctx_mut
        .fixtures
        .iter()
        .find(|(name, _)| name == &fixture_name_str)
    {
        match CString::new(fixture.responses.as_str()) {
            Ok(c_str) => {
                let responses_ptr = c_str.into_raw();
                let name_cstring = CString::new(fixture_name_str).unwrap();
                let name_ptr = name_cstring.into_raw();
                
                ctx_mut.fixture_responses_cache.push((name_ptr, responses_ptr));
                unsafe {
                    *responses = responses_ptr;
                }
            }
            Err(_) => {
                if !error.is_null() {
                    let err = Box::new(GreenerServermockError {
                        message: string_to_c_char("failed to create C string for fixture responses"),
                    });
                    unsafe {
                        *error = Box::into_raw(err);
                    }
                }
            }
        }
    } else if !error.is_null() {
        let err = Box::new(GreenerServermockError {
            message: string_to_c_char("fixture not found"),
        });
        unsafe {
            *error = Box::into_raw(err);
        }
    }
}

/// Deletes a GreenerServermockError object.
///
/// # Safety
/// The caller must ensure that all pointers are valid if not null.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn greener_servermock_error_delete(error: *const GreenerServermockError) {
    if !error.is_null() {
        let _error_box = unsafe { Box::from_raw(error as *mut GreenerServermockError) };
    }
}
