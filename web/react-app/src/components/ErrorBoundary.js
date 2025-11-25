import React from 'react';
import errorHandler from '../services/errorHandler';

/**
 * 全局错误边界组件
 * 捕获组件渲染错误并显示友好的错误页面
 * 验证需求：11.4
 */
class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    };
  }

  static getDerivedStateFromError(error) {
    // 更新 state 使下一次渲染能够显示降级后的 UI
    return { hasError: true };
  }

  componentDidCatch(error, errorInfo) {
    // 记录错误到错误处理服务
    errorHandler.handleError(error, {
      type: 'component_error',
      componentStack: errorInfo.componentStack,
    });

    // 更新状态
    this.setState({
      error,
      errorInfo,
    });
  }

  handleReset = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    });
  };

  render() {
    if (this.state.hasError) {
      // 自定义降级 UI
      return (
        <div className="container mt-5">
          <div className="row justify-content-center">
            <div className="col-md-8">
              <div className="card border-danger">
                <div className="card-header bg-danger text-white">
                  <h4 className="mb-0">
                    <i className="bi bi-exclamation-triangle-fill me-2"></i>
                    Something went wrong
                  </h4>
                </div>
                <div className="card-body">
                  <p className="card-text">
                    We're sorry, but something unexpected happened. The error has been logged and we'll look into it.
                  </p>

                  {process.env.NODE_ENV === 'development' && this.state.error && (
                    <div className="mt-3">
                      <h5>Error Details (Development Only):</h5>
                      <div className="alert alert-warning">
                        <strong>Error:</strong> {this.state.error.toString()}
                      </div>
                      {this.state.errorInfo && (
                        <details className="mt-2">
                          <summary className="btn btn-sm btn-outline-secondary">
                            Show Component Stack
                          </summary>
                          <pre className="mt-2 p-2 bg-light border rounded" style={{ fontSize: '0.85rem' }}>
                            {this.state.errorInfo.componentStack}
                          </pre>
                        </details>
                      )}
                    </div>
                  )}

                  <div className="mt-4">
                    <button
                      className="btn btn-primary me-2"
                      onClick={this.handleReset}
                    >
                      <i className="bi bi-arrow-clockwise me-1"></i>
                      Try Again
                    </button>
                    <button
                      className="btn btn-outline-secondary"
                      onClick={() => (window.location.href = '/')}
                    >
                      <i className="bi bi-house-door me-1"></i>
                      Go to Home
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
