# Design Document: Process Operation Error Feedback Fix

## Overview

修复 SupervisorClient 中 XML-RPC 响应解析的 bug，确保进程操作（启动/停止/重启）能够正确报告成功或失败。

当前问题：`Client.Call` 返回原始 XML 字符串，但 `StartProcess` 等方法尝试将其断言为 Go 原生类型（如 `bool`），导致类型断言失败，成功的操作被错误地报告为失败。

## Architecture

### 当前架构问题

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  API Handler    │────▶│ SupervisorClient│────▶│   XML-RPC       │
│  (processes.go) │     │ (supervisor.go) │     │   Client        │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                               │
                               ▼
                        result.(bool) ❌
                        类型断言失败！
```

### 修复后架构

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  API Handler    │────▶│ SupervisorClient│────▶│   XML-RPC       │
│  (processes.go) │     │ (supervisor.go) │     │   Client        │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                               │
                               ▼
                        ┌─────────────────┐
                        │ Response Parser │
                        │ - parseBoolResp │
                        │ - parseFaultResp│
                        └─────────────────┘
```

## Components and Interfaces

### 1. XML-RPC Response Parser Functions

新增统一的响应解析函数，位于 `internal/supervisor/xmlrpc/supervisor.go`：

```go
// parseBooleanResponse 解析 XML-RPC 布尔响应
// 返回: (success bool, isFault bool, faultMsg string, err error)
func parseBooleanResponse(xmlResponse string) (bool, bool, string, error)

// parseFaultResponse 解析 XML-RPC Fault 响应
// 返回: (faultCode int, faultString string, isFault bool)
func parseFaultResponse(xmlResponse string) (int, string, bool)
```

### 2. 修改后的 StartProcess 方法

```go
func (s *SupervisorClient) StartProcess(name string) error {
    result, err := s.client.Call("supervisor.startProcess", []interface{}{name})
    if err != nil {
        return err
    }

    xmlResponse, ok := result.(string)
    if !ok {
        return fmt.Errorf("unexpected response type: %T", result)
    }

    // 检查是否是 fault 响应
    if faultCode, faultString, isFault := parseFaultResponse(xmlResponse); isFault {
        // ALREADY_STARTED 视为成功（幂等操作）
        if strings.Contains(faultString, "ALREADY_STARTED") {
            return nil
        }
        return fmt.Errorf("supervisor fault [%d]: %s", faultCode, faultString)
    }

    // 解析布尔响应
    success, _, _, err := parseBooleanResponse(xmlResponse)
    if err != nil {
        return err
    }

    if !success {
        return fmt.Errorf("supervisor rejected start request for process %s", name)
    }

    return nil
}
```

### 3. 修改后的 StopProcess 方法

```go
func (s *SupervisorClient) StopProcess(name string) error {
    // ... 现有的状态检查逻辑 ...

    result, err := s.client.Call("supervisor.stopProcess", []interface{}{name})
    if err != nil {
        if strings.Contains(err.Error(), "NOT_RUNNING") {
            return nil // 幂等操作
        }
        return err
    }

    xmlResponse, ok := result.(string)
    if !ok {
        return fmt.Errorf("unexpected response type: %T", result)
    }

    // 检查是否是 fault 响应
    if faultCode, faultString, isFault := parseFaultResponse(xmlResponse); isFault {
        // NOT_RUNNING 视为成功（幂等操作）
        if strings.Contains(faultString, "NOT_RUNNING") {
            return nil
        }
        return fmt.Errorf("supervisor fault [%d]: %s", faultCode, faultString)
    }

    // 解析布尔响应
    success, _, _, err := parseBooleanResponse(xmlResponse)
    if err != nil {
        return err
    }

    if !success {
        return fmt.Errorf("supervisor rejected stop request for process %s", name)
    }

    return nil
}
```

## Data Models

### XML-RPC 响应格式

#### 成功响应（布尔值）

```xml
<?xml version="1.0"?>
<methodResponse>
  <params>
    <param>
      <value><boolean>1</boolean></value>
    </param>
  </params>
</methodResponse>
```

#### Fault 响应

```xml
<?xml version="1.0"?>
<methodResponse>
  <fault>
    <value>
      <struct>
        <member>
          <name>faultCode</name>
          <value><int>60</int></value>
        </member>
        <member>
          <name>faultString</name>
          <value><string>ALREADY_STARTED: process iptv-feedback-success</string></value>
        </member>
      </struct>
    </value>
  </fault>
</methodResponse>
```

### Supervisor Fault Codes

| Code | Name | Description |
|------|------|-------------|
| 10 | BAD_NAME | 进程名不存在 |
| 60 | ALREADY_STARTED | 进程已经在运行 |
| 70 | NOT_RUNNING | 进程未在运行 |
| 80 | SPAWN_ERROR | 进程启动失败 |
| 90 | ABNORMAL_TERMINATION | 进程异常终止 |



## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Boolean Response Parsing Round-Trip

*For any* valid XML-RPC boolean response (containing `<boolean>0</boolean>` or `<boolean>1</boolean>`), the `parseBooleanResponse` function SHALL correctly extract the boolean value and return it without error.

**Validates: Requirements 1.1, 1.2**

### Property 2: Fault Response Parsing Completeness

*For any* valid XML-RPC fault response containing faultCode and faultString, the `parseFaultResponse` function SHALL extract both values correctly, and the returned error message SHALL contain both the code and the string.

**Validates: Requirements 1.3, 2.1, 2.2, 2.3**

### Property 3: Malformed XML Graceful Handling

*For any* malformed or incomplete XML string, the parsing functions SHALL return a descriptive error instead of panicking or returning incorrect results.

**Validates: Requirements 3.3**

### Property 4: StartProcess XML Response Handling

*For any* XML response from Supervisor (success, failure, or fault), the `StartProcess` method SHALL correctly interpret the response and return the appropriate result (nil for success, error for failure).

**Validates: Requirements 1.4**

## Error Handling

### Error Types

1. **Network Errors**: Connection failures, timeouts - handled by existing `Client.Call`
2. **XML Parse Errors**: Malformed responses - return descriptive error with raw XML snippet
3. **Supervisor Faults**: Operation rejected by Supervisor - return fault code and message
4. **Operation Rejected**: Boolean false response - return descriptive error

### Idempotent Operations

- `StartProcess`: ALREADY_STARTED fault → return nil (success)
- `StopProcess`: NOT_RUNNING fault → return nil (success)

This ensures operations are safe to retry without causing false errors.

## Testing Strategy

### Unit Tests

1. Test `parseBooleanResponse` with valid boolean XML
2. Test `parseFaultResponse` with valid fault XML
3. Test parsing functions with malformed XML
4. Test `StartProcess` with mock responses

### Property-Based Tests

使用 Go 的 `testing/quick` 包进行属性测试：

1. **Property 1**: Generate random boolean values, construct XML, parse, verify round-trip
2. **Property 2**: Generate random faultCode/faultString pairs, construct XML, parse, verify extraction
3. **Property 3**: Generate random malformed strings, verify graceful error handling
4. **Property 4**: Generate various XML response types, verify correct interpretation

### Test Configuration

- Minimum 100 iterations per property test
- Each test tagged with: **Feature: process-operation-error-feedback, Property N: [description]**
