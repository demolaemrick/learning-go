ALTER TABLE comments
    ADD CONSTRAINT fk_comments_user_id
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE comments
    ADD CONSTRAINT fk_comments_post_id
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE;