begin;

create table
    if not exists matches (
        id int primary key generated always as identity,
        issuer_user_id int not null,
        receiver_user_id int not null,
        issuer_cat_id int not null,
        receiver_cat_id int not null,
        has_been_approved_or_rejected boolean not null default false,
        msg text not null,
        created_at timestamptz not null default now ()
    );

create index if not exists idx_matches_issuer_user_id on matches (issuer_user_id);

create index if not exists idx_matches_issuer_user_id on matches (issuer_user_id);

create index if not exists idx_matches_issuer_cat_id on matches (issuer_cat_id);

create index if not exists idx_matches_receiver_cat_id on matches (receiver_cat_id);

commit;