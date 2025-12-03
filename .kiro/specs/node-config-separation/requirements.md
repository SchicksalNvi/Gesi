# Requirements Document

## Introduction

Go-CESI 当前将所有节点配置存储在主配置文件 `config.toml` 中。随着节点数量增长到上百个，这种方式会导致配置文件臃肿、难以维护。本需求旨在将节点配置分离到独立的 `nodelist.toml` 文件中，同时保持向后兼容性，确保现有用户的配置不受影响。

## Glossary

- **System**: Go-CESI 应用程序
- **Node Configuration**: Supervisor 节点的连接信息（名称、环境、主机、端口、认证信息）
- **Main Configuration**: 系统级配置（服务器端口、数据库路径、管理员账户、性能参数）
- **Node List File**: 独立的节点配置文件 `nodelist.toml`
- **Config Directory**: 配置文件目录 `config/`
- **Legacy Configuration**: 在 `config.toml` 中包含节点配置的旧格式
- **Hot Reload**: 运行时重新加载配置而不重启应用

## Requirements

### Requirement 1

**User Story:** 作为系统管理员，我希望将节点配置独立管理，以便在节点数量增长时保持配置文件的可维护性。

#### Acceptance Criteria

1. WHEN the System starts THEN the System SHALL load node configurations from `config/nodelist.toml` if it exists
2. WHEN `config/nodelist.toml` contains node entries THEN the System SHALL parse each node with the same structure as legacy format (name, environment, host, port, username, password)
3. WHEN both `config.toml` and `config/nodelist.toml` contain node configurations THEN the System SHALL merge nodes from both sources
4. WHEN a node name appears in both files THEN the System SHALL use the configuration from `config/nodelist.toml` and log a warning about the duplicate
5. WHEN `config/nodelist.toml` does not exist THEN the System SHALL load nodes from `config.toml` without errors

### Requirement 2

**User Story:** 作为系统管理员，我希望配置文件结构清晰，以便快速理解系统配置和节点配置的职责边界。

#### Acceptance Criteria

1. WHEN the System initializes THEN the System SHALL create a `config/` directory if it does not exist
2. WHEN the System starts for the first time THEN the System SHALL provide example configuration files in the `config/` directory
3. WHEN reading configuration THEN the System SHALL load `config/config.toml` for system-level settings
4. WHEN reading configuration THEN the System SHALL load `config/nodelist.toml` for node-specific settings
5. WHEN environment variables are referenced in either file THEN the System SHALL expand them correctly (e.g., `${NODE_PASSWORD}`)

### Requirement 3

**User Story:** 作为系统管理员，我希望能够独立热重载节点列表，以便在不影响系统配置的情况下动态添加或修改节点。

#### Acceptance Criteria

1. WHEN `config/nodelist.toml` is modified THEN the System SHALL detect the file change
2. WHEN `config/nodelist.toml` is reloaded THEN the System SHALL validate the new node configurations
3. WHEN node configuration validation fails THEN the System SHALL keep the current node list and log the validation error
4. WHEN node configuration validation succeeds THEN the System SHALL update the active node list
5. WHEN the node list is updated THEN the System SHALL trigger the existing configuration callback mechanism

### Requirement 4

**User Story:** 作为现有用户，我希望升级后我的配置仍然有效，以便无需手动迁移配置文件。

#### Acceptance Criteria

1. WHEN the System starts with only `config.toml` containing nodes THEN the System SHALL load nodes from `config.toml` successfully
2. WHEN the System detects legacy node configuration THEN the System SHALL log an informational message suggesting migration to `config/nodelist.toml`
3. WHEN the System runs with legacy configuration THEN the System SHALL function identically to previous versions
4. WHEN the System provides migration guidance THEN the System SHALL include clear steps in the log output
5. WHEN both configuration formats are present THEN the System SHALL prioritize `config/nodelist.toml` for overlapping node names

### Requirement 5

**User Story:** 作为开发者，我希望配置加载逻辑清晰且可测试，以便维护和扩展配置系统。

#### Acceptance Criteria

1. WHEN loading configuration THEN the System SHALL use a single unified configuration structure
2. WHEN parsing node configurations THEN the System SHALL use the same validation logic for both file sources
3. WHEN configuration errors occur THEN the System SHALL provide clear error messages indicating the file and line number
4. WHEN testing configuration loading THEN the System SHALL support loading from custom paths
5. WHEN configuration is accessed THEN the System SHALL provide thread-safe read access to the merged configuration
