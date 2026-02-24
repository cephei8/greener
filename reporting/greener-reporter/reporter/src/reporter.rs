use crate::batcher::Batcher;
use crate::errors::ReporterError;
use crate::ingress::IngressClient;
use crate::models::{SessionRequest, TestcaseRequest};
use std::sync::Arc;
use tokio::runtime::Runtime;
use tokio::sync::Mutex;

#[repr(C)]
pub struct Reporter {
    runtime: Arc<Runtime>,
    ingress: Arc<IngressClient>,
    batcher: Arc<Mutex<Batcher>>,
}

impl Reporter {
    pub fn new(endpoint: String, api_key: String) -> Result<Self, ReporterError> {
        let runtime = Arc::new(
            Runtime::new()
                .map_err(|e| ReporterError::Unknown(format!("error creating runtime: {}", e)))?,
        );
        let ingress = Arc::new(IngressClient::new(endpoint, api_key).map_err(|e| {
            ReporterError::Unknown(format!("error creating ingress client: {}", e))
        })?);
        let batcher = Arc::new(Mutex::new(Batcher::new(runtime.clone(), ingress.clone())));

        Ok(Reporter {
            runtime,
            ingress,
            batcher,
        })
    }

    pub fn add_testcase(&self, testcase: TestcaseRequest) -> Result<(), ReporterError> {
        let batcher = self.runtime.block_on(self.batcher.lock());
        self.runtime.block_on(batcher.add(testcase))
    }

    pub fn shutdown(&self) -> Result<(), ReporterError> {
        let mut batcher = self.runtime.block_on(self.batcher.lock());
        self.runtime.block_on(batcher.shutdown())
    }

    pub fn pop_error(&self) -> Option<ReporterError> {
        let batcher = self.runtime.block_on(self.batcher.lock());
        self.runtime.block_on(batcher.pop_error())
    }

    pub fn create_session(&self, session: SessionRequest) -> Result<String, ReporterError> {
        self.runtime.block_on(self.ingress.create_session(session))
    }
}
