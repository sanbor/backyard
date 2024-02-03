create table if not exists users_posts (
    user_id text not null,
    post_id text not null,
    relation_type text check( relation_type in ('AUTHOR','EDIT','VIEW') ) not null default 'AUTHOR',
    created_at datetime not null default current_timestamp,
    updated_at datetime not null default current_timestamp,
    constraint users_posts_post_id_FK foreign key (post_id) references posts(post_id) on delete cascade,
	constraint users_posts_user_id_FK foreign key (user_id) references users(user_id) on delete cascade
);

create index users_posts_user_id_idx on users_posts (user_id);
create index users_posts_post_id_idx on users_posts (post_id);
create index users_posts_post_id_user_id_idx on users_posts (post_id,user_id);
