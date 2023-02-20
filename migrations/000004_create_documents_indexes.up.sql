CREATE INDEX IF NOT EXISTS documents_title_index ON documents USING GIN (to_tsvector('simple', title));
CREATE INDEX IF NOT EXISTS documents_tags_index ON documents USING GIN (tags);