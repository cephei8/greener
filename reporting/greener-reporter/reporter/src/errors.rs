use std::error::Error;
use std::fmt;

#[derive(Debug)]
pub enum ReporterError {
    InvalidArgument(String),
    Ingress(String, u16),
    Unknown(String),
}

impl fmt::Display for ReporterError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ReporterError::InvalidArgument(msg) => write!(f, "InvalidArgument error: {}", msg),
            ReporterError::Ingress(msg, code) => write!(f, "Ingress error: code={}, {}", code, msg),
            ReporterError::Unknown(msg) => write!(f, "Unknown error: {}", msg),
        }
    }
}

impl Error for ReporterError {}
