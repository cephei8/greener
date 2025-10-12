mod user;

use clap::{Parser, Subcommand};
use pbkdf2::pbkdf2_hmac;
use rand::RngCore;
use sea_orm::{ActiveModelTrait, ConnectionTrait, Database, DatabaseConnection, DbBackend, Set};
use sha2::Sha256;
use std::error::Error;
use uuid::Uuid;

#[derive(Parser)]
#[command(name = "greener-admin-cli")]
#[command(about = "Admin utility for Greener", long_about = None)]
struct Cli {
    #[arg(long, help = "Database connection URL")]
    db_url: String,

    #[arg(long, help = "Database type (postgres, mysql, sqlite)")]
    db_type: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    #[command(about = "Create a new user in the users table")]
    CreateUser {
        #[arg(long, help = "User name")]
        username: String,

        #[arg(long, help = "User password")]
        password: String,
    },
}

async fn init_db(url: &str, db_type: &str) -> Result<DatabaseConnection, Box<dyn Error>> {
    let backend = match db_type.to_lowercase().as_str() {
        "postgres" | "postgresql" => DbBackend::Postgres,
        "mysql" => DbBackend::MySql,
        "sqlite" | "sqlite3" => DbBackend::Sqlite,
        _ => {
            return Err(format!(
                "Unsupported database type: {} (supported: postgres, mysql, sqlite)",
                db_type
            ).into());
        }
    };

    let db = Database::connect(url)
        .await
        .map_err(|e| format!("Failed to connect to {} database: {}", db_type, e))?;

    if db.get_database_backend() != backend {
        return Err(format!(
            "Database backend mismatch: expected {:?}, got {:?}",
            backend,
            db.get_database_backend()
        ).into());
    }

    db.ping().await.map_err(|e| format!("Failed to ping database: {}", e))?;

    Ok(db)
}

fn hash_password(password: &str) -> Result<(Vec<u8>, Vec<u8>), Box<dyn Error>> {
    let mut salt = vec![0u8; 32];
    rand::thread_rng().fill_bytes(&mut salt);

    let mut hash = vec![0u8; 32];
    pbkdf2_hmac::<Sha256>(password.as_bytes(), &salt, 100_000, &mut hash);

    Ok((salt, hash))
}

async fn create_user_action(
    db_url: &str,
    db_type: &str,
    username: &str,
    password: &str,
) -> Result<(), Box<dyn Error>> {
    let db = init_db(db_url, db_type).await?;

    let (password_salt, password_hash) = hash_password(password)?;

    let now = chrono::Utc::now();

    let user = user::ActiveModel {
        id: Set(Uuid::new_v4()),
        username: Set(username.to_string()),
        password_salt: Set(password_salt),
        password_hash: Set(password_hash),
        created_at: Set(now.into()),
        updated_at: Set(now.into()),
    };

    user.insert(&db)
        .await
        .map_err(|e| format!("Failed to create user: {}", e))?;

    println!("User created successfully: {}", username);
    Ok(())
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let cli = Cli::parse();

    match &cli.command {
        Commands::CreateUser { username, password } => {
            create_user_action(&cli.db_url, &cli.db_type, username, password).await?;
        }
    }

    Ok(())
}
