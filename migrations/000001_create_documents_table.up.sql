CREATE TABLE IF NOT EXISTS documents (
	document_id serial PRIMARY KEY,
    user_id integer DEFAULT 0,
    url_s3 text,
    filetype text,
    created_at timestamp(0) with time zone DEFAULT NOW(),
	title text,
	tags text[],
    is_private boolean
);