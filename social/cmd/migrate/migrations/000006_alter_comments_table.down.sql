ALTER TABLE comments
    DROP CONSTRAINT IF EXISTS fk_comments_user_id;

ALTER TABLE comments
    DROP CONSTRAINT IF EXISTS fk_comments_post_id;