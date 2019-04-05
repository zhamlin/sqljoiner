create table users  (
    id uuid primary key,
    age int,
    name text
);

create table user_addresses (
    user_id uuid references users (id),
    address_id uuid references addresses (id)
);
