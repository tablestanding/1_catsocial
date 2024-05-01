create table
    if not exists users (
        id int primary key generated always as identity,
        email text unique not null,
        hashed_pw text not null,
        name text not null,
        created_at timestamptz not null default now ()
    );