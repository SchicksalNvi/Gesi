# Design Document

## Overview

本设计为 Go-CESI 添加环境管理功能，通过新增两个前端页面（环境列表和环境详情）复用现有后端 API，实现按环境维度聚合和管理节点的能力。设计遵循现有架构模式，零侵入式集成到当前系统。

## Architecture

### 系统层次

```
┌─────────────────────────────────────────┐
│         Frontend (React)                │
│  ┌────────────────────────────────────┐ │
│  │  Environment Pages                 │ │
│  │  - EnvironmentList                 │ │
│  │  - EnvironmentDetail               │ │
│  └────────────────────────────────────┘ │
│  ┌────────────────────────────────────┐ │
│  │  API Client                        │ │
│  │  - environmentsApi                 │ │
│  └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
                  ↓ HTTP/WebSocket
┌─────────────────────────────────────────┐
│         Backend (Go)                    │
│  ┌────────────────────────────────────┐ │
│  │  Existing APIs                     │ │
│  │  - GET /api/environments           │ │
│  │  - GET /api/environments/:name     │ │
│  └────────────────────────────────────┘ │
│  ┌────────────────────────────────────┐ │
│  │  SupervisorService                 │ │
│  │  - GetEnvironments()               │ │
│  │  - GetEnvironmentDetails()         │ │
│  └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

### 设计原则

1. **复用优先**：完全复用现有后端 API 和 WebSocket 机制
2. **零破坏性**：不修改现有数据结构和接口
3. **一致性**：遵循现有页面的布局和交互模式（参考 Nodes 页面）

## Components and Interfaces

### Frontend Components

#### 1. EnvironmentList 页面

**职责**：显示所有环境的卡片列表

**Props**: 无

**State**:
```typescript
{
  environments: Environment[],  // 环境列表
  loading: boolean,              // 加载状态
}
```

**主要方法**:
- `loadEnvironments()`: 调用 API 加载环境列表
- `handleEnvironmentClick(envName: string)`: 导航到环境详情页

#### 2. EnvironmentDetail 页面

**职责**：显示特定环境下的所有节点

**Props**: 无（从 URL 参数获取环境名）

**State**:
```typescript
{
  environment: EnvironmentDetail | null,  // 环境详情
  loading: boolean,                       // 加载状态
  error: string | null,                   // 错误信息
}
```

**主要方法**:
- `loadEnvironmentDetail(envName: string)`: 加载环境详情
- `handleNodeClick(nodeName: string)`: 导航到节点详情页

#### 3. EnvironmentCard 组件

**职责**：环境卡片展示

**Props**:
```typescript
{
  environment: Environment,
  onClick: () => void,
}
```

### API Client

#### environmentsApi

```typescript
export const environmentsApi = {
  // 获取所有环境
  getEnvironments: () => apiClient.get<EnvironmentsResponse>('/environments'),
  
  // 获取特定环境详情
  getEnvironmentDetail: (name: string) => 
    apiClient.get<EnvironmentDetailResponse>(`/environments/${name}`),
};
```

### Routes

新增路由：
```typescript
<Route path="environments" element={<EnvironmentList />} />
<Route path="environments/:environmentName" element={<EnvironmentDetail />} />
```

## Data Models

### Frontend Types

```typescript
// 环境列表项
interface Environment {
  name: string;           // 环境名称
  members: NodeSummary[]; // 节点列表
}

// 节点摘要（用于环境列表）
interface NodeSummary {
  name: string;
  host: string;
  port: number;
  is_connected: boolean;
  last_ping: string;
}

// 环境详情
interface EnvironmentDetail {
  name: string;
  members: NodeDetail[];
}

// 节点详情（用于环境详情页）
interface NodeDetail extends NodeSummary {
  processes: number;  // 进程数量
}

// API 响应
interface EnvironmentsResponse {
  status: string;
  environments: Environment[];
}

interface EnvironmentDetailResponse {
  status: string;
  environment: EnvironmentDetail;
}
```

### Backend Models

**无需修改**，复用现有结构：
- `Node` 模型已有 `Environment` 字段
- `SupervisorService` 已实现环境分组逻辑

## Correctness Properties


*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Environment list renders all required fields

*For any* environment data returned by the API, the rendered environment card should display the environment name, total node count, online node count, and offline node count.

**Validates: Requirements 1.2**

### Property 2: Navigation to environment detail

*For any* environment in the list, clicking on it should navigate to the URL `/environments/{environmentName}`.

**Validates: Requirements 2.1**

### Property 3: Environment detail renders environment name and nodes

*For any* environment detail data, the rendered page should display the environment name and a list of all nodes in that environment.

**Validates: Requirements 2.2**

### Property 4: Node list renders all required fields

*For any* node in an environment, the rendered node item should display name, host, port, connection status, and last ping time.

**Validates: Requirements 2.3**

### Property 5: Navigation to node detail

*For any* node in the environment detail page, clicking on it should navigate to the URL `/nodes/{nodeName}`.

**Validates: Requirements 3.1**

### Property 6: WebSocket subscription on mount

*For any* environment page (list or detail), when the component mounts, it should establish a WebSocket connection and subscribe to node status updates.

**Validates: Requirements 4.1**

### Property 7: Real-time status update

*For any* node status change event received via WebSocket, the displayed node status should update to reflect the new state.

**Validates: Requirements 4.2**

### Property 8: WebSocket cleanup on unmount

*For any* environment page, when the component unmounts, it should unsubscribe from WebSocket events and clean up resources.

**Validates: Requirements 4.3**

### Property 9: Status tag color mapping

*For any* node, if the node is online (is_connected = true), the status tag should be green; if offline (is_connected = false), the status tag should be red.

**Validates: Requirements 5.2**

### Property 10: Loading state display

*For any* page state where loading is true, the page should display a loading spinner or skeleton.

**Validates: Requirements 5.4**

## Error Handling

### Frontend Error Scenarios

1. **API 调用失败**
   - 显示错误提示（使用 Ant Design Message 组件）
   - 提供重试按钮
   - 记录错误到控制台

2. **环境不存在（404）**
   - 显示友好的 404 页面
   - 提供返回环境列表的链接

3. **WebSocket 连接失败**
   - 降级到轮询模式（可选）
   - 显示连接状态提示

4. **空数据状态**
   - 环境列表为空：显示 Empty 组件，提示配置节点
   - 环境下无节点：显示提示信息

### Error Handling Pattern

```typescript
try {
  const response = await environmentsApi.getEnvironments();
  setEnvironments(response.environments);
} catch (error) {
  console.error('Failed to load environments:', error);
  message.error('Failed to load environments. Please try again.');
}
```

## Testing Strategy

### Unit Testing

使用 **React Testing Library** 和 **Jest** 进行组件测试：

1. **EnvironmentList 组件**
   - 测试空状态渲染
   - 测试加载状态渲染
   - 测试错误状态渲染
   - 测试环境卡片点击导航

2. **EnvironmentDetail 组件**
   - 测试环境不存在时的 404 处理
   - 测试节点列表渲染
   - 测试节点点击导航

3. **API Client**
   - Mock axios 测试 API 调用
   - 测试错误处理

### Property-Based Testing

使用 **fast-check** 库进行属性测试：

**配置**：每个属性测试运行 100 次迭代

**测试策略**：
- 生成随机环境数据（环境名、节点列表、连接状态）
- 验证渲染输出包含所有必需字段
- 验证导航行为的正确性
- 验证状态映射的一致性

**标注格式**：每个属性测试必须包含注释：
```typescript
// **Feature: environment-management, Property 1: Environment list renders all required fields**
```

### Integration Testing

1. **路由集成**
   - 测试从环境列表到环境详情的完整导航流程
   - 测试从环境详情到节点详情的导航

2. **WebSocket 集成**
   - 测试实时状态更新
   - 测试订阅和取消订阅

## Implementation Notes

### 1. 复用现有组件

- 复用 `NodeCard` 组件的样式和布局模式
- 复用 `useWebSocket` hook 处理实时更新
- 复用 `apiClient` 进行 API 调用

### 2. 性能优化

- 使用 React.memo 优化环境卡片渲染
- 使用 useMemo 缓存计算的统计数据（在线/离线节点数）
- WebSocket 事件去重，避免重复渲染

### 3. 可访问性

- 所有交互元素支持键盘导航
- 使用语义化 HTML 标签
- 为状态标签添加 aria-label

### 4. 响应式设计

- 使用 Ant Design Grid 系统
- 移动端：单列布局
- 平板：双列布局
- 桌面：三列布局

## Deployment Considerations

### 无需后端变更

- 后端 API 已存在，无需部署新服务
- 仅需部署前端静态资源

### 前端部署步骤

1. 构建 React 应用：`npm run build`
2. 将 `dist/` 目录内容部署到 Go 服务器的静态资源目录
3. 重启 Go 服务器（如需要）

### 兼容性

- 向后兼容：不影响现有功能
- 浏览器支持：Chrome 90+, Firefox 88+, Safari 14+
