# Implementation Plan

- [x] 1. Create configuration directory structure and example files
  - Create `config/` directory if it doesn't exist
  - Create `config/config.example.toml` with system configuration template
  - Create `config/nodelist.example.toml` with node configuration template
  - Update `.gitignore` to exclude `config/config.toml` and `config/nodelist.toml` but include example files
  - _Requirements: 2.1, 2.2_

- [x] 2. Implement ConfigLoader for multi-source configuration loading
  - [x] 2.1 Create `internal/config/loader.go` with ConfigLoader struct
    - Implement `NewConfigLoader(mainPath, nodeListPath string)` constructor
    - Implement `LoadMainConfig()` to load system configuration from config.toml
    - Implement `LoadNodeList()` to load nodes from nodelist.toml (return empty slice if file doesn't exist)
    - Implement `Load()` to load and merge complete configuration
    - _Requirements: 1.1, 1.5, 2.3, 2.4_

  - [x] 2.2 Write property test for node configuration parsing round-trip
    - **Property 1: Node configuration parsing round-trip**
    - **Validates: Requirements 1.2**

  - [x] 2.3 Implement `MergeNodes()` for node list merging with priority
    - Accept two slices of NodeConfig (main and nodelist)
    - Create map to track node names from nodelist
    - Add all nodes from nodelist first
    - Add nodes from main config only if name doesn't exist
    - Log warning for duplicate node names
    - Return merged slice
    - _Requirements: 1.3, 1.4, 4.5_

  - [x] 2.4 Write property test for node list merging with priority
    - **Property 2: Node list merging with priority**
    - **Validates: Requirements 1.3, 1.4**

  - [x] 2.5 Implement environment variable expansion in configuration
    - Add `expandEnvVars(cfg *Config)` function
    - Use `os.ExpandEnv()` for string fields containing `${VAR_NAME}`
    - Apply to node passwords, admin password, database path, etc.
    - _Requirements: 2.5_

  - [x] 2.6 Write property test for environment variable expansion
    - **Property 3: Environment variable expansion**
    - **Validates: Requirements 2.5**

  - [x] 2.7 Write unit tests for ConfigLoader
    - Test loading config.toml only (legacy mode)
    - Test loading nodelist.toml only
    - Test loading both files
    - Test nodelist.toml not existing
    - Test empty node lists
    - _Requirements: 1.1, 1.5, 2.3, 2.4_

- [x] 3. Update configuration validation to support both sources
  - [x] 3.1 Refactor `internal/config/validator.go` to validate NodeConfig consistently
    - Extract node validation logic into `ValidateNode(node NodeConfig)` function
    - Ensure validation is source-agnostic
    - Validate required fields (name, host, port)
    - Validate port range (1-65535)
    - _Requirements: 5.2_

  - [x] 3.2 Write property test for validation consistency across sources
    - **Property 4: Validation consistency across sources**
    - **Validates: Requirements 5.2**

  - [x] 3.3 Enhance error messages to include source file information
    - Wrap TOML parsing errors with file path
    - Include line numbers when available from viper/toml parser
    - Format: "config/nodelist.toml:15: invalid port value"
    - _Requirements: 5.3_

  - [x] 3.4 Write property test for error message source information
    - **Property 7: Error messages contain source information**
    - **Validates: Requirements 5.3**

- [x] 4. Implement independent hot reload for node list
  - [x] 4.1 Extend ConfigManager to support node list watching
    - Add `nodeListWatcher *fsnotify.Watcher` field
    - Add `nodeListCallback func([]NodeConfig)` field
    - Implement `WatchNodeList(callback func([]NodeConfig))` method
    - Create separate goroutine for nodelist file watching
    - _Requirements: 3.1_

  - [x] 4.2 Implement node list reload logic with validation
    - On nodelist.toml change, reload using ConfigLoader
    - Validate new node configurations
    - If validation fails, keep current nodes and log error
    - If validation succeeds, update Config.Nodes atomically
    - Trigger node list callback
    - _Requirements: 3.2, 3.3, 3.4, 3.5_

  - [x] 4.3 Write property test for hot reload atomicity
    - **Property 5: Hot reload atomicity**
    - **Validates: Requirements 3.2, 3.3, 3.4**

  - [x] 4.4 Write unit tests for node list hot reload
    - Test file change detection
    - Test validation failure keeps old config
    - Test validation success updates config
    - Test callback triggering
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 5. Ensure thread-safe configuration access
  - [x] 5.1 Review and verify RWMutex usage in ConfigManager
    - Ensure all Config reads use RLock
    - Ensure all Config writes use Lock
    - Verify no race conditions in hot reload paths
    - _Requirements: 5.5_

  - [x] 5.2 Write property test for concurrent read safety
    - **Property 6: Concurrent read safety**
    - **Validates: Requirements 5.5**

- [x] 6. Update main application to use new configuration structure
  - [x] 6.1 Update `cmd/main.go` to use ConfigLoader
    - Replace direct `config.Load()` with `ConfigLoader.Load()`
    - Set paths to `config/config.toml` and `config/nodelist.toml`
    - Fall back to `config.toml` if `config/` directory doesn't exist (backward compatibility)
    - _Requirements: 1.5, 4.1_

  - [x] 6.2 Add migration guidance logging
    - Detect if `config.toml` contains `[[nodes]]` section
    - If nodelist.toml doesn't exist, log informational message
    - Message should suggest creating config/ directory and moving nodes
    - Include clear migration steps in log output
    - _Requirements: 4.2, 4.4_

  - [x] 6.3 Write integration tests for application startup
    - Test startup with legacy config.toml only
    - Test startup with config/ directory structure
    - Test startup with both formats present
    - Test migration guidance logging
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

- [x] 7. Update documentation and examples
  - Update README.md with new configuration structure
  - Document migration steps from legacy format
  - Add comments to example configuration files
  - Update configuration section in product documentation
  - _Requirements: 2.2, 4.4_

- [x] 8. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.
