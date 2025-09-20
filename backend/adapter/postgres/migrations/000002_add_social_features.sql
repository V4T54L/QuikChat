-- +migrate Up
CREATE TYPE friendship_status AS ENUM ('pending', 'accepted');

CREATE TABLE friendships (
    user_id1 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_id2 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status friendship_status NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id1, user_id2),
    CHECK (user_id1 < user_id2) -- Ensures canonical ordering to prevent duplicates like (A,B) and (B,A)
);

CREATE INDEX idx_friendships_user_id1 ON friendships(user_id1);
CREATE INDEX idx_friendships_user_id2 ON friendships(user_id2);

CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    handle VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    photo_url VARCHAR(255),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL, -- Owner can leave, ownership transfers
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_groups_handle ON groups(LOWER(handle));

CREATE TABLE group_members (
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

CREATE INDEX idx_group_members_group_id ON group_members(group_id);
CREATE INDEX idx_group_members_user_id ON group_members(user_id);

-- +migrate Down
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS friendships;
DROP TYPE IF EXISTS friendship_status;

