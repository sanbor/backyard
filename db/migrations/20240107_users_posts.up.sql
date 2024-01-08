CREATE TABLE IF NOT EXISTS users_posts (
    user_id TEXT NOT NULL,
    post_id TEXT NOT NULL,
    relation_type TEXT CHECK( relation_type IN ('AUTHOR','EDIT','VIEW') ) NOT NULL DEFAULT 'AUTHOR',
    createdAt DATETIME NOT NULL,
    updatedAt DATETIME NOT NULL,
    CONSTRAINT users_posts_post_id_FK FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
	CONSTRAINT users_posts_user_id_FK FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX users_posts_user_id_IDX ON users_posts (user_id);
CREATE INDEX users_posts_post_id_IDX ON users_posts (post_id);
CREATE INDEX users_posts_post_id_user_id_IDX ON users_posts (post_id,user_id);
