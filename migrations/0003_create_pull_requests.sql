-- +goose Up
CREATE TABLE pr_statuses (
    id SERIAL PRIMARY KEY,
    code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

INSERT INTO pr_statuses (code, name, description) VALUES
    ('OPEN', 'Open', 'Pull Request is open for review'),
    ('MERGED', 'Merged', 'Pull Request has been merged');

CREATE INDEX idx_pr_statuses_code ON pr_statuses(code);

CREATE TABLE pull_requests (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    author_id VARCHAR(50) REFERENCES users(id) ON DELETE RESTRICT,
    status_id INTEGER REFERENCES pr_statuses(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    merged_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE TABLE pr_reviewers (
    pr_id VARCHAR(50) REFERENCES pull_requests(id) ON DELETE CASCADE,
    user_id VARCHAR(50) REFERENCES users(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (pr_id, user_id)
);

CREATE INDEX idx_pull_requests_author_id ON pull_requests(author_id);
CREATE INDEX idx_pull_requests_status_id ON pull_requests(status_id);
CREATE INDEX idx_pr_reviewers_user_id ON pr_reviewers(user_id);
CREATE INDEX idx_pull_requests_merged_at ON pull_requests(merged_at) WHERE merged_at IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS pr_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS pr_statuses CASCADE;
