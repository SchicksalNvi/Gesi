-- 数据库索引优化迁移
-- 添加复合索引以提升查询性能

-- LogEntry表的复合索引
-- 用于日志查询中的常见过滤条件组合
CREATE INDEX IF NOT EXISTS idx_log_entries_timestamp_level ON log_entries(timestamp, level);
CREATE INDEX IF NOT EXISTS idx_log_entries_node_level_timestamp ON log_entries(node_id, level, timestamp);
CREATE INDEX IF NOT EXISTS idx_log_entries_source_timestamp ON log_entries(source, timestamp);
CREATE INDEX IF NOT EXISTS idx_log_entries_process_timestamp ON log_entries(process_name, timestamp);
CREATE INDEX IF NOT EXISTS idx_log_entries_category_severity ON log_entries(category, severity);
CREATE INDEX IF NOT EXISTS idx_log_entries_parsed_archived ON log_entries(parsed, archived);

-- LogAlert表的复合索引
-- 用于告警查询中的状态和时间过滤
CREATE INDEX IF NOT EXISTS idx_log_alerts_status_level ON log_alerts(status, level);
CREATE INDEX IF NOT EXISTS idx_log_alerts_acknowledged_resolved ON log_alerts(acknowledged, resolved);
CREATE INDEX IF NOT EXISTS idx_log_alerts_rule_status ON log_alerts(rule_id, status);
CREATE INDEX IF NOT EXISTS idx_log_alerts_first_seen_status ON log_alerts(first_seen, status);

-- LogStatistics表的复合索引
-- 用于统计查询中的时间和分组条件
CREATE INDEX IF NOT EXISTS idx_log_statistics_date_hour ON log_statistics(date, hour);
CREATE INDEX IF NOT EXISTS idx_log_statistics_date_level ON log_statistics(date, level);
CREATE INDEX IF NOT EXISTS idx_log_statistics_node_date ON log_statistics(node_id, date);
CREATE INDEX IF NOT EXISTS idx_log_statistics_source_date ON log_statistics(source, date);

-- Configuration表的复合索引
-- 用于配置查询中的作用域和分类过滤
CREATE INDEX IF NOT EXISTS idx_configurations_scope_category ON configurations(scope, category);
CREATE INDEX IF NOT EXISTS idx_configurations_node_scope ON configurations(node_id, scope);
CREATE INDEX IF NOT EXISTS idx_configurations_user_scope ON configurations(user_id, scope);
CREATE INDEX IF NOT EXISTS idx_configurations_category_type ON configurations(category, type);

-- ConfigurationHistory表的复合索引
-- 用于配置历史查询中的时间和变更类型过滤
CREATE INDEX IF NOT EXISTS idx_config_history_created_at_change_type ON configuration_histories(created_at, change_type);
CREATE INDEX IF NOT EXISTS idx_config_history_config_created_at ON configuration_histories(config_id, created_at);
CREATE INDEX IF NOT EXISTS idx_config_history_created_by_created_at ON configuration_histories(created_by, created_at);

-- BackupRecord表的复合索引
-- 用于备份记录查询中的类型和状态过滤
CREATE INDEX IF NOT EXISTS idx_backup_records_backup_type_status ON backup_records(backup_type, status);
CREATE INDEX IF NOT EXISTS idx_backup_records_created_at_status ON backup_records(created_at, status);
CREATE INDEX IF NOT EXISTS idx_backup_records_created_by_created_at ON backup_records(created_by, created_at);

-- DataExportRecord表的复合索引
-- 用于导出记录查询中的类型和状态过滤
CREATE INDEX IF NOT EXISTS idx_export_records_export_type_status ON data_export_records(export_type, status);
CREATE INDEX IF NOT EXISTS idx_export_records_created_at_status ON data_export_records(created_at, status);
CREATE INDEX IF NOT EXISTS idx_export_records_created_by_created_at ON data_export_records(created_by, created_at);

-- DataImportRecord表的复合索引
-- 用于导入记录查询中的类型和状态过滤
CREATE INDEX IF NOT EXISTS idx_import_records_import_type_status ON data_import_records(import_type, status);
CREATE INDEX IF NOT EXISTS idx_import_records_created_at_status ON data_import_records(created_at, status);
CREATE INDEX IF NOT EXISTS idx_import_records_created_by_created_at ON data_import_records(created_by, created_at);

-- SystemSettings表的复合索引
-- 用于系统设置查询中的分类过滤
CREATE INDEX IF NOT EXISTS idx_system_settings_category_key ON system_settings(category, key);
CREATE INDEX IF NOT EXISTS idx_system_settings_is_public_category ON system_settings(is_public, category);

-- WebhookLog表的复合索引
-- 用于Webhook日志查询中的状态和时间过滤
CREATE INDEX IF NOT EXISTS idx_webhook_logs_webhook_created_at ON webhook_logs(webhook_id, created_at);
CREATE INDEX IF NOT EXISTS idx_webhook_logs_success_created_at ON webhook_logs(success, created_at);
CREATE INDEX IF NOT EXISTS idx_webhook_logs_event_created_at ON webhook_logs(event, created_at);

-- ActivityLog表的复合索引
-- 用于活动日志查询中的用户和时间过滤
CREATE INDEX IF NOT EXISTS idx_activity_logs_user_created_at ON activity_logs(username, created_at);
CREATE INDEX IF NOT EXISTS idx_activity_logs_action_created_at ON activity_logs(action, created_at);
CREATE INDEX IF NOT EXISTS idx_activity_logs_level_created_at ON activity_logs(level, created_at);
CREATE INDEX IF NOT EXISTS idx_activity_logs_resource_created_at ON activity_logs(resource, created_at);
CREATE INDEX IF NOT EXISTS idx_activity_logs_user_id_created_at ON activity_logs(user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_activity_logs_status_created_at ON activity_logs(status, created_at);

-- 为分页查询优化的索引
-- 通常按创建时间倒序排列
CREATE INDEX IF NOT EXISTS idx_log_entries_created_at_desc ON log_entries(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_log_alerts_created_at_desc ON log_alerts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_configurations_created_at_desc ON configurations(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_backup_records_created_at_desc ON backup_records(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_export_records_created_at_desc ON data_export_records(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_import_records_created_at_desc ON data_import_records(created_at DESC);

-- 为软删除查询优化的索引
-- 大多数查询都会过滤已删除的记录
CREATE INDEX IF NOT EXISTS idx_log_entries_deleted_at_null ON log_entries(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_log_alerts_deleted_at_null ON log_alerts(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_configurations_deleted_at_null ON configurations(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_backup_records_deleted_at_null ON backup_records(deleted_at) WHERE deleted_at IS NULL;

-- User表的复合索引
-- 用于用户查询中的状态和角色过滤
CREATE INDEX IF NOT EXISTS idx_users_is_active_is_admin ON users(is_active, is_admin);
CREATE INDEX IF NOT EXISTS idx_users_last_login_is_active ON users(last_login, is_active);

-- 为全文搜索优化的索引（SQLite不支持GIN索引，使用FTS虚拟表）
-- 用于日志消息搜索的FTS表
-- CREATE VIRTUAL TABLE IF NOT EXISTS log_entries_fts USING fts5(message, content='log_entries', content_rowid='id');
-- CREATE TRIGGER IF NOT EXISTS log_entries_fts_insert AFTER INSERT ON log_entries BEGIN
--   INSERT INTO log_entries_fts(rowid, message) VALUES (new.id, new.message);
-- END;