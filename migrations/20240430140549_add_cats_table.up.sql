begin;

create table
    if not exists cats (
        id int primary key generated always as identity,
        user_id int not null,
        race text not null,
        sex text not null,
        name text not null,
        name_normalized text not null,
        age_in_month int not null,
        description text not null,
        image_urls text[] not null,
        has_matched boolean not null default false,
        created_at timestamptz not null default now()
    );

create index if not exists idx_cats_user_id on cats (user_id);

create index if not exists idx_cats_has_matched on cats (has_matched);

create index if not exists idx_cats_sex on cats (sex);

create index if not exists idx_cats_race on cats (race);

create index if not exists idx_cats_age_in_month on cats (age_in_month);

create extension if not exists pg_trgm;

create index if not exists idx_cats_name_normalized on cats using gin(name_normalized gin_trgm_ops);

commit;