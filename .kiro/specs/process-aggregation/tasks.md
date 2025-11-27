# Implementation Plan

- [x] 1. 后端 API 实现
  - 创建 `internal/api/processes.go` 文件
  - 实现进程聚合逻辑
  - 实现批量操作接口
  - _Requirements: 1.1, 1.2, 1.3, 3.1, 3.2, 3.3, 3.4_

- [x] 1.1 实现 ProcessesAPI 结构和聚合接口
  - 创建 ProcessesAPI 结构体
  - 实现 GetAggregatedProcesses 方法
  - 从所有节点收集进程信息
  - 按进程名称聚合实例
  - 计算统计信息（总数、运行数、停止数）
  - _Requirements: 1.1, 1.2, 1.3, 1.4_

- [x] 1.2 实现批量启动接口
  - 实现 BatchStartProcess 方法
  - 并行调用所有节点的 StartProcess
  - 收集操作结果
  - 返回操作摘要
  - _Requirements: 3.1, 3.4, 3.5_

- [x] 1.3 实现批量停止接口
  - 实现 BatchStopProcess 方法
  - 并行调用所有节点的 StopProcess
  - 处理节点不可用场景
  - _Requirements: 3.2, 3.4, 3.5, 6.1, 6.2_

- [x] 1.4 实现批量重启接口
  - 实现 BatchRestartProcess 方法
  - 并行调用所有节点的 RestartProcess
  - 添加超时控制
  - _Requirements: 3.3, 3.4, 3.5_

- [x] 1.5 编写属性测试 - 聚合完整性
  - **Property 1: 聚合完整性**
  - **Validates: Requirements 1.1, 1.2**
  - 生成随机节点和进程数据
  - 验证聚合结果包含所有进程且无重复

- [x] 1.6 编写属性测试 - 实例计数一致性
  - **Property 2: 实例计数一致性**
  - **Validates: Requirements 1.3**
  - 生成随机进程实例
  - 验证 total = running + stopped = len(instances)

- [x] 1.7 编写属性测试 - 批量操作覆盖性
  - **Property 3: 批量操作覆盖性**
  - **Validates: Requirements 3.1, 3.2, 3.3, 3.4**
  - 生成随机操作结果
  - 验证 success + failure = total

- [x] 1.8 编写属性测试 - 节点故障隔离性
  - **Property 5: 节点故障隔离性**
  - **Validates: Requirements 6.1, 6.2**
  - 模拟随机节点故障
  - 验证操作继续且结果正确标记

- [x] 2. 路由注册
  - 在 `internal/api/api.go` 中注册新路由
  - 添加认证中间件
  - 添加输入验证
  - _Requirements: 1.1, 3.1, 3.2, 3.3_

- [x] 2.1 注册进程聚合路由
  - 创建 /api/processes 路由组
  - 注册 GET /api/processes/aggregated
  - 注册 POST /api/processes/:process_name/start
  - 注册 POST /api/processes/:process_name/stop
  - 注册 POST /api/processes/:process_name/restart
  - 添加认证中间件
  - _Requirements: 1.1, 3.1, 3.2, 3.3_

- [x] 2.2 添加输入验证
  - 验证进程名称格式
  - 防止 SQL 注入
  - 清理输入数据
  - _Requirements: 3.1, 3.2, 3.3_

- [x] 3. 前端类型定义
  - 在 `web/react-app/src/types/index.ts` 中添加新类型
  - 定义 AggregatedProcess 接口
  - 定义 ProcessInstance 接口
  - 定义 BatchOperationResult 接口
  - _Requirements: 1.3, 3.5_

- [x] 4. 前端 API 客户端
  - 创建 `web/react-app/src/api/processes.ts`
  - 实现 getAggregated 方法
  - 实现 batchStart 方法
  - 实现 batchStop 方法
  - 实现 batchRestart 方法
  - _Requirements: 1.1, 3.1, 3.2, 3.3_

- [x] 5. 进程聚合页面组件
  - 创建 `web/react-app/src/pages/Processes/index.tsx`
  - 实现进程列表显示
  - 实现搜索功能
  - 实现批量操作按钮
  - 实现展开/收起功能
  - _Requirements: 1.1, 1.3, 2.1, 2.2, 2.3, 3.1, 3.2, 3.3, 4.1_

- [x] 5.1 实现基础页面结构
  - 创建页面组件框架
  - 添加搜索框
  - 添加进程列表容器
  - 添加加载状态
  - _Requirements: 1.1, 2.1_

- [x] 5.2 实现进程列表渲染
  - 显示进程名称
  - 显示实例统计（总数、运行数、停止数）
  - 显示节点列表
  - 添加展开/收起按钮
  - _Requirements: 1.3, 1.4, 4.1_

- [x] 5.3 实现搜索和过滤
  - 实现搜索输入处理
  - 实现实时过滤逻辑
  - 添加搜索结果高亮
  - 处理空结果状态
  - _Requirements: 2.1, 2.2, 2.3, 2.4_

- [x] 5.4 编写属性测试 - 搜索过滤正确性
  - **Property 4: 搜索过滤正确性**
  - **Validates: Requirements 2.1**
  - 生成随机进程名称和搜索字符串
  - 验证过滤结果的正确性

- [x] 5.5 实现批量操作功能
  - 添加批量启动按钮
  - 添加批量停止按钮
  - 添加批量重启按钮
  - 实现操作确认对话框
  - 显示操作结果摘要
  - _Requirements: 3.1, 3.2, 3.3, 3.5_

- [x] 6. 进程实例详情组件
  - 创建 `web/react-app/src/pages/Processes/ProcessInstanceList.tsx`
  - 显示实例详情表格
  - 实现单实例操作按钮
  - 实现日志查看功能
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 5.1, 5.2, 5.3, 5.4_

- [x] 6.1 实现实例列表表格
  - 显示节点名称和主机
  - 显示进程状态
  - 显示 PID 和运行时间
  - 显示描述信息
  - _Requirements: 4.2_

- [x] 6.2 实现单实例操作
  - 添加启动按钮
  - 添加停止按钮
  - 添加重启按钮
  - 处理操作加载状态
  - _Requirements: 4.3_

- [x] 6.3 实现日志查看功能
  - 添加日志按钮
  - 复用现有日志 Modal 组件
  - 显示节点和进程信息
  - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [x] 7. 路由和导航集成
  - 在 `web/react-app/src/App.tsx` 中添加路由
  - 在导航菜单中添加入口
  - 添加权限控制
  - _Requirements: 7.1, 7.2, 7.3, 7.4_

- [x] 7.1 添加前端路由
  - 在 App.tsx 中添加 /processes 路由
  - 配置路由保护
  - _Requirements: 7.1, 7.2_

- [x] 7.2 更新导航菜单
  - 在主导航中添加 Processes 菜单项
  - 添加图标
  - 配置高亮逻辑
  - _Requirements: 7.1, 7.2, 7.3_

- [x] 7.3 添加权限控制
  - 检查用户权限
  - 隐藏无权限用户的菜单项
  - _Requirements: 7.4_

- [x] 8. WebSocket 实时更新
  - 在进程聚合页面中集成 WebSocket
  - 监听进程状态变化事件
  - 更新聚合数据
  - _Requirements: 1.5, 6.3_

- [x] 8.1 集成 WebSocket 连接
  - 复用现有 useWebSocket hook
  - 订阅进程状态变化事件
  - _Requirements: 1.5_

- [x] 8.2 实现状态更新逻辑
  - 处理进程状态变化消息
  - 更新对应的聚合数据
  - 更新实例计数
  - _Requirements: 1.5, 6.3_

- [x] 8.3 编写属性测试 - 实时更新一致性
  - **Property 6: 实时更新一致性**
  - **Validates: Requirements 1.5**
  - 模拟进程状态变化
  - 验证更新在 5 秒内反映

- [x] 9. 错误处理和用户体验优化
  - 添加错误提示
  - 添加加载状态
  - 添加空状态显示
  - 优化操作反馈
  - _Requirements: 2.2, 4.4, 6.1, 6.2, 6.4_

- [x] 9.1 实现错误处理
  - 捕获 API 错误
  - 显示友好的错误消息
  - 提供重试选项
  - _Requirements: 4.4, 6.4_

- [x] 9.2 优化加载状态
  - 添加骨架屏
  - 添加操作加载指示器
  - 防止重复提交
  - _Requirements: 1.1, 3.1, 3.2, 3.3_

- [x] 9.3 实现空状态显示
  - 无进程时的提示
  - 搜索无结果的提示
  - 节点全部不可用的提示
  - _Requirements: 2.2, 6.1_

- [x] 10. Checkpoint - 确保所有测试通过
  - 确保所有测试通过，如有问题请询问用户
