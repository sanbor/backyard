CREATE TABLE IF NOT EXISTS user_post (
    user_id TEXT,
    post_id TEXT,
    relation_type TEXT,
    createdAt DATETIME,
    updatedAt DATETIME,
    CONSTRAINT user_post_posts_FK FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
	CONSTRAINT user_post_users_FK FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX user_post_user_id_IDX ON user_post (user_id);
CREATE INDEX user_post_post_id_IDX ON user_post (post_id);
CREATE INDEX user_post_post_id_user_id_IDX ON user_post (post_id,user_id);
