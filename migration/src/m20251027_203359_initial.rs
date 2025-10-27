use sea_orm_migration::{prelude::*, schema::*};

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .create_table(
                Table::create()
                    .table(Users::Table)
                    .col(uuid(Users::Id).primary_key())
                    .col(string(Users::Username))
                    .col(binary(Users::PasswordSalt))
                    .col(binary(Users::PasswordHash))
                    .col(
                        timestamp_with_time_zone(Users::CreatedAt)
                            .default(Expr::current_timestamp()),
                    )
                    .col(
                        timestamp_with_time_zone(Users::UpdatedAt)
                            .default(Expr::current_timestamp()),
                    )
                    .to_owned(),
            )
            .await?;

        manager
            .create_index(
                Index::create()
                    .name("ix_users_username")
                    .table(Users::Table)
                    .col(Users::Username)
                    .to_owned(),
            )
            .await?;

        manager
            .create_table(
                Table::create()
                    .table(Apikeys::Table)
                    .col(uuid(Apikeys::Id).primary_key())
                    .col(string_null(Apikeys::Description))
                    .col(binary(Apikeys::SecretSalt))
                    .col(binary(Apikeys::SecretHash))
                    .col(
                        timestamp_with_time_zone(Apikeys::CreatedAt)
                            .default(Expr::current_timestamp()),
                    )
                    .col(
                        timestamp_with_time_zone(Apikeys::UpdatedAt)
                            .default(Expr::current_timestamp()),
                    )
                    .col(uuid(Apikeys::UserId))
                    .foreign_key(
                        ForeignKey::create()
                            .from(Apikeys::Table, Apikeys::UserId)
                            .to(Users::Table, Users::Id)
                            .on_delete(ForeignKeyAction::Cascade),
                    )
                    .to_owned(),
            )
            .await?;

        manager
            .create_index(
                Index::create()
                    .name("ix_apikeys_user_id")
                    .table(Apikeys::Table)
                    .col(Apikeys::UserId)
                    .to_owned(),
            )
            .await?;

        manager
            .create_table(
                Table::create()
                    .table(Sessions::Table)
                    .col(uuid(Sessions::Id).primary_key())
                    .col(string_null(Sessions::Description))
                    .col(json_null(Sessions::Baggage))
                    .col(
                        timestamp_with_time_zone(Sessions::CreatedAt)
                            .default(Expr::current_timestamp()),
                    )
                    .col(
                        timestamp_with_time_zone(Sessions::UpdatedAt)
                            .default(Expr::current_timestamp()),
                    )
                    .col(uuid(Sessions::UserId))
                    .foreign_key(
                        ForeignKey::create()
                            .from(Sessions::Table, Sessions::UserId)
                            .to(Users::Table, Users::Id)
                            .on_delete(ForeignKeyAction::Cascade),
                    )
                    .to_owned(),
            )
            .await?;

        manager
            .create_index(
                Index::create()
                    .name("ix_sessions_user_id")
                    .table(Sessions::Table)
                    .col(Sessions::UserId)
                    .to_owned(),
            )
            .await?;

        manager
            .create_table(
                Table::create()
                    .table(Labels::Table)
                    .col(pk_auto(Labels::Id))
                    .col(uuid(Labels::SessionId))
                    .col(string(Labels::Key))
                    .col(string_null(Labels::Value))
                    .col(uuid(Labels::UserId))
                    .col(
                        timestamp_with_time_zone(Labels::CreatedAt)
                            .default(Expr::current_timestamp()),
                    )
                    .col(
                        timestamp_with_time_zone(Labels::UpdatedAt)
                            .default(Expr::current_timestamp()),
                    )
                    .foreign_key(
                        ForeignKey::create()
                            .from(Labels::Table, Labels::SessionId)
                            .to(Sessions::Table, Sessions::Id)
                            .on_delete(ForeignKeyAction::Cascade),
                    )
                    .foreign_key(
                        ForeignKey::create()
                            .from(Labels::Table, Labels::UserId)
                            .to(Users::Table, Users::Id)
                            .on_delete(ForeignKeyAction::Cascade),
                    )
                    .to_owned(),
            )
            .await?;

        manager
            .create_index(
                Index::create()
                    .name("ix_labels_session_id")
                    .table(Labels::Table)
                    .col(Labels::SessionId)
                    .to_owned(),
            )
            .await?;

        manager
            .create_index(
                Index::create()
                    .name("ix_labels_user_id")
                    .table(Labels::Table)
                    .col(Labels::UserId)
                    .to_owned(),
            )
            .await?;

        manager
            .create_table(
                Table::create()
                    .table(Testcases::Table)
                    .col(uuid(Testcases::Id).primary_key())
                    .col(uuid(Testcases::SessionId))
                    .col(string(Testcases::Name))
                    .col(string_null(Testcases::Classname))
                    .col(string_null(Testcases::File))
                    .col(string_null(Testcases::Testsuite))
                    .col(string_null(Testcases::Output))
                    .col(integer(Testcases::Status))
                    .col(json_null(Testcases::Baggage))
                    .col(
                        timestamp_with_time_zone(Testcases::CreatedAt)
                            .default(Expr::current_timestamp()),
                    )
                    .col(
                        timestamp_with_time_zone(Testcases::UpdatedAt)
                            .default(Expr::current_timestamp()),
                    )
                    .col(uuid(Testcases::UserId))
                    .foreign_key(
                        ForeignKey::create()
                            .from(Testcases::Table, Testcases::SessionId)
                            .to(Sessions::Table, Sessions::Id)
                            .on_delete(ForeignKeyAction::Cascade),
                    )
                    .foreign_key(
                        ForeignKey::create()
                            .from(Testcases::Table, Testcases::UserId)
                            .to(Users::Table, Users::Id)
                            .on_delete(ForeignKeyAction::Cascade),
                    )
                    .to_owned(),
            )
            .await?;

        manager
            .create_index(
                Index::create()
                    .name("ix_testcases_session_id")
                    .table(Testcases::Table)
                    .col(Testcases::SessionId)
                    .to_owned(),
            )
            .await?;

        manager
            .create_index(
                Index::create()
                    .name("ix_testcases_user_id")
                    .table(Testcases::Table)
                    .col(Testcases::UserId)
                    .to_owned(),
            )
            .await?;

        Ok(())
    }

    async fn down(&self, _manager: &SchemaManager) -> Result<(), DbErr> {
        Ok(())
    }
}

#[derive(Iden)]
enum Users {
    Table,
    Id,
    Username,
    PasswordSalt,
    PasswordHash,
    CreatedAt,
    UpdatedAt,
}

#[derive(Iden)]
enum Apikeys {
    Table,
    Id,
    Description,
    SecretSalt,
    SecretHash,
    UserId,
    CreatedAt,
    UpdatedAt,
}

#[derive(Iden)]
enum Sessions {
    Table,
    Id,
    Description,
    Baggage,
    CreatedAt,
    UpdatedAt,
    UserId,
}

#[derive(Iden)]
enum Labels {
    Table,
    Id,
    SessionId,
    Key,
    Value,
    UserId,
    CreatedAt,
    UpdatedAt,
}

#[derive(Iden)]
enum Testcases {
    Table,
    Id,
    SessionId,
    Name,
    Classname,
    File,
    Testsuite,
    Output,
    Status,
    Baggage,
    CreatedAt,
    UpdatedAt,
    UserId,
}
