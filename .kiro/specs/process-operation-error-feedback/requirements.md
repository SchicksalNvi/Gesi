# Requirements Document

## Introduction

修复进程操作（启动/停止/重启）的结果解析问题。当前 SupervisorClient 的 StartProcess 方法无法正确解析 XML-RPC 响应，导致成功的操作被错误地报告为失败。

根本原因：`Client.Call` 返回原始 XML 字符串，但 `StartProcess` 方法尝试将其断言为 `bool` 类型，这永远会失败。

## Glossary

- **SupervisorClient**: XML-RPC 客户端，与 Supervisor 进程管理器通信
- **XML-RPC**: 远程过程调用协议，使用 XML 编码请求和响应
- **Fault_Response**: Supervisor 返回的错误响应，包含 faultCode 和 faultString

## Requirements

### Requirement 1: 正确解析 XML-RPC 布尔响应

**User Story:** As a system administrator, I want process operations to correctly report success or failure, so that I can trust the operation results.

#### Acceptance Criteria

1. WHEN Supervisor returns a successful response with `<boolean>1</boolean>`, THE SupervisorClient SHALL return nil (success)
2. WHEN Supervisor returns a successful response with `<boolean>0</boolean>`, THE SupervisorClient SHALL return an error indicating the operation was rejected
3. WHEN Supervisor returns a fault response, THE SupervisorClient SHALL extract the faultString and return it as an error
4. THE StartProcess method SHALL correctly parse the XML response instead of expecting a Go bool type

### Requirement 2: 正确解析 XML-RPC Fault 响应

**User Story:** As a system administrator, I want to see the actual error message when Supervisor rejects an operation, so that I can diagnose the problem.

#### Acceptance Criteria

1. WHEN Supervisor returns a fault response, THE SupervisorClient SHALL extract the faultCode from the XML
2. WHEN Supervisor returns a fault response, THE SupervisorClient SHALL extract the faultString from the XML
3. THE error message SHALL include both the faultCode and faultString for debugging
4. IF the fault indicates "ALREADY_STARTED", THE StartProcess method SHALL treat it as success (idempotent operation)
5. IF the fault indicates "NOT_RUNNING", THE StopProcess method SHALL treat it as success (idempotent operation)

### Requirement 3: 统一的响应解析函数

**User Story:** As a developer, I want a consistent way to parse XML-RPC responses, so that all operations handle responses correctly.

#### Acceptance Criteria

1. THE SupervisorClient SHALL have a unified function to parse boolean responses from XML
2. THE SupervisorClient SHALL have a unified function to detect and parse fault responses
3. THE parsing functions SHALL handle malformed XML gracefully and return descriptive errors
4. THE parsing functions SHALL be used by StartProcess, StopProcess, and other operation methods
