/**
 * 统一的错误处理服务
 * 用于处理应用中的所有错误，提供一致的错误处理和用户反馈
 */

import { toast } from 'sonner';

class ErrorHandler {
  constructor() {
    this.isDevelopment = process.env.NODE_ENV === 'development';
  }

  /**
   * 处理错误
   * @param {Error|Object} error - 错误对象
   * @param {Object} options - 选项
   * @param {boolean} options.showToast - 是否显示 toast 通知
   * @param {string} options.context - 错误上下文信息
   * @param {Function} options.onError - 自定义错误处理回调
   */
  handleError(error, options = {}) {
    const {
      showToast = true,
      context = '',
      onError = null,
    } = options;

    // 提取错误信息
    const errorInfo = this.extractErrorInfo(error);

    // 记录错误（开发环境）
    if (this.isDevelopment) {
      console.group(`🔴 Error${context ? ` in ${context}` : ''}`);
      console.error('Error:', error);
      console.error('Error Info:', errorInfo);
      if (error.stack) {
        console.error('Stack:', error.stack);
      }
      console.groupEnd();
    } else {
      // 生产环境只记录简要信息
      console.error(`Error${context ? ` in ${context}` : ''}:`, errorInfo.message);
    }

    // 显示用户友好的错误消息
    if (showToast) {
      this.showErrorToast(errorInfo);
    }

    // 执行自定义错误处理
    if (onError && typeof onError === 'function') {
      onError(errorInfo);
    }

    // 返回错误信息供调用者使用
    return errorInfo;
  }

  /**
   * 提取错误信息
   * @param {Error|Object} error - 错误对象
   * @returns {Object} 标准化的错误信息
   */
  extractErrorInfo(error) {
    // 默认错误信息
    const defaultError = {
      code: 'UNKNOWN_ERROR',
      message: '发生未知错误',
      details: null,
      status: 500,
    };

    // 处理 null 或 undefined
    if (!error) {
      return defaultError;
    }

    // 处理 Axios 错误
    if (error.response) {
      const { status, data } = error.response;
      
      // 标准化的 API 错误响应
      if (data && data.error) {
        return {
          code: data.error.code || 'API_ERROR',
          message: data.error.message || '服务器错误',
          details: data.error.details || null,
          status: status,
        };
      }

      // 非标准化的错误响应
      return {
        code: 'API_ERROR',
        message: data.message || data.error || this.getStatusMessage(status),
        details: data,
        status: status,
      };
    }

    // 处理网络错误
    if (error.request) {
      return {
        code: 'NETWORK_ERROR',
        message: '网络连接失败，请检查您的网络连接',
        details: null,
        status: 0,
      };
    }

    // 处理标准 Error 对象
    if (error instanceof Error) {
      return {
        code: error.name || 'ERROR',
        message: error.message || '发生错误',
        details: error.stack || null,
        status: 500,
      };
    }

    // 处理字符串错误
    if (typeof error === 'string') {
      return {
        code: 'ERROR',
        message: error,
        details: null,
        status: 500,
      };
    }

    // 处理对象错误
    if (typeof error === 'object') {
      return {
        code: error.code || 'ERROR',
        message: error.message || '发生错误',
        details: error.details || error,
        status: error.status || 500,
      };
    }

    return defaultError;
  }

  /**
   * 显示错误 Toast
   * @param {Object} errorInfo - 错误信息
   */
  showErrorToast(errorInfo) {
    const { code, message, status } = errorInfo;

    // 根据错误类型显示不同的消息
    let toastMessage = message;
    let toastTitle = '错误';

    switch (code) {
      case 'UNAUTHORIZED':
        toastTitle = '未授权';
        toastMessage = '您需要登录才能访问此资源';
        break;
      case 'FORBIDDEN':
        toastTitle = '权限不足';
        toastMessage = '您没有权限执行此操作';
        break;
      case 'NOT_FOUND':
        toastTitle = '未找到';
        toastMessage = message || '请求的资源不存在';
        break;
      case 'VALIDATION_ERROR':
        toastTitle = '验证错误';
        break;
      case 'NETWORK_ERROR':
        toastTitle = '网络错误';
        break;
      case 'TIMEOUT':
        toastTitle = '请求超时';
        toastMessage = '请求超时，请稍后重试';
        break;
      default:
        if (status >= 500) {
          toastTitle = '服务器错误';
          toastMessage = '服务器遇到问题，请稍后重试';
        }
    }

    // 显示 toast
    toast.error(toastMessage, {
      description: this.isDevelopment ? `错误代码: ${code}` : undefined,
      duration: 5000,
    });
  }

  /**
   * 根据 HTTP 状态码获取默认消息
   * @param {number} status - HTTP 状态码
   * @returns {string} 状态消息
   */
  getStatusMessage(status) {
    const statusMessages = {
      400: '请求参数错误',
      401: '未授权，请先登录',
      403: '权限不足',
      404: '请求的资源不存在',
      408: '请求超时',
      409: '资源冲突',
      422: '请求参数验证失败',
      429: '请求过于频繁，请稍后重试',
      500: '服务器内部错误',
      502: '网关错误',
      503: '服务暂时不可用',
      504: '网关超时',
    };

    return statusMessages[status] || `请求失败 (${status})`;
  }

  /**
   * 处理 API 错误
   * @param {Error} error - API 错误
   * @param {Object} options - 选项
   */
  handleApiError(error, options = {}) {
    return this.handleError(error, {
      ...options,
      context: options.context || 'API Request',
    });
  }

  /**
   * 处理表单验证错误
   * @param {Object} errors - 验证错误对象
   * @param {Object} options - 选项
   */
  handleValidationErrors(errors, options = {}) {
    const { showToast = true } = options;

    if (!errors || typeof errors !== 'object') {
      return;
    }

    // 提取所有错误消息
    const errorMessages = Object.entries(errors)
      .map(([field, message]) => `${field}: ${message}`)
      .join('\n');

    if (showToast) {
      toast.error('表单验证失败', {
        description: errorMessages,
        duration: 5000,
      });
    }

    if (this.isDevelopment) {
      console.error('Validation Errors:', errors);
    }

    return errors;
  }

  /**
   * 处理异步操作错误
   * @param {Function} asyncFn - 异步函数
   * @param {Object} options - 选项
   * @returns {Promise} 异步操作结果
   */
  async handleAsync(asyncFn, options = {}) {
    try {
      return await asyncFn();
    } catch (error) {
      this.handleError(error, options);
      throw error; // 重新抛出错误，让调用者可以处理
    }
  }

  /**
   * 创建错误边界处理器
   * @param {Object} options - 选项
   * @returns {Function} 错误处理函数
   */
  createErrorBoundaryHandler(options = {}) {
    return (error, errorInfo) => {
      this.handleError(error, {
        ...options,
        context: 'React Error Boundary',
        showToast: true,
      });

      if (this.isDevelopment) {
        console.error('Component Stack:', errorInfo.componentStack);
      }
    };
  }

  /**
   * 处理 Promise 拒绝
   * @param {Error} error - Promise 拒绝错误
   */
  handlePromiseRejection(error) {
    this.handleError(error, {
      context: 'Unhandled Promise Rejection',
      showToast: false, // 不显示 toast，避免干扰用户
    });
  }

  /**
   * 处理全局错误
   * @param {Error} error - 全局错误
   */
  handleGlobalError(error) {
    this.handleError(error, {
      context: 'Global Error',
      showToast: true,
    });
  }
}

// 创建单例实例
const errorHandler = new ErrorHandler();

// 设置全局错误处理器（仅在浏览器环境）
if (typeof window !== 'undefined') {
  // 处理未捕获的错误
  window.addEventListener('error', (event) => {
    errorHandler.handleGlobalError(event.error);
  });

  // 处理未处理的 Promise 拒绝
  window.addEventListener('unhandledrejection', (event) => {
    errorHandler.handlePromiseRejection(event.reason);
  });
}

export default errorHandler;
