# Implementation Plan

- [x] 1. 创建 API 客户端和类型定义
  - 在 `src/api/` 目录创建 `environments.ts`，实现 `getEnvironments()` 和 `getEnvironmentDetail(name)` 方法
  - 在 `src/types/index.ts` 添加 `Environment`, `EnvironmentDetail`, `NodeSummary`, `NodeDetail` 类型定义
  - _Requirements: 1.1, 2.2_

- [ ] 2. 实现环境列表页面
- [x] 2.1 创建 EnvironmentList 页面组件
  - 在 `src/pages/Environments/` 创建 `index.tsx`
  - 实现环境列表加载逻辑（调用 `environmentsApi.getEnvironments()`）
  - 实现加载状态、错误状态和空状态的 UI
  - 集成 WebSocket 实时更新（复用 `useWebSocket` hook）
  - _Requirements: 1.1, 1.3, 1.4, 4.1, 4.2, 4.3_

- [x] 2.2 创建 EnvironmentCard 组件
  - 在 `src/pages/Environments/` 创建 `EnvironmentCard.tsx`
  - 实现卡片布局，显示环境名称、节点总数、在线/离线节点数
  - 实现点击导航到环境详情页
  - 使用 Ant Design Card 和 Tag 组件
  - _Requirements: 1.2, 2.1, 5.1, 5.3_

- [ ] 2.3 编写环境列表页面的属性测试
  - **Property 1: Environment list renders all required fields**
  - **Validates: Requirements 1.2**

- [ ] 2.4 编写环境列表页面的属性测试
  - **Property 2: Navigation to environment detail**
  - **Validates: Requirements 2.1**

- [ ] 2.5 编写环境列表页面的属性测试
  - **Property 6: WebSocket subscription on mount**
  - **Validates: Requirements 4.1**

- [ ] 2.6 编写环境列表页面的属性测试
  - **Property 8: WebSocket cleanup on unmount**
  - **Validates: Requirements 4.3**

- [ ] 2.7 编写环境列表页面的属性测试
  - **Property 10: Loading state display**
  - **Validates: Requirements 5.4**

- [ ] 3. 实现环境详情页面
- [x] 3.1 创建 EnvironmentDetail 页面组件
  - 在 `src/pages/Environments/` 创建 `EnvironmentDetail.tsx`
  - 从 URL 参数获取环境名称
  - 实现环境详情加载逻辑（调用 `environmentsApi.getEnvironmentDetail(name)`）
  - 实现 404 错误处理（环境不存在时）
  - 实现节点列表渲染，显示节点名称、主机、端口、连接状态、最后心跳时间
  - 实现节点点击导航到节点详情页
  - 集成 WebSocket 实时更新
  - _Requirements: 2.2, 2.3, 2.4, 2.5, 3.1, 4.1, 4.2, 4.3_

- [x] 3.2 实现节点状态标签组件
  - 在 EnvironmentDetail 中实现节点状态标签
  - 在线节点显示绿色标签，离线节点显示红色标签
  - _Requirements: 5.2_

- [ ] 3.3 编写环境详情页面的属性测试
  - **Property 3: Environment detail renders environment name and nodes**
  - **Validates: Requirements 2.2**

- [ ] 3.4 编写环境详情页面的属性测试
  - **Property 4: Node list renders all required fields**
  - **Validates: Requirements 2.3**

- [ ] 3.5 编写环境详情页面的属性测试
  - **Property 5: Navigation to node detail**
  - **Validates: Requirements 3.1**

- [ ] 3.6 编写环境详情页面的属性测试
  - **Property 7: Real-time status update**
  - **Validates: Requirements 4.2**

- [ ] 3.7 编写环境详情页面的属性测试
  - **Property 9: Status tag color mapping**
  - **Validates: Requirements 5.2**

- [ ] 4. 集成路由和导航
- [x] 4.1 更新 App.tsx 添加环境路由
  - 在 `src/App.tsx` 中添加 `/environments` 和 `/environments/:environmentName` 路由
  - 确保路由在 ProtectedRoute 保护下
  - _Requirements: 1.1, 2.1_

- [x] 4.2 更新主布局导航菜单
  - 在 `src/layouts/MainLayout.tsx` 的侧边栏菜单中添加"Environments"菜单项
  - 使用合适的图标（如 AppstoreOutlined）
  - _Requirements: 1.1_

- [x] 5. 最终检查点
  - 确保所有测试通过，如有问题请询问用户
