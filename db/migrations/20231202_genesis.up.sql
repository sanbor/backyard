create table if not exists posts (
    post_id text primary key,
    title text,
    content text,
    draft boolean not null,
    created_at datetime not null,
    updated_at datetime not null
);

create table if not exists users (
    user_id text primary key,
    username text not null,
    email text,
    password text not null,
    created_at datetime not null,
    updated_at datetime not null
);

create unique index usersname_unique_idx on users(username);

