-- +goose Up
CREATE TABLE IF NOT EXISTS users (
                                     id BIGINT AUTO_INCREMENT PRIMARY KEY,
                                     name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    hash_pass VARCHAR(255) NOT NULL,
    create_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE IF NOT EXISTS teams (
                                     id BIGINT AUTO_INCREMENT PRIMARY KEY,
                                     name VARCHAR(100) NOT NULL,
    owner_id BIGINT NOT NULL,
    create_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
    );

CREATE TABLE IF NOT EXISTS team_members (
                                            user_id BIGINT NOT NULL,
                                            team_id BIGINT NOT NULL,
                                            role VARCHAR(20) DEFAULT 'member',
    PRIMARY KEY (user_id, team_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
    );

CREATE TABLE IF NOT EXISTS tasks (
                                     id BIGINT AUTO_INCREMENT PRIMARY KEY,
                                     team_id BIGINT NOT NULL,
                                     owner_id BIGINT NOT NULL,
                                     title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'todo',
    assignee_id BIGINT,
    create_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    complete_at TIMESTAMP NULL DEFAULT NULL,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (assignee_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
    );

CREATE TABLE IF NOT EXISTS task_history (
                                            id BIGINT AUTO_INCREMENT PRIMARY KEY,
                                            task_id BIGINT NOT NULL,
                                            change_by BIGINT NOT NULL,
                                            field VARCHAR(50) NOT NULL,
    old_value TEXT,
    new_value TEXT,
    change_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (change_by) REFERENCES users(id) ON DELETE CASCADE
    );

CREATE TABLE IF NOT EXISTS task_comments (
                                             id BIGINT AUTO_INCREMENT PRIMARY KEY,
                                             task_id BIGINT NOT NULL,
                                             user_id BIGINT NOT NULL,
                                             content TEXT NOT NULL,
                                             create_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                             FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    );

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_team_members_team_id ON team_members(team_id);
CREATE INDEX idx_team_members_user_id ON team_members(user_id);
CREATE INDEX idx_tasks_team_id ON tasks(team_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_assignee_id ON tasks(assignee_id);
CREATE INDEX idx_tasks_owner_id ON tasks(owner_id);
CREATE INDEX idx_task_history_task_id ON task_history(task_id);
CREATE INDEX idx_task_history_change_by ON task_history(change_by);
CREATE INDEX idx_task_comments_task_id ON task_comments(task_id);
CREATE INDEX idx_task_comments_user_id ON task_comments(user_id);


-- +goose Down

DROP TABLE IF EXISTS task_comments;
DROP TABLE IF EXISTS task_history;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS users;

