create table if not exists config (
    config_id text not null,
    active boolean not null,
    backyard_version text not null,
    title_home text not null,
    desc_home text not null,
    image_home text not null,
    favicon_home text not null,
    footer_html text not null,
    admin_user_id text not null,
    created_at datetime not null default current_timestamp,
    updated_at datetime not null default current_timestamp
);

create index config_id_idx on users_posts (user_id);

insert into config values (
    "config-f24099f5-a0b8-49bb-983e-43854436cd84",
    true,
    "0.0.1",
    "Backyard",
    "",
    "static/images/house.png",
    "static/images/cowboy.ico",
    "powered by backyard",
    "user_f24099f5-a0b8-49bb-983e-43854436cd84",
    current_timestamp,
    current_timestamp
);

insert into users values (
    "user-f24099f5-a0b8-49bb-983e-43854436cd84",
    "backyard",
    "admin@example.com",
    "user-f24099f5-a0b8-49bb-983e-43854436cd84",
    current_timestamp,
    current_timestamp
);
