# Requirements Document

## Introduction

本需求文档定义了 Activity Logs 页面的真实数据集成功能。当前 Activity Logs 页面使用 mock 数据，需要移除 mock 数据并集成后端真实的活动日志 API，实现完整的用户操作审计和系统事件追踪功能。

## Glossary

- **Activity Log System**: 活动日志系统，记录和展示用户操作及系统事件的完整系统
- **Frontend**: React 前端应用，负责展示日志数据
- **Backend API**: Go 后端 API，提供日志数据的查询和管理接口
- **Mock Data**: 前端硬编码的模拟数据，用于开发阶段测试
- **Real-time Event**: 实时事件，如进程状态变化、节点掉线等系统事件
- **User Action**: 用户操作，如启动进程、停止进程、创建用户等
- **Audit Trail**: 审计追踪，完整记录系统中所有操作的历史记录

## Requirements

### Requirement 1

**User Story:** 作为系统管理员，我希望在 Activity Logs 页面看到真实的用户操作记录，以便审计系统使用情况和追踪问题。

#### Acceptance Criteria

1. WHEN 用户访问 Activity Logs 页面 THEN THE Frontend SHALL 调用后端 API 获取真实的活动日志数据
2. WHEN 活动日志数据返回 THEN THE Frontend SHALL 展示包含时间戳、操作用户、Source、Node、Process、Action、Message 的完整信息
3. WHEN 用户执行进程操作（启动、停止、重启）THEN THE Backend API SHALL 记录包含用户名、节点名、进程名、操作类型、时间戳的日志条目
4. WHEN 系统检测到进程状态变化 THEN THE Backend API SHALL 记录包含节点名、进程名、状态变化、时间戳的系统事件日志
5. WHEN 用户执行用户管理操作 THEN THE Backend API SHALL 记录包含操作者、目标用户、操作类型、时间戳的日志条目

### Requirement 2

**User Story:** 作为系统管理员，我希望能够筛选和搜索活动日志，以便快速定位特定的操作或事件。

#### Acceptance Criteria

1. WHEN 用户在搜索框输入关键词 THEN THE Frontend SHALL 向后端发送包含搜索参数的 API 请求
2. WHEN 用户选择时间范围 THEN THE Frontend SHALL 仅展示该时间范围内的日志条目
3. WHEN 用户选择用户名筛选 THEN THE Frontend SHALL 仅展示该用户的操作日志
4. WHEN 用户选择操作类型筛选 THEN THE Frontend SHALL 仅展示该类型的操作日志
5. WHEN 用户选择资源类型筛选（Process、Node、User）THEN THE Frontend SHALL 仅展示该资源类型的日志

### Requirement 3

**User Story:** 作为系统管理员，我希望活动日志能够分页展示，以便在大量日志中高效浏览。

#### Acceptance Criteria

1. WHEN 活动日志数量超过单页显示限制 THEN THE Frontend SHALL 展示分页控件
2. WHEN 用户点击下一页 THEN THE Frontend SHALL 请求下一页的日志数据
3. WHEN 用户更改每页显示数量 THEN THE Frontend SHALL 根据新的页面大小重新请求数据
4. WHEN 分页数据加载 THEN THE Frontend SHALL 展示总记录数和当前页码信息
5. WHEN 后端返回分页数据 THEN THE Backend API SHALL 包含 total、page、page_size、has_next、has_prev 等分页元数据

### Requirement 4

**User Story:** 作为开发人员，我希望移除所有 mock 数据代码，以便代码库保持整洁且仅包含生产代码。

#### Acceptance Criteria

1. WHEN 代码审查时 THEN THE Frontend SHALL 不包含任何硬编码的 mock 日志数据
2. WHEN 前端组件初始化 THEN THE Frontend SHALL 直接调用真实 API 而非使用 mock 数据
3. WHEN API 调用失败 THEN THE Frontend SHALL 展示适当的错误信息而非回退到 mock 数据
4. WHEN 开发环境运行 THEN THE Frontend SHALL 使用与生产环境相同的 API 调用逻辑
5. WHEN 代码构建 THEN THE Frontend SHALL 不包含任何与 mock 数据相关的条件编译代码

### Requirement 5

**User Story:** 作为系统管理员，我希望看到系统事件日志（如进程掉线、节点断连），以便及时发现和处理系统问题。

#### Acceptance Criteria

1. WHEN 进程状态从 RUNNING 变为 STOPPED THEN THE Backend API SHALL 记录进程停止事件日志
2. WHEN 进程状态从 STOPPED 变为 RUNNING THEN THE Backend API SHALL 记录进程启动事件日志
3. WHEN 节点连接断开 THEN THE Backend API SHALL 记录节点掉线事件日志
4. WHEN 节点重新连接 THEN THE Backend API SHALL 记录节点恢复事件日志
5. WHEN 系统事件发生 THEN THE Backend API SHALL 在日志中标记事件来源为 "system" 而非用户名

### Requirement 6

**User Story:** 作为系统管理员，我希望活动日志能够实时更新，以便及时看到最新的操作和事件。

#### Acceptance Criteria

1. WHEN 用户在 Activity Logs 页面 THEN THE Frontend SHALL 定期轮询后端 API 获取最新日志
2. WHEN 新日志条目到达 THEN THE Frontend SHALL 在列表顶部插入新条目
3. WHEN 用户正在浏览历史日志 THEN THE Frontend SHALL 不自动滚动到顶部以避免干扰用户
4. WHEN 轮询间隔到达 THEN THE Frontend SHALL 仅在用户位于第一页时自动刷新
5. WHEN 用户手动点击刷新按钮 THEN THE Frontend SHALL 立即重新加载当前页的日志数据

### Requirement 7

**User Story:** 作为系统管理员，我希望能够导出活动日志，以便进行离线分析或合规审计。

#### Acceptance Criteria

1. WHEN 用户点击导出按钮 THEN THE Frontend SHALL 请求后端生成日志导出文件
2. WHEN 导出请求包含筛选条件 THEN THE Backend API SHALL 仅导出符合筛选条件的日志
3. WHEN 导出文件生成完成 THEN THE Frontend SHALL 触发浏览器下载导出文件
4. WHEN 导出文件格式为 CSV THEN THE Backend API SHALL 包含所有日志字段的列标题
5. WHEN 导出数据量较大 THEN THE Frontend SHALL 展示导出进度提示
