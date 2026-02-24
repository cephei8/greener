use axum::{
    Json, Router,
    extract::State,
    http::{HeaderMap, HeaderValue, StatusCode},
    routing::post,
};
use serde_json::{Value, json};
use std::net::SocketAddr;
use std::sync::Arc;
use tokio::sync::Mutex;

#[derive(Debug, Default)]
struct ServerState {
    responses: Value,
}

type SharedState = (Arc<Mutex<ServerState>>, Arc<Mutex<Vec<ApiCall>>>);

use crate::ApiCall;

pub fn start_server(
    runtime: Arc<tokio::runtime::Runtime>,
    responses: String,
    recorded_calls: Arc<Mutex<Vec<ApiCall>>>,
) -> Result<u16, String> {
    let responses_json: Value = serde_json::from_str(&responses).unwrap();

    let state = Arc::new(Mutex::new(ServerState {
        responses: responses_json,
    }));
    let shared_state = (state, recorded_calls.clone());

    let addr = SocketAddr::from(([127, 0, 0, 1], 0));

    let (listener, port) = runtime.block_on(async {
        match tokio::net::TcpListener::bind(addr).await {
            Ok(listener) => {
                let local_addr = listener
                    .local_addr()
                    .map_err(|e| format!("Failed to get local address: {}", e))?;
                Ok((listener, local_addr.port()))
            }
            Err(e) => Err(format!("Failed to bind to address: {}", e)),
        }
    })?;

    let app = Router::new()
        .route("/api/v1/ingress/sessions", post(create_session))
        .route("/api/v1/ingress/testcases", post(create_testcases))
        .with_state(shared_state);

    runtime.spawn(async move {
        axum::serve(listener, app).await.unwrap();
    });

    Ok(port)
}

#[axum::debug_handler]
async fn create_session(
    State((state, recorded_calls)): State<SharedState>,
    Json(mut session): Json<Value>,
) -> (StatusCode, HeaderMap, Json<Value>) {
    let state = state.lock().await;
    let mut calls = recorded_calls.lock().await;

    if let Some(labels) = session.get("labels") {
        if !labels.is_null() {
            if let Some(labels_array) = labels.as_array() {
                let labels_str = labels_array
                    .iter()
                    .filter_map(|label| {
                        if let (Some(key), value) = (label.get("key")?.as_str(), label.get("value"))
                        {
                            match value {
                                Some(v) if v.is_string() => {
                                    Some(format!("{}={}", key, v.as_str().unwrap()))
                                }
                                _ => Some(key.to_string()),
                            }
                        } else {
                            None
                        }
                    })
                    .collect::<Vec<_>>()
                    .join(",");

                session["labels"] = json!(labels_str);
            }
        } else {
            session["labels"] = json!(null);
        }
    }

    calls.push(ApiCall {
        func: "createSession".to_string(),
        payload: session,
    });

    let create_session_response = state.responses.get("createSessionResponse").unwrap();

    let status = create_session_response
        .get("status")
        .and_then(|s| s.as_str())
        .unwrap();

    let payload = create_session_response.get("payload").unwrap();

    match status {
        "success" => {
            let id = payload.get("id").unwrap();
            (
                StatusCode::OK,
                json_content_type(),
                Json(json!({ "id": id })),
            )
        }
        "error" => (
            StatusCode::BAD_REQUEST,
            json_content_type(),
            Json(json!({
                "code": payload.get("code").and_then(|c| c.as_i64()).unwrap(),
                "ingressCode": payload.get("ingressCode").and_then(|c| c.as_i64()).unwrap(),
                "message": payload.get("message").and_then(|m| m.as_str()).unwrap()
            })),
        ),
        _ => panic!(),
    }
}

#[axum::debug_handler]
async fn create_testcases(
    State((state, recorded_calls)): State<SharedState>,
    Json(testcase): Json<Value>,
) -> (StatusCode, HeaderMap, Json<Value>) {
    let mut response_headers = HeaderMap::new();
    response_headers.insert("Content-Type", HeaderValue::from_static("application/json"));

    let state = state.lock().await;
    let mut calls = recorded_calls.lock().await;

    calls.push(ApiCall {
        func: "report".to_string(),
        payload: testcase.clone(),
    });

    let status = state
        .responses
        .get("status")
        .and_then(Value::as_str)
        .unwrap_or("success");

    match status {
        "success" => (StatusCode::OK, json_content_type(), Json::default()),
        "error" => {
            let report_response = state.responses.get("reportResponse").unwrap();
            let payload = report_response.get("payload").unwrap();

            (
                StatusCode::BAD_REQUEST,
                json_content_type(),
                Json(json!({
                    "code": payload.get("code").and_then(|c| c.as_i64()).unwrap(),
                    "ingressCode": payload.get("ingressCode").and_then(|c| c.as_i64()).unwrap(),
                    "message": payload.get("message").and_then(|m| m.as_str()).unwrap()
                })),
            )
        }
        _ => panic!(),
    }
}

fn json_content_type() -> HeaderMap {
    let mut headers = HeaderMap::new();
    headers.insert("Content-Type", HeaderValue::from_static("application/json"));
    headers
}
