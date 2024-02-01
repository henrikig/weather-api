create table if not exists cities (
  id serial primary key,
  name text not null,
  lat numeric not null,
  long numeric not null
);

create index if not exists cities_name_idx on cities (name);
