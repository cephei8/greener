use crate::errors::ReporterError;
use crate::ingress::IngressClient;
use crate::models::TestcaseRequest;
use std::collections::VecDeque;
use std::sync::Arc;
use std::sync::atomic::Ordering;
use std::time::Duration;
use tokio::runtime::Runtime;
use tokio::sync::mpsc;
use tokio::time::{self, Instant};

enum BatcherMesssage {
    Testcase(TestcaseRequest),
    Shutdown,
}

pub struct Batcher {
    sender: mpsc::Sender<BatcherMesssage>,
    is_accepting: Arc<std::sync::atomic::AtomicBool>,
    worker_handle: Option<tokio::task::JoinHandle<()>>,
    errors: Arc<tokio::sync::Mutex<VecDeque<ReporterError>>>,
}

impl Batcher {
    pub fn new(runtime: Arc<Runtime>, ingress: Arc<IngressClient>) -> Self {
        let (sender, mut receiver) = mpsc::channel::<BatcherMesssage>(1000);
        let is_accepting = Arc::new(std::sync::atomic::AtomicBool::new(true));
        let errors = Arc::new(tokio::sync::Mutex::new(VecDeque::new()));

        let worker_handle = runtime.spawn({
            let errors = errors.clone();

            async move {
                let mut batch = Vec::<TestcaseRequest>::new();
                let mut last_send = Instant::now();
                let batch_timeout = Duration::from_secs(5);
                let max_batch_size = 100;

                loop {
                    tokio::select! {
                        Some(msg) = receiver.recv() => {
                            match msg {
                                BatcherMesssage::Testcase(testcase) => {
                                    batch.push(testcase);
                                    if batch.len() >= max_batch_size {
                                        if let Err(e) = ingress.create_testcases(std::mem::take(&mut batch)).await {
                                            let mut errors_guard = errors.lock().await;
                                            errors_guard.push_back(e);
                                        }
                                        last_send = Instant::now();
                                    }
                                }
                                BatcherMesssage::Shutdown => {
                                    if !batch.is_empty() {
                                        if let Err(e) = ingress.create_testcases(std::mem::take(&mut batch)).await {
                                            let mut errors_guard = errors.lock().await;
                                            errors_guard.push_back(e);
                                        }
                                    }
                                    break;
                                }
                            }
                        }
                        _ = time::sleep_until(last_send + batch_timeout) => {
                            if !batch.is_empty() {
                                if let Err(e) = ingress.create_testcases(std::mem::take(&mut batch)).await {
                                    let mut errors_guard = errors.lock().await;
                                    errors_guard.push_back(e);
                                }
                                last_send = Instant::now();
                            }
                        }
                    }
                }
            }
        });

        Self {
            sender,
            is_accepting,
            worker_handle: Some(worker_handle),
            errors,
        }
    }

    pub async fn add(&self, testcase: TestcaseRequest) -> Result<(), ReporterError> {
        if self.is_accepting.load(Ordering::SeqCst) {
            self.sender
                .send(BatcherMesssage::Testcase(testcase))
                .await
                .map_err(|e| {
                    ReporterError::Unknown(format!(
                        "error sending testcase to batcher queue: {}",
                        e
                    ))
                })?;
        }
        Ok(())
    }

    pub async fn shutdown(&mut self) -> Result<(), ReporterError> {
        self.is_accepting.store(false, Ordering::SeqCst);
        self.sender
            .send(BatcherMesssage::Shutdown)
            .await
            .map_err(|e| {
                ReporterError::Unknown(format!("error sending shutdown to batcher queue: {}", e))
            })?;
        if let Some(worker_handle) = self.worker_handle.take() {
            match worker_handle.await {
                Ok(_) => Ok(()),
                Err(e) => Err(ReporterError::Unknown(format!(
                    "error joining batcher worker thread: {}",
                    e
                ))),
            }
        } else {
            Ok(())
        }
    }

    pub async fn pop_error(&self) -> Option<ReporterError> {
        let mut errors_guard = self.errors.lock().await;
        errors_guard.pop_front()
    }
}
