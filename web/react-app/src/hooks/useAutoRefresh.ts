import { useEffect, useRef } from 'react';
import { useStore } from '@/store';

/**
 * 自动刷新 Hook
 * 从全局 store 读取 autoRefreshEnabled 和 refreshInterval
 * 
 * @param callback 要执行的回调函数
 */
export const useAutoRefresh = (callback: () => void | Promise<void>) => {
  const { autoRefreshEnabled, refreshInterval } = useStore();
  const callbackRef = useRef(callback);

  useEffect(() => {
    callbackRef.current = callback;
  }, [callback]);

  useEffect(() => {
    if (!autoRefreshEnabled || refreshInterval <= 0) return;

    const id = setInterval(() => {
      callbackRef.current();
    }, refreshInterval * 1000);

    return () => clearInterval(id);
  }, [autoRefreshEnabled, refreshInterval]);
};
