-- +goose Up

CREATE TABLE feature_suggestions (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    text       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_feature_suggestions_org_id ON feature_suggestions(org_id);

-- +goose Down

DROP INDEX IF EXISTS idx_feature_suggestions_org_id;
DROP TABLE IF EXISTS feature_suggestions;
