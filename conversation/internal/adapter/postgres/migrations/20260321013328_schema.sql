-- +goose Up
CREATE TABLE users (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ
);

CREATE TABLE conversations (
	id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	title      VARCHAR(16) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE messages (
	id                BYTEA NOT NULL,
	conversation_id   UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
	parent_message_id BYTEA,
	role              SMALLINT NOT NULL,
	text              TEXT NOT NULL,
	created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

	PRIMARY KEY (id),
	
	CONSTRAINT messages_id_conv_unique UNIQUE (id, conversation_id),
	
	CONSTRAINT messages_parent_fkey 
		FOREIGN KEY (parent_message_id, conversation_id) 
		REFERENCES messages(id, conversation_id) 
		ON DELETE CASCADE,

	CONSTRAINT messages_no_self_ref CHECK (id <> parent_message_id)
);

CREATE UNIQUE INDEX idx_one_root_per_conversation 
ON messages (conversation_id) 
WHERE parent_message_id IS NULL;

CREATE TABLE fork (
	parent_message_id         BYTEA NOT NULL,
	selected_child_message_id BYTEA NOT NULL,
	conversation_id           UUID NOT NULL,

	FOREIGN KEY (parent_message_id, conversation_id) 
		REFERENCES messages(id, conversation_id) ON DELETE CASCADE,
	FOREIGN KEY (selected_child_message_id, conversation_id) 
		REFERENCES messages(id, conversation_id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE fork;
DROP TABLE messages;
DROP TABLE conversations;
DROP TABLE users;
