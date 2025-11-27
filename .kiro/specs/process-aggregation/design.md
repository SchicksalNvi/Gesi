# Design Document

## Overview

进程聚合视图功能为微服务架构提供跨节点的统一进程管理界面。核心思路是按进程名称聚合分布在多个节点上的同名进程实例，提供批量操作能力，同时保留单实例精确控制。

设计遵循以下原则：
- **数据结构优先**：进程名 → 实例列表的清晰映射
- **零破坏性**：新增独立 API 和页面，复用现有基础设施
- **简洁实现**：避免过度抽象，直接解决问题

## Architecture

### 后端架构

```
┌─────────────────────────────────────────────────────────┐
│                    API Layer                            │
│  /api/processes (新增)                                  │
│    - GET /api/processes/aggregated                      │
│    - POST /api/processes/:name/start                    │
│    - POST /api/processes/:name/stop                     │
│    - POST /api/processes/:name/restart                  │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│              SupervisorService (复用)                    │
│  - GetAllNodes()                                        │
│  - GetNodeProcesses(nodeName)                           │
│  - StartProcess(nodeName, processName)                  │
│  - StopProcess(nodeName, processName)                   │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                  Node Layer (复用)                       │
│  多个 Node 实例，每个管理一个 Supervisor                 │
└─────────────────────────────────────────────────────────┘
```

### 前端架构

```
┌─────────────────────────────────────────────────────────┐
│                  /processes 路由                         │
│                                                         │
│  ┌─────────────────────────────────────────────────┐   │
│  │  ProcessAggregationPage                         │   │
│  │  - 搜索框                                        │   │
│  │  - 进程列表（可展开）                             │   │
│  │  - 批量操作按钮                                   │   │
│  └─────────────────────────────────────────────────┘   │
│                                                         │
│  ┌─────────────────────────────────────────────────┐   │
│  │  ProcessInstanceList (展开后)                    │   │
│  │  - 实例详情表格                                   │   │
│  │  - 单实例操作按钮                                 │   │
│  └─────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                  API Client                             │
│  processesApi.getAggregated()                           │
│  processesApi.batchStart(processName)                   │
│  processesApi.batchStop(processName)                    │
│  processesApi.batchRestart(processName)                 │
└─────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 后端组件

#### 1. ProcessesAPI (新增)

```go
type ProcessesAPI struct {
    service *supervisor.SupervisorService
}

// GetAggregatedProcesses 获取聚合的进程列表
func (api *ProcessesAPI) GetAggregatedProcesses(c *gin.Context) {
    // 返回格式：
    // {
    //   "status": "success",
    //   "processes": [
    //     {
    //       "name": "api-service",
    //       "total_instances": 3,
    //       "running_instances": 2,
    //       "stopped_instances": 1,
    //       "instances": [
    //         {
    //           "node_name": "node1",
    //           "node_host": "192.168.1.10",
    //           "state": 20,
    //           "state_string": "RUNNING",
    //           "pid": 12345,
    //           "uptime": 3600,
    //           "uptime_human": "1.0h"
    //         }
    //       ]
    //     }
    //   ]
    // }
}

// BatchStartProcess 批量启动进程
func (api *ProcessesAPI) BatchStartProcess(c *gin.Context) {
    // 并行调用所有节点的 StartProcess
    // 返回操作结果摘要
}

// BatchStopProcess 批量停止进程
func (api *ProcessesAPI) BatchStopProcess(c *gin.Context)

// BatchRestartProcess 批量重启进程
func (api *ProcessesAPI) BatchRestartProcess(c *gin.Context)
```

#### 2. SupervisorService (复用现有)

无需修改，直接使用现有方法：
- `GetAllNodes()` - 获取所有节点
- `GetNodeProcesses(nodeName)` - 获取节点进程
- `StartProcess(nodeName, processName)` - 启动进程
- `StopProcess(nodeName, processName)` - 停止进程

### 前端组件

#### 1. ProcessAggregationPage

主页面组件，负责：
- 显示聚合的进程列表
- 搜索和过滤
- 批量操作
- 展开/收起实例详情

#### 2. ProcessInstanceList

实例详情组件，显示：
- 节点名称和主机
- 进程状态
- PID 和运行时间
- 单实例操作按钮

#### 3. processesApi

API 客户端模块：
```typescript
export const processesApi = {
  getAggregated: () => axios.get('/api/processes/aggregated'),
  batchStart: (processName: string) => 
    axios.post(`/api/processes/${processName}/start`),
  batchStop: (processName: string) => 
    axios.post(`/api/processes/${processName}/stop`),
  batchRestart: (processName: string) => 
    axios.post(`/api/processes/${processName}/restart`),
};
```

## Data Models

### AggregatedProcess (后端返回)

```go
type AggregatedProcess struct {
    Name              string            `json:"name"`
    TotalInstances    int               `json:"total_instances"`
    RunningInstances  int               `json:"running_instances"`
    StoppedInstances  int               `json:"stopped_instances"`
    Instances         []ProcessInstance `json:"instances"`
}

type ProcessInstance struct {
    NodeName      string  `json:"node_name"`
    NodeHost      string  `json:"node_host"`
    NodePort      int     `json:"node_port"`
    State         int     `json:"state"`
    StateString   string  `json:"state_string"`
    PID           int     `json:"pid"`
    Uptime        float64 `json:"uptime"`
    UptimeHuman   string  `json:"uptime_human"`
    Description   string  `json:"description"`
    Group         string  `json:"group"`
}
```

### BatchOperationResult (批量操作返回)

```go
type BatchOperationResult struct {
    ProcessName      string                  `json:"process_name"`
    TotalInstances   int                     `json:"total_instances"`
    SuccessCount     int                     `json:"success_count"`
    FailureCount     int                     `json:"failure_count"`
    Results          []InstanceOperationResult `json:"results"`
}

type InstanceOperationResult struct {
    NodeName  string `json:"node_name"`
    Success   bool   `json:"success"`
    Error     string `json:"error,omitempty"`
}
```

### 前端类型定义

```typescript
export interface AggregatedProcess {
  name: string;
  total_instances: number;
  running_instances: number;
  stopped_instances: number;
  instances: ProcessInstance[];
}

export interface ProcessInstance {
  node_name: string;
  node_host: string;
  node_port: number;
  state: number;
  state_string: string;
  pid: number;
  uptime: number;
  uptime_human: string;
  description: string;
  group: string;
}

export interface BatchOperationResult {
  process_name: string;
  total_instances: number;
  success_count: number;
  failure_count: number;
  results: InstanceOperationResult[];
}

export interface InstanceOperationResult {
  node_name: string;
  success: boolean;
  error?: string;
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: 聚合完整性

*For any* 进程聚合请求，返回的进程列表应包含所有已连接节点上的所有进程，且每个进程名称只出现一次
**Validates: Requirements 1.1, 1.2**

### Property 2: 实例计数一致性

*For any* 聚合进程，total_instances 应等于 running_instances + stopped_instances，且应等于 instances 数组的长度
**Validates: Requirements 1.3**

### Property 3: 批量操作覆盖性

*For any* 批量操作请求，系统应对该进程名下的所有实例执行操作，且 success_count + failure_count 应等于 total_instances
**Validates: Requirements 3.1, 3.2, 3.3, 3.4**

### Property 4: 搜索过滤正确性

*For any* 搜索查询字符串，返回的进程列表中的每个进程名称都应包含该查询字符串（不区分大小写）
**Validates: Requirements 2.1**

### Property 5: 节点故障隔离性

*For any* 批量操作，如果某个节点不可用，系统应继续操作其他可用节点，且不可用节点的实例应在结果中标记为失败
**Validates: Requirements 6.1, 6.2**

### Property 6: 实时更新一致性

*For any* 进程状态变化，WebSocket 推送的更新应在 5 秒内反映在聚合视图中
**Validates: Requirements 1.5**

## Error Handling

### 后端错误处理

1. **节点不可用**
   - 跳过不可达节点，继续处理其他节点
   - 在结果中标记失败的节点和原因
   - 返回部分成功的结果

2. **进程操作失败**
   - 记录失败的具体原因（权限、进程不存在等）
   - 不中断批量操作流程
   - 在结果摘要中提供详细错误信息

3. **并发控制**
   - 使用 sync.WaitGroup 等待所有节点操作完成
   - 设置合理的超时时间（如 30 秒）
   - 超时后返回已完成的结果

### 前端错误处理

1. **API 调用失败**
   - 显示友好的错误提示
   - 提供重试选项
   - 记录错误日志

2. **部分成功场景**
   - 显示操作结果摘要
   - 高亮失败的实例
   - 提供查看详细错误的选项

3. **WebSocket 断开**
   - 自动重连
   - 降级到轮询模式
   - 显示连接状态提示

## Testing Strategy

### 单元测试

#### 后端单元测试

1. **聚合逻辑测试**
   - 测试空节点列表
   - 测试单节点多进程
   - 测试多节点同名进程
   - 测试节点不可用场景

2. **批量操作测试**
   - 测试全部成功场景
   - 测试部分失败场景
   - 测试全部失败场景
   - 测试超时处理

#### 前端单元测试

1. **组件渲染测试**
   - 测试空状态显示
   - 测试进程列表渲染
   - 测试展开/收起功能

2. **搜索过滤测试**
   - 测试搜索匹配
   - 测试空结果
   - 测试大小写不敏感

### 属性测试

使用 Go 的 `testing/quick` 包进行属性测试：

1. **Property 1: 聚合完整性**
   - 生成随机节点和进程数据
   - 验证聚合结果包含所有进程且无重复

2. **Property 2: 实例计数一致性**
   - 生成随机进程实例
   - 验证计数字段的数学关系

3. **Property 3: 批量操作覆盖性**
   - 生成随机操作结果
   - 验证成功和失败计数之和等于总数

4. **Property 4: 搜索过滤正确性**
   - 生成随机进程名称和搜索字符串
   - 验证过滤结果的正确性

5. **Property 5: 节点故障隔离性**
   - 模拟随机节点故障
   - 验证操作继续且结果正确标记

### 集成测试

1. **端到端流程测试**
   - 启动测试 Supervisor 实例
   - 测试完整的聚合和操作流程
   - 验证 WebSocket 实时更新

2. **并发测试**
   - 模拟多用户同时操作
   - 验证数据一致性
   - 测试性能表现

## Implementation Notes

### 后端实现要点

1. **并行处理**
   ```go
   var wg sync.WaitGroup
   results := make(chan InstanceOperationResult, len(instances))
   
   for _, instance := range instances {
       wg.Add(1)
       go func(inst ProcessInstance) {
           defer wg.Done()
           // 执行操作
           results <- result
       }(instance)
   }
   
   wg.Wait()
   close(results)
   ```

2. **超时控制**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   ```

3. **错误聚合**
   - 收集所有操作结果
   - 区分成功和失败
   - 提供详细的错误信息

### 前端实现要点

1. **状态管理**
   - 使用 useState 管理进程列表
   - 使用 useState 管理展开状态
   - 使用 useState 管理加载状态

2. **搜索优化**
   - 使用 useMemo 缓存过滤结果
   - 防抖处理搜索输入

3. **WebSocket 集成**
   - 复用现有 WebSocket 连接
   - 监听进程状态变化事件
   - 更新对应的聚合数据

### 路由配置

后端路由：
```go
processes := api.Group("/processes")
{
    processes.GET("/aggregated", processesAPI.GetAggregatedProcesses)
    processes.POST("/:process_name/start", processesAPI.BatchStartProcess)
    processes.POST("/:process_name/stop", processesAPI.BatchStopProcess)
    processes.POST("/:process_name/restart", processesAPI.BatchRestartProcess)
}
```

前端路由：
```typescript
{
  path: '/processes',
  element: <ProcessAggregationPage />
}
```

导航菜单：
```typescript
{
  key: 'processes',
  icon: <AppstoreOutlined />,
  label: 'Processes',
  path: '/processes'
}
```
