import { useEffect, useRef, useState } from 'react';
import api from '../api/client';

/**
 * 自动刷新 Hook
 * 从系统设置中读取刷新间隔，并自动执行回调函数
 * 
 * @param callback 要执行的回调函数
 * @param enabled 是否启用自动刷新（默认从用户偏好读取）
 * @param dependencies 依赖项数组
 */
export const useAutoRefresh = (
  callback: () => void | Promise<void>,
  enabled?: boolean,
  dependencies: any[] = []
) => {
  const [refreshInterval, setRefreshInterval] = useState<number>(30); // 默认 30 秒
  const [autoRefreshEnabled, setAutoRefreshEnabled] = useState<boolean>(true);
  const intervalRef = useRef<NodeJS.Timeout | null>(null);
  const callbackRef = useRef(callback);

  // 更新回调引用
  useEffect(() => {
    callbackRef.current = callback;
  }, [callback]);

  // 加载系统设置和用户偏好
  useEffect(() => {
    const loadSettings = async () => {
      try {
        // 加载用户偏好
        const prefsResponse = await api.get('/system-settings/user-preferences');
        if (prefsResponse.data) {
          const prefs = prefsResponse.data;
          setAutoRefreshEnabled(prefs.auto_refresh !== false);
          if (prefs.refresh_interval) {
            setRefreshInterval(prefs.refresh_interval);
          }
        }

        // 如果用户偏好中没有设置，则从系统设置中读取
        if (!prefsResponse.data?.refresh_interval) {
          const settingsResponse = await api.get('/system-settings');
          if (settingsResponse.data?.settings?.refresh_interval) {
            setRefreshInterval(parseInt(settingsResponse.data.settings.refresh_interval, 10));
          }
        }
      } catch (error) {
        console.error('Failed to load refresh settings:', error);
      }
    };

    loadSettings();
  }, []);

  // 设置定时器
  useEffect(() => {
    // 清除旧的定时器
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }

    // 判断是否启用自动刷新
    const shouldEnable = enabled !== undefined ? enabled : autoRefreshEnabled;

    if (shouldEnable && refreshInterval > 0) {
      // 立即执行一次
      callbackRef.current();

      // 设置定时器
      intervalRef.current = setInterval(() => {
        callbackRef.current();
      }, refreshInterval * 1000);
    }

    // 清理函数
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [refreshInterval, autoRefreshEnabled, enabled, ...dependencies]);

  return {
    refreshInterval,
    autoRefreshEnabled,
    setRefreshInterval,
    setAutoRefreshEnabled,
  };
};
