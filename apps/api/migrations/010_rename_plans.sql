-- +goose Up

ALTER TYPE plan RENAME VALUE 'hobbyist' TO 'hobby';
ALTER TYPE plan RENAME VALUE 'indie' TO 'solo';
ALTER TYPE plan RENAME VALUE 'studio' TO 'startup';
ALTER TYPE plan RENAME VALUE 'agency' TO 'enterprise';

-- +goose Down

ALTER TYPE plan RENAME VALUE 'hobby' TO 'hobbyist';
ALTER TYPE plan RENAME VALUE 'solo' TO 'indie';
ALTER TYPE plan RENAME VALUE 'startup' TO 'studio';
ALTER TYPE plan RENAME VALUE 'enterprise' TO 'agency';
