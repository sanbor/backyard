CREATE TABLE IF NOT EXISTS users_posts (
    user_id TEXT,
    post_id TEXT,
    relation_type TEXT,
    createdAt DATETIME,
    updatedAt DATETIME,
    CONSTRAINT users_posts_post_id_FK FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
	CONSTRAINT users_posts_user_id_FK FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX users_posts_user_id_IDX ON user_post (user_id);
CREATE INDEX users_posts_post_id_IDX ON user_post (post_id);
CREATE INDEX users_posts_post_id_user_id_IDX ON user_post (post_id,user_id);
