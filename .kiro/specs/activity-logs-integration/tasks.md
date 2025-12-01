# Implementation Plan

- [x] 1. 创建前端 API 客户端和类型定义
  - 创建 `web/react-app/src/api/activityLogs.ts` 文件
  - 定义 ActivityLog 接口和相关类型
  - 实现 getActivityLogs、getRecentLogs、getLogStatistics、exportLogs 方法
  - 使用现有的 apiClient 进行 HTTP 请求
  - _Requirements: 1.1, 2.1_

- [x] 1.1 编写 API 客户端的单元测试
  - 创建 `web/react-app/src/api/__tests__/activityLogs.test.ts`
  - 测试 API 方法正确构造请求
  - 测试响应正确解析
  - 测试错误正确处理
  - _Requirements: 1.1_

- [x] 2. 更新前端类型定义
  - 在 `web/react-app/src/types/index.ts` 中添加 ActivityLog 类型
  - 添加 ActivityLogsFilters 类型
  - 添加 PaginationInfo 类型
  - 确保类型与后端模型一致
  - _Requirements: 1.2_

- [x] 3. 重构 Logs 页面组件移除 mock 数据
  - 删除 `web/react-app/src/pages/Logs/index.tsx` 中的所有 mock 数据
  - 删除 mockSystemLogs 和 mockActivityLogs 数组
  - 删除 mock 数据生成逻辑
  - 使用 activityLogsAPI 替换 mock 数据
  - 实现真实的 loadLogs 函数调用后端 API
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [x] 4. 实现筛选和搜索功能
  - 实现搜索框的 API 参数传递
  - 实现时间范围筛选
  - 实现用户名筛选
  - 实现操作类型筛选
  - 实现资源类型筛选
  - 确保筛选条件正确传递到后端
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

- [x] 5. 实现分页功能
  - 使用后端返回的分页元数据
  - 实现页码切换
  - 实现每页大小调整
  - 显示总记录数和当前页信息
  - 处理分页状态管理
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 6. 实现错误处理和加载状态
  - 添加 loading 状态显示
  - 实现 API 错误处理
  - 显示用户友好的错误消息
  - 添加重试机制
  - 确保不回退到 mock 数据
  - _Requirements: 4.3_

- [x] 6.1 编写 Logs 页面组件的单元测试
  - 创建 `web/react-app/src/pages/Logs/__tests__/index.test.tsx`
  - 测试组件正确渲染
  - 测试 API 调用正确触发
  - 测试筛选器正确应用
  - 测试分页控件正确工作
  - 测试错误状态正确显示
  - _Requirements: 1.1, 2.1, 3.1_

- [x] 7. 后端添加日志导出功能
  - 在 `internal/api/activity_logs.go` 中添加 ExportLogs 端点
  - 实现 CSV 格式导出
  - 支持筛选条件
  - 实现流式输出处理大文件
  - 在 `internal/api/api.go` 中注册路由
  - _Requirements: 7.1, 7.2, 7.3, 7.4_

- [x] 8. 后端添加系统事件记录功能
  - 在 `internal/services/activity_log.go` 中添加 LogSystemEvent 方法
  - 实现进程状态变化记录
  - 实现节点连接状态记录
  - 确保系统事件标记为 "system" 来源
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

- [x] 9. 集成进程操作日志记录
  - 在 `internal/api/processes.go` 的 StartProcess 中添加日志记录
  - 在 StopProcess 中添加日志记录
  - 在 RestartProcess 中添加日志记录
  - 在批量操作中添加日志记录
  - 确保记录包含用户名、节点名、进程名、时间戳
  - _Requirements: 1.3_

- [x] 10. 集成系统事件监控
  - 在 `internal/supervisor/service.go` 中添加状态变化监听
  - 监听进程状态变化事件
  - 监听节点连接状态变化
  - 调用 Activity Log Service 记录事件
  - _Requirements: 1.4, 5.1, 5.2, 5.3, 5.4_

- [x] 10.1 编写后端 API 的单元测试
  - 创建或更新 `internal/api/activity_logs_test.go`
  - 测试各端点正确响应
  - 测试参数验证
  - 测试分页逻辑
  - 测试筛选逻辑
  - 测试导出功能
  - _Requirements: 2.1, 3.1, 7.1_

- [x] 10.2 编写 Activity Log Service 的单元测试
  - 创建或更新 `internal/services/activity_log_test.go`
  - 测试日志记录功能
  - 测试查询功能
  - 测试统计功能
  - 测试导出功能
  - 测试系统事件记录
  - _Requirements: 1.3, 1.4, 5.1_

- [x] 11. 前端实现自动刷新功能
  - 使用 setInterval 实现轮询（30 秒间隔）
  - 仅在第一页时自动刷新
  - 添加自动刷新开关
  - 组件卸载时清理定时器
  - 避免在用户浏览历史时自动滚动
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [x] 12. 前端实现导出功能
  - 添加导出按钮点击处理
  - 调用后端导出 API
  - 使用 Blob 和 URL.createObjectURL 触发下载
  - 显示导出进度提示
  - 处理导出错误
  - _Requirements: 7.1, 7.3, 7.5_

- [x] 13. 数据库优化
  - 为 activity_logs 表的 created_at 字段添加索引
  - 为 username 字段添加索引
  - 为 action 字段添加索引
  - 验证索引创建成功
  - _Requirements: 性能优化_

- [x] 14. 编写属性测试
  - **Property 1: API Response Consistency**
  - **Validates: Requirements 1.1, 3.5**

- [x] 14.1 编写属性测试
  - **Property 2: Filter Application Correctness**
  - **Validates: Requirements 2.1, 2.2, 2.3, 2.4, 2.5**

- [x] 14.2 编写属性测试
  - **Property 3: Pagination Consistency**
  - **Validates: Requirements 3.1, 3.2, 3.3, 3.4**

- [x] 14.3 编写属性测试
  - **Property 4: Mock Data Absence**
  - **Validates: Requirements 4.1, 4.2, 4.3, 4.4, 4.5**

- [x] 14.4 编写属性测试
  - **Property 5: User Action Logging Completeness**
  - **Validates: Requirements 1.3**

- [x] 14.5 编写属性测试
  - **Property 6: System Event Logging Completeness**
  - **Validates: Requirements 1.4, 5.1, 5.2, 5.3, 5.4, 5.5**

- [x] 14.6 编写属性测试
  - **Property 7: Real-time Update Consistency**
  - **Validates: Requirements 6.1, 6.2, 6.4**

- [x] 14.7 编写属性测试
  - **Property 8: Export Data Completeness**
  - **Validates: Requirements 7.1, 7.2, 7.3, 7.4**

- [x] 15. Checkpoint - 确保所有测试通过
  - 运行所有前端测试
  - 运行所有后端测试
  - 修复任何失败的测试
  - 确保代码覆盖率达标
  - 如有问题请询问用户

- [x] 16. 手动测试和验证
  - 访问 Activity Logs 页面验证数据加载
  - 测试各种筛选条件组合
  - 测试分页功能
  - 执行进程操作验证日志记录
  - 测试导出功能
  - 测试自动刷新功能
  - 测试错误场景
  - _Requirements: 所有需求_
