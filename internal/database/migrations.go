package database

import (
	"database/sql"
	"log"
)

func RunMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR(50) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			display_name VARCHAR(100),
			avatar_url TEXT,
			bio TEXT,
			status VARCHAR(100),
			role VARCHAR(20) DEFAULT 'user',
			is_psychologist BOOLEAN DEFAULT FALSE,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS topics (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title VARCHAR(255) NOT NULL,
			description TEXT,
			is_public BOOLEAN DEFAULT TRUE,
			created_by UUID REFERENCES users(id) ON DELETE CASCADE,
			votes_count INT DEFAULT 0,
			messages_count INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS topic_votes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			topic_id UUID REFERENCES topics(id) ON DELETE CASCADE,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			vote_type VARCHAR(10) DEFAULT 'up',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(topic_id, user_id)
		)`,

		`CREATE TABLE IF NOT EXISTS groups (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) NOT NULL,
			description TEXT,
			avatar_url TEXT,
			is_private BOOLEAN DEFAULT FALSE,
			created_by UUID REFERENCES users(id) ON DELETE CASCADE,
			members_count INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS group_members (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			group_id UUID REFERENCES groups(id) ON DELETE CASCADE,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			role VARCHAR(20) DEFAULT 'member',
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(group_id, user_id)
		)`,

		`CREATE TABLE IF NOT EXISTS messages (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			content TEXT NOT NULL,
			topic_id UUID REFERENCES topics(id) ON DELETE CASCADE,
			group_id UUID REFERENCES groups(id) ON DELETE CASCADE,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			parent_id UUID REFERENCES messages(id) ON DELETE CASCADE,
			quoted_message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
			is_edited BOOLEAN DEFAULT FALSE,
			is_deleted BOOLEAN DEFAULT FALSE,
			edited_at TIMESTAMP,
			deleted_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS reactions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			message_id UUID REFERENCES messages(id) ON DELETE CASCADE,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			emoji VARCHAR(10) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(message_id, user_id, emoji)
		)`,

		`CREATE TABLE IF NOT EXISTS sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title VARCHAR(255) NOT NULL,
			description TEXT,
			session_type VARCHAR(20) DEFAULT 'webinar',
			hms_room_id VARCHAR(255),
			hms_room_code VARCHAR(255),
			psychologist_id UUID REFERENCES users(id) ON DELETE CASCADE,
			max_participants INT DEFAULT 50,
			scheduled_at TIMESTAMP NOT NULL,
			duration_minutes INT DEFAULT 60,
			is_private BOOLEAN DEFAULT FALSE,
			status VARCHAR(20) DEFAULT 'scheduled',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS session_participants (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			session_id UUID REFERENCES sessions(id) ON DELETE CASCADE,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			left_at TIMESTAMP,
			UNIQUE(session_id, user_id)
		)`,

		`CREATE TABLE IF NOT EXISTS appointments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			psychologist_id UUID REFERENCES users(id) ON DELETE CASCADE,
			client_id UUID REFERENCES users(id) ON DELETE CASCADE,
			title VARCHAR(255),
			description TEXT,
			scheduled_at TIMESTAMP NOT NULL,
			duration_minutes INT DEFAULT 60,
			status VARCHAR(20) DEFAULT 'pending',
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS notifications (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			type VARCHAR(50) NOT NULL,
			title VARCHAR(255) NOT NULL,
			content TEXT,
			link TEXT,
			is_read BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS user_status (
			user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			is_online BOOLEAN DEFAULT FALSE,
			last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS user_blocks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			blocked_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, blocked_user_id)
		)`,

		`CREATE TABLE IF NOT EXISTS conversations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user1_id UUID REFERENCES users(id) ON DELETE CASCADE,
			user2_id UUID REFERENCES users(id) ON DELETE CASCADE,
			last_message_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user1_id, user2_id)
		)`,

		`CREATE TABLE IF NOT EXISTS direct_messages (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			conversation_id UUID REFERENCES conversations(id) ON DELETE CASCADE,
			sender_id UUID REFERENCES users(id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			is_read BOOLEAN DEFAULT FALSE,
			is_edited BOOLEAN DEFAULT FALSE,
			is_deleted BOOLEAN DEFAULT FALSE,
			edited_at TIMESTAMP,
			deleted_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS message_read_receipts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			message_id UUID REFERENCES messages(id) ON DELETE CASCADE,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			read_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(message_id, user_id)
		)`,

		`CREATE TABLE IF NOT EXISTS typing_indicators (
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			room_id VARCHAR(255) NOT NULL,
			started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY(user_id, room_id)
		)`,

		`CREATE INDEX IF NOT EXISTS idx_messages_topic ON messages(topic_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_group ON messages(group_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_user ON messages(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_parent ON messages(parent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_reactions_message ON reactions(message_id)`,
		`CREATE INDEX IF NOT EXISTS idx_topic_votes_topic ON topic_votes(topic_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_psychologist ON sessions(psychologist_id)`,
		`CREATE INDEX IF NOT EXISTS idx_appointments_psychologist ON appointments(psychologist_id)`,
		`CREATE INDEX IF NOT EXISTS idx_appointments_client ON appointments(client_id)`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_blocks_user ON user_blocks(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_conversations_users ON conversations(user1_id, user2_id)`,
		`CREATE INDEX IF NOT EXISTS idx_direct_messages_conversation ON direct_messages(conversation_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_status_online ON user_status(is_online)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return err
		}
	}

	log.Println("Migrations completed successfully")
	return nil
}
