begin;

create table
    if not exists matches (
        id int primary key generated always as identity,
        issuer_user_id int not null,
        receiver_user_id int not null,
        issuer_cat_id int not null,
        receiver_cat_id int not null,
        created_at timestamptz not null default now ()
    );

create index if not exists search_matches_1 on matches (issuer_user_id);

create index if not exists search_matches_2 on matches (issuer_user_id);

create index if not exists search_matches_3 on matches (issuer_cat_id);

create index if not exists search_matches_4 on matches (receiver_cat_id);

commit;