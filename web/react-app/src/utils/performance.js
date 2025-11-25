/**
 * 前端性能监控工具
 * 验证需求：11.3
 */

/**
 * 测量页面加载性能
 */
export const measurePageLoad = () => {
  if (!window.performance || !window.performance.timing) {
    console.warn('Performance API not supported');
    return null;
  }

  const timing = window.performance.timing;
  const navigation = window.performance.navigation;

  // 计算各项指标
  const metrics = {
    // 页面加载总时间
    pageLoadTime: timing.loadEventEnd - timing.navigationStart,
    
    // DOM 准备时间
    domReadyTime: timing.domContentLoadedEventEnd - timing.navigationStart,
    
    // DNS 查询时间
    dnsTime: timing.domainLookupEnd - timing.domainLookupStart,
    
    // TCP 连接时间
    tcpTime: timing.connectEnd - timing.connectStart,
    
    // 请求响应时间
    requestTime: timing.responseEnd - timing.requestStart,
    
    // DOM 解析时间
    domParseTime: timing.domComplete - timing.domInteractive,
    
    // 资源加载时间
    resourceLoadTime: timing.loadEventEnd - timing.domContentLoadedEventEnd,
    
    // 导航类型
    navigationType: navigation.type,
    
    // 重定向次数
    redirectCount: navigation.redirectCount,
  };

  // 在开发环境输出详细信息
  if (process.env.NODE_ENV === 'development') {
    console.log('📊 Performance Metrics:', metrics);
    console.log(`  Page Load Time: ${metrics.pageLoadTime}ms`);
    console.log(`  DOM Ready Time: ${metrics.domReadyTime}ms`);
    console.log(`  Request Time: ${metrics.requestTime}ms`);
  }

  return metrics;
};

/**
 * 测量 Web Vitals 指标
 */
export const measureWebVitals = (callback) => {
  // 使用 web-vitals 库测量核心指标
  if (typeof window !== 'undefined') {
    import('web-vitals').then(({ getCLS, getFID, getFCP, getLCP, getTTFB }) => {
      getCLS(callback); // Cumulative Layout Shift
      getFID(callback); // First Input Delay
      getFCP(callback); // First Contentful Paint
      getLCP(callback); // Largest Contentful Paint
      getTTFB(callback); // Time to First Byte
    });
  }
};

/**
 * 发送性能指标到服务器
 */
export const sendMetrics = (metrics) => {
  if (process.env.NODE_ENV === 'production') {
    // 使用 sendBeacon API 发送数据（不阻塞页面卸载）
    if (navigator.sendBeacon) {
      const blob = new Blob([JSON.stringify(metrics)], {
        type: 'application/json',
      });
      navigator.sendBeacon('/api/metrics/performance', blob);
    } else {
      // 降级方案：使用普通 fetch
      fetch('/api/metrics/performance', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(metrics),
        keepalive: true,
      }).catch((err) => {
        console.error('Failed to send metrics:', err);
      });
    }
  }
};

/**
 * 监控组件渲染性能
 */
export const measureComponentRender = (componentName) => {
  const startTime = performance.now();

  return () => {
    const endTime = performance.now();
    const renderTime = endTime - startTime;

    if (process.env.NODE_ENV === 'development' && renderTime > 16) {
      // 超过 16ms（60fps）时警告
      console.warn(
        `⚠️  Slow render: ${componentName} took ${renderTime.toFixed(2)}ms`
      );
    }

    return renderTime;
  };
};

/**
 * 初始化性能监控
 */
export const initPerformanceMonitoring = () => {
  // 页面加载完成后测量性能
  if (document.readyState === 'complete') {
    const metrics = measurePageLoad();
    if (metrics) {
      sendMetrics(metrics);
    }
  } else {
    window.addEventListener('load', () => {
      setTimeout(() => {
        const metrics = measurePageLoad();
        if (metrics) {
          sendMetrics(metrics);
        }
      }, 0);
    });
  }

  // 测量 Web Vitals
  measureWebVitals((metric) => {
    if (process.env.NODE_ENV === 'development') {
      console.log(`📈 ${metric.name}:`, metric.value);
    }
    sendMetrics({
      name: metric.name,
      value: metric.value,
      rating: metric.rating,
    });
  });
};

export default {
  measurePageLoad,
  measureWebVitals,
  sendMetrics,
  measureComponentRender,
  initPerformanceMonitoring,
};
