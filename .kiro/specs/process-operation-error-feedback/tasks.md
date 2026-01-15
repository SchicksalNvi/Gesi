# Implementation Plan: Process Operation Error Feedback Fix

## Overview

修复 SupervisorClient 中 XML-RPC 响应解析的 bug，实现正确的布尔响应和 Fault 响应解析。

## Tasks

- [x] 1. 实现 XML-RPC 响应解析函数
  - [x] 1.1 实现 parseBooleanResponse 函数
    - 在 `internal/supervisor/xmlrpc/supervisor.go` 中添加函数
    - 从 XML 中提取 `<boolean>` 标签的值
    - 返回 (success bool, err error)
    - _Requirements: 1.1, 1.2_

  - [x] 1.2 实现 parseFaultResponse 函数
    - 在 `internal/supervisor/xmlrpc/supervisor.go` 中添加函数
    - 从 XML 中提取 faultCode 和 faultString
    - 返回 (faultCode int, faultString string, isFault bool)
    - _Requirements: 2.1, 2.2, 2.3_

  - [x] 1.3 编写解析函数的单元测试
    - 测试有效的布尔响应解析
    - 测试有效的 Fault 响应解析
    - 测试畸形 XML 的错误处理
    - _Requirements: 3.3_

- [x] 2. 修复 StartProcess 方法
  - [x] 2.1 修改 StartProcess 使用新的解析函数
    - 使用 parseBooleanResponse 解析成功响应
    - 使用 parseFaultResponse 检测和处理 Fault
    - 处理 ALREADY_STARTED 为成功（幂等）
    - _Requirements: 1.4, 2.4_

  - [x] 2.2 编写 StartProcess 的属性测试
    - **Property 4: StartProcess XML Response Handling**
    - **Validates: Requirements 1.4**

- [x] 3. 修复 StopProcess 方法
  - [x] 3.1 修改 StopProcess 使用新的解析函数
    - 使用 parseBooleanResponse 解析成功响应
    - 使用 parseFaultResponse 检测和处理 Fault
    - 处理 NOT_RUNNING 为成功（幂等）
    - _Requirements: 2.5_

- [x] 4. Checkpoint - 验证修复
  - 运行现有测试确保无回归
  - 手动测试进程启动/停止操作
  - 确保成功操作正确报告为成功

- [x] 5. 编写属性测试
  - [x] 5.1 编写布尔响应解析属性测试
    - **Property 1: Boolean Response Parsing Round-Trip**
    - **Validates: Requirements 1.1, 1.2**

  - [x] 5.2 编写 Fault 响应解析属性测试
    - **Property 2: Fault Response Parsing Completeness**
    - **Validates: Requirements 1.3, 2.1, 2.2, 2.3**

  - [x] 5.3 编写畸形 XML 处理属性测试
    - **Property 3: Malformed XML Graceful Handling**
    - **Validates: Requirements 3.3**

- [x] 6. Final Checkpoint
  - 确保所有测试通过
  - 验证批量操作正确报告结果

## Notes

- 所有任务都是必需的，包括属性测试
- 核心修复在 Task 1 和 Task 2，这是解决问题的关键
- Property tests 提供额外的正确性保证
