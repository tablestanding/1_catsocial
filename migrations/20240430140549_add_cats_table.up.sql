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

create index if not exists search_cats_1 on cats (user_id);

create index if not exists search_cats_2 on cats (has_matched);

create index if not exists search_cats_3 on cats (sex);

create index if not exists search_cats_4 on cats (race);

create index if not exists search_cats_5 on cats (age_in_month);

create extension if not exists pg_trgm;

create index if not exists search_cats_trgm_1 on cats using gin(name_normalized gin_trgm_ops);

commit;