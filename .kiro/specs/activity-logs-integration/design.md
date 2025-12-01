# Design Document

## Overview

本设计文档描述了 Activity Logs 页面从 mock 数据迁移到真实 API 集成的完整方案。系统将移除前端所有硬编码的模拟数据，实现与后端 Activity Log API 的完整集成，提供用户操作审计和系统事件追踪功能。

设计遵循现有的架构模式：
- 前端使用 React + Ant Design + TypeScript
- 后端使用 Go + Gin + GORM
- API 通信使用 RESTful 风格
- 认证使用 JWT Bearer Token

## Architecture

### System Components

```
┌─────────────────────────────────────────────────────────────┐
│                      Frontend (React)                        │
│  ┌────────────────┐  ┌──────────────┐  ┌─────────────────┐ │
│  │  Logs Page     │  │  API Client  │  │  Type Definitions│ │
│  │  Component     │──│  (activityLogs)│──│  (ActivityLog)  │ │
│  └────────────────┘  └──────────────┘  └─────────────────┘ │
└──────────────────────────┬──────────────────────────────────┘
                           │ HTTP/JSON
                           │ (JWT Auth)
┌──────────────────────────┴──────────────────────────────────┐
│                      Backend (Go/Gin)                        │
│  ┌────────────────┐  ┌──────────────┐  ┌─────────────────┐ │
│  │  Activity Logs │  │  Activity Log│  │  Activity Log   │ │
│  │  API Handler   │──│  Service     │──│  Repository     │ │
│  └────────────────┘  └──────────────┘  └─────────────────┘ │
└──────────────────────────┬──────────────────────────────────┘
                           │
                    ┌──────┴──────┐
                    │   SQLite    │
                    │  (activity_ │
                    │   logs)     │
                    └─────────────┘
```

### Data Flow

1. **用户操作记录流程**：
   ```
   User Action → API Handler → Activity Log Service → Database
   ```

2. **系统事件记录流程**：
   ```
   System Event → Supervisor Service → Activity Log Service → Database
   ```

3. **日志查询流程**：
   ```
   Frontend → API Client → Activity Logs API → Service → Database → Response
   ```

## Components and Interfaces

### Frontend Components

#### 1. Activity Logs API Client (`web/react-app/src/api/activityLogs.ts`)

新建文件，提供与后端 Activity Logs API 的交互接口：

```typescript
interface ActivityLog {
  id: number;
  created_at: string;
  level: string;
  message: string;
  action: string;
  resource: string;
  target: string;
  user_id: string;
  username: string;
  ip_address: string;
  user_agent: string;
  details: string;
  status: string;
  duration: number;
}

interface ActivityLogsResponse {
  status: string;
  data: {
    logs: ActivityLog[];
    pagination: {
      page: number;
      page_size: number;
      total: number;
      total_pages: number;
      has_next: boolean;
      has_prev: boolean;
    };
  };
}

interface ActivityLogsFilters {
  level?: string;
  action?: string;
  resource?: string;
  username?: string;
  start_time?: string;
  end_time?: string;
  page?: number;
  page_size?: number;
}

// API 方法
getActivityLogs(filters: ActivityLogsFilters): Promise<ActivityLogsResponse>
getRecentLogs(limit: number): Promise<ActivityLog[]>
getLogStatistics(days: number): Promise<LogStatistics>
exportLogs(filters: ActivityLogsFilters): Promise<Blob>
```

#### 2. Logs Page Component (`web/react-app/src/pages/Logs/index.tsx`)

重构现有组件，移除 mock 数据：

**主要变更**：
- 移除所有 `mockSystemLogs` 和 `mockActivityLogs` 数据
- 使用 `activityLogsAPI` 调用真实接口
- 实现分页、筛选、搜索功能
- 添加错误处理和加载状态
- 实现自动刷新（轮询）
- 实现导出功能

**状态管理**：
```typescript
const [activityLogs, setActivityLogs] = useState<ActivityLog[]>([]);
const [loading, setLoading] = useState(false);
const [pagination, setPagination] = useState({...});
const [filters, setFilters] = useState<ActivityLogsFilters>({});
const [autoRefresh, setAutoRefresh] = useState(true);
```

### Backend Components

#### 1. Activity Logs API Handler (`internal/api/activity_logs.go`)

现有实现已完整，包含以下端点：
- `GET /api/activity-logs` - 获取活动日志列表（支持分页和筛选）
- `GET /api/activity-logs/recent` - 获取最近的日志
- `GET /api/activity-logs/statistics` - 获取日志统计信息
- `DELETE /api/activity-logs/clean` - 清理旧日志

**需要新增**：
- `GET /api/activity-logs/export` - 导出日志为 CSV

#### 2. Activity Log Service (`internal/services/activity_log.go`)

现有实现已完整，提供以下功能：
- `LogActivity()` - 记录活动日志
- `LogWithContext()` - 从 Gin 上下文记录日志
- `LogError()` - 记录错误日志
- `GetActivityLogs()` - 查询日志（支持分页和筛选）
- `GetRecentLogs()` - 获取最近日志
- `GetLogStatistics()` - 获取统计信息
- `CleanOldLogs()` - 清理旧日志

**便捷方法**：
- `LogLogin()` - 记录登录
- `LogLogout()` - 记录登出
- `LogProcessAction()` - 记录进程操作
- `LogGroupAction()` - 记录组操作
- `LogUserAction()` - 记录用户操作

**需要新增**：
- `ExportLogs()` - 导出日志为 CSV 格式
- `LogSystemEvent()` - 记录系统事件（进程掉线、节点断连等）

#### 3. Integration Points

需要在以下位置集成日志记录：

**进程操作**（`internal/api/processes.go`）：
- `StartProcess()` - 调用 `activityLogService.LogProcessAction(c, "start", nodeName, processName)`
- `StopProcess()` - 调用 `activityLogService.LogProcessAction(c, "stop", nodeName, processName)`
- `RestartProcess()` - 调用 `activityLogService.LogProcessAction(c, "restart", nodeName, processName)`

**系统事件**（`internal/supervisor/service.go`）：
- 进程状态变化监控
- 节点连接状态监控

## Data Models

### Activity Log Model

现有模型（`internal/models/activity_log.go`）已完整定义：

```go
type ActivityLog struct {
    ID        uint
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt
    
    // 日志基本信息
    Level    string  // INFO, WARNING, ERROR
    Message  string  // 日志消息
    Action   string  // 操作类型
    Resource string  // 资源类型
    Target   string  // 目标对象
    
    // 用户信息
    UserID   string
    Username string
    
    // 请求信息
    IPAddress string
    UserAgent string
    
    // 额外信息
    Details  string  // JSON 格式
    Status   string  // success, error, warning
    Duration int64   // 毫秒
}
```

### Frontend Type Definitions

在 `web/react-app/src/types/index.ts` 中添加：

```typescript
export interface ActivityLog {
  id: number;
  created_at: string;
  level: 'INFO' | 'WARNING' | 'ERROR';
  message: string;
  action: string;
  resource: string;
  target: string;
  user_id: string;
  username: string;
  ip_address: string;
  user_agent: string;
  details: string;
  status: 'success' | 'error' | 'warning';
  duration: number;
}

export interface ActivityLogsFilters {
  level?: string;
  action?: string;
  resource?: string;
  username?: string;
  start_time?: string;
  end_time?: string;
  page?: number;
  page_size?: number;
}

export interface PaginationInfo {
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: API Response Consistency

*For any* valid API request to `/api/activity-logs`, the response structure should always match the defined `ActivityLogsResponse` interface with status, data, logs array, and pagination metadata.

**Validates: Requirements 1.1, 3.5**

### Property 2: Filter Application Correctness

*For any* combination of filters (level, action, resource, username, time range), the returned logs should only include entries that match ALL specified filter criteria.

**Validates: Requirements 2.1, 2.2, 2.3, 2.4, 2.5**

### Property 3: Pagination Consistency

*For any* page number and page size, the sum of all logs across all pages should equal the total count, and no log entry should appear on multiple pages.

**Validates: Requirements 3.1, 3.2, 3.3, 3.4**

### Property 4: Mock Data Absence

*For any* code path in the frontend Logs component, there should be no references to hardcoded mock data arrays or mock data generation functions.

**Validates: Requirements 4.1, 4.2, 4.3, 4.4, 4.5**

### Property 5: User Action Logging Completeness

*For any* user-initiated process operation (start, stop, restart), an activity log entry should be created with the correct username, action, node name, process name, and timestamp.

**Validates: Requirements 1.3**

### Property 6: System Event Logging Completeness

*For any* system event (process state change, node disconnection), an activity log entry should be created with source marked as "system" and containing the relevant event details.

**Validates: Requirements 1.4, 5.1, 5.2, 5.3, 5.4, 5.5**

### Property 7: Real-time Update Consistency

*For any* new log entry created in the database, if the user is on the first page of the Activity Logs view and auto-refresh is enabled, the new entry should appear in the UI within the polling interval.

**Validates: Requirements 6.1, 6.2, 6.4**

### Property 8: Export Data Completeness

*For any* export request with filters, the exported CSV file should contain exactly the same log entries that would be displayed in the UI with those filters applied.

**Validates: Requirements 7.1, 7.2, 7.3, 7.4**

## Error Handling

### Frontend Error Handling

1. **API 调用失败**：
   - 显示用户友好的错误消息
   - 不回退到 mock 数据
   - 提供重试选项

2. **网络超时**：
   - 显示超时提示
   - 自动重试机制（最多 3 次）

3. **认证失败**：
   - 重定向到登录页面
   - 清除本地存储的 token

4. **数据格式错误**：
   - 记录错误到控制台
   - 显示通用错误消息

### Backend Error Handling

1. **数据库查询失败**：
   - 返回 500 错误
   - 记录详细错误日志

2. **参数验证失败**：
   - 返回 400 错误
   - 返回具体的验证错误信息

3. **权限不足**：
   - 返回 403 错误
   - 记录未授权访问尝试

## Testing Strategy

### Unit Tests

#### Frontend Unit Tests

**文件**: `web/react-app/src/pages/Logs/__tests__/index.test.tsx`

测试内容：
1. 组件正确渲染
2. API 调用正确触发
3. 筛选器正确应用
4. 分页控件正确工作
5. 错误状态正确显示
6. 加载状态正确显示

**文件**: `web/react-app/src/api/__tests__/activityLogs.test.ts`

测试内容：
1. API 方法正确构造请求
2. 响应正确解析
3. 错误正确处理

#### Backend Unit Tests

**文件**: `internal/api/activity_logs_test.go`

测试内容：
1. 各端点正确响应
2. 参数验证正确工作
3. 分页逻辑正确
4. 筛选逻辑正确
5. 错误情况正确处理

**文件**: `internal/services/activity_log_test.go`

测试内容：
1. 日志记录功能正确
2. 查询功能正确
3. 统计功能正确
4. 导出功能正确

### Property-Based Tests

使用 Go 的 `testing/quick` 包和 TypeScript 的 `fast-check` 库进行属性测试。

#### Backend Property Tests

**文件**: `internal/services/activity_log_property_test.go`

测试属性：
- Property 2: Filter Application Correctness
- Property 3: Pagination Consistency
- Property 5: User Action Logging Completeness
- Property 6: System Event Logging Completeness

#### Frontend Property Tests

**文件**: `web/react-app/src/api/__tests__/activityLogs.property.test.ts`

测试属性：
- Property 1: API Response Consistency
- Property 4: Mock Data Absence

### Integration Tests

**文件**: `internal/api/activity_logs_integration_test.go`

测试场景：
1. 完整的日志记录和查询流程
2. 多用户并发记录日志
3. 大量日志的分页查询
4. 复杂筛选条件的组合
5. 导出大量日志数据

### Manual Testing Checklist

1. 访问 Activity Logs 页面，验证数据正确加载
2. 测试各种筛选条件组合
3. 测试分页功能
4. 执行进程操作，验证日志正确记录
5. 测试导出功能
6. 测试自动刷新功能
7. 测试错误场景（网络断开、API 错误等）

## Implementation Notes

### 移除 Mock 数据的步骤

1. 删除 `mockSystemLogs` 和 `mockActivityLogs` 数组定义
2. 删除 `loadLogs()` 中的 mock 数据生成逻辑
3. 替换为真实的 API 调用
4. 确保所有条件分支都使用真实数据

### 自动刷新实现

使用 `setInterval` 实现轮询：
- 默认间隔：30 秒
- 仅在第一页时自动刷新
- 用户可以关闭自动刷新
- 组件卸载时清理定时器

### 导出功能实现

后端：
- 生成 CSV 格式
- 包含所有字段的列标题
- 支持大文件流式输出

前端：
- 使用 Blob 和 URL.createObjectURL
- 触发浏览器下载
- 显示导出进度

### 系统事件记录集成

在 Supervisor Service 中添加状态变化监听：
- 监听进程状态变化事件
- 监听节点连接状态变化
- 调用 Activity Log Service 记录事件

## Performance Considerations

1. **数据库索引**：
   - `created_at` 字段索引（用于时间范围查询）
   - `username` 字段索引（用于用户筛选）
   - `action` 字段索引（用于操作类型筛选）

2. **分页优化**：
   - 使用 LIMIT 和 OFFSET
   - 避免查询总数时的全表扫描

3. **前端优化**：
   - 使用 React.memo 避免不必要的重渲染
   - 防抖搜索输入
   - 虚拟滚动（如果日志量很大）

4. **缓存策略**：
   - 统计数据可以缓存 5 分钟
   - 最近日志可以缓存 30 秒

## Security Considerations

1. **认证**：所有 API 端点都需要 JWT 认证
2. **授权**：仅管理员可以清理日志和导出日志
3. **输入验证**：所有用户输入都需要验证和清理
4. **SQL 注入防护**：使用 GORM 的参数化查询
5. **敏感信息**：不记录密码等敏感信息到日志

## Migration Plan

1. **Phase 1**: 创建前端 API 客户端
2. **Phase 2**: 重构 Logs 页面组件，移除 mock 数据
3. **Phase 3**: 添加导出功能到后端
4. **Phase 4**: 集成系统事件记录
5. **Phase 5**: 添加自动刷新功能
6. **Phase 6**: 测试和优化
