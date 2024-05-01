begin;

drop index if exists idx_cats_user_id;

drop index if exists idx_cats_has_matched;

drop index if exists idx_cats_sex;

drop index if exists idx_cats_race;

drop index if exists idx_cats_age_in_month;

drop index if exists idx_cats_name_normalized;

drop table if exists cats;

commit;