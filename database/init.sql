-- wjecoffeetaste 咖啡品鉴社区 初始化脚本（PostgreSQL）
-- 由 postgres 官方镜像 /docker-entrypoint-initdb.d 首次启动时自动执行。

CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  username VARCHAR(64) NOT NULL,
  email VARCHAR(128) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  avatar VARCHAR(255),
  bio VARCHAR(512),
  role VARCHAR(16) NOT NULL DEFAULT 'user',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  CONSTRAINT uni_users_username UNIQUE (username),
  CONSTRAINT uni_users_email UNIQUE (email)
);

CREATE TABLE IF NOT EXISTS tasting_notes (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  coffee_name VARCHAR(128) NOT NULL,
  origin VARCHAR(128),
  roast_level VARCHAR(16) NOT NULL,
  flavor_tags JSONB DEFAULT '[]',
  aroma_score DOUBLE PRECISION DEFAULT 0,
  acidity_score DOUBLE PRECISION DEFAULT 0,
  body_score DOUBLE PRECISION DEFAULT 0,
  overall_score DOUBLE PRECISION DEFAULT 0,
  brew_method VARCHAR(64),
  brew_recipe_id BIGINT DEFAULT 0,
  notes_text TEXT,
  image_url VARCHAR(255),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  CONSTRAINT fk_note_user FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS brew_recipes (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  name VARCHAR(128) NOT NULL,
  device VARCHAR(64),
  water_temp INT DEFAULT 0,
  grind_size VARCHAR(32),
  ratio VARCHAR(32),
  steps JSONB DEFAULT '[]',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  CONSTRAINT fk_recipe_user FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS coffee_beans (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  origin VARCHAR(128),
  process_method VARCHAR(16),
  flavor_tags JSONB DEFAULT '[]',
  description TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  CONSTRAINT uni_coffee_beans_name UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS comments (
  id BIGSERIAL PRIMARY KEY,
  note_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  content VARCHAR(1000) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  CONSTRAINT fk_comment_note FOREIGN KEY (note_id) REFERENCES tasting_notes(id),
  CONSTRAINT fk_comment_user FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS likes (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  note_id BIGINT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  CONSTRAINT idx_like_user_note UNIQUE (user_id, note_id)
);

CREATE TABLE IF NOT EXISTS user_follows (
  follower_id BIGINT NOT NULL,
  following_id BIGINT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  PRIMARY KEY (follower_id, following_id)
);

-- 种子数据
INSERT INTO users (username, email, password_hash, bio, role) VALUES
  ('admin', 'admin@coffeetaste.local', '$2a$10$VGETME6mK/u27yF1UwKHkuh0b36LjEpJjw2c4J2L7wPph1pcG0cVO', '咖啡平台管理员', 'admin'),
  ('barista', 'barista@coffeetaste.local', '$2a$10$txSqFgLTRQZHGsde2i9vPuyh0WeH3adS0BTHSc..i8Y8FbtF4/rri', '精品咖啡爱好者', 'user'),
  ('roaster', 'roaster@coffeetaste.local', '$2a$10$txSqFgLTRQZHGsde2i9vPuyh0WeH3adS0BTHSc..i8Y8FbtF4/rri', '烘焙师', 'user');

INSERT INTO coffee_beans (name, origin, process_method, flavor_tags, description) VALUES
  ('埃塞俄比亚耶加雪菲', '埃塞俄比亚', 'washed', '["柑橘","茉莉","蜂蜜"]', '经典水洗耶加雪菲，明亮柑橘酸质。'),
  ('哥伦比亚慧兰', '哥伦比亚', 'washed', '["坚果","焦糖","红苹果"]', '平衡甜感，坚果与焦糖尾韵。'),
  ('哥斯达黎加蜜处理', '哥斯达黎加', 'honey', '["莓果","红糖","葡萄干"]', '蜜处理带来醇厚甜感与莓果香气。'),
  ('印尼曼特宁', '印度尼西亚', 'washed', '["草本","黑巧克力","香料"]', '醇厚浓郁，草本与黑巧风味。');

INSERT INTO brew_recipes (user_id, name, device, water_temp, grind_size, ratio, steps) VALUES
  (2, '手冲三段式', '手冲壶', 92, '中细', '1:15', '[{"step_number":1,"description":"闷蒸30秒","duration_seconds":30},{"step_number":2,"description":"第一段注水至150ml","duration_seconds":20},{"step_number":3,"description":"第二段注水至300ml","duration_seconds":30}]'),
  (3, '法压壶经典', '法压壶', 94, '中粗', '1:14', '[{"step_number":1,"description":"注水并搅拌","duration_seconds":10},{"step_number":2,"description":"浸泡4分钟","duration_seconds":240},{"step_number":3,"description":"缓慢压杆","duration_seconds":15}]');

INSERT INTO tasting_notes (user_id, coffee_name, origin, roast_level, flavor_tags, aroma_score, acidity_score, body_score, overall_score, brew_method, brew_recipe_id, notes_text, image_url) VALUES
  (2, '埃塞俄比亚耶加雪菲', '埃塞俄比亚', 'light', '["柑橘","茉莉"]', 8.5, 8.0, 7.0, 8.3, '手冲', 1, '花香明显，柑橘酸质明亮，回甘持久。', 'https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?w=600'),
  (3, '哥伦比亚慧兰', '哥伦比亚', 'medium', '["坚果","焦糖"]', 7.5, 6.8, 7.8, 7.6, '法压', 2, '甜感平衡，坚果香气浓郁。', 'https://images.unsplash.com/photo-1447933601403-0c6688de566e?w=600'),
  (2, '哥斯达黎加蜜处理', '哥斯达黎加', 'medium', '["莓果","红糖"]', 8.0, 7.2, 8.0, 7.9, '手冲', 0, '莓果酸甜与红糖甜感交织。', 'https://images.unsplash.com/photo-1517701604599-bb29b565090c?w=600');

INSERT INTO comments (note_id, user_id, content) VALUES
  (1, 3, '我也很喜欢这只耶加雪菲，柑橘调太棒了！'),
  (2, 2, '慧兰做奶咖也很合适。');

INSERT INTO likes (user_id, note_id) VALUES
  (3, 1),
  (2, 2),
  (1, 1);

INSERT INTO user_follows (follower_id, following_id) VALUES
  (2, 3),
  (3, 2);
