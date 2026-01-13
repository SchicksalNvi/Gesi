import { useState, useEffect } from 'react';

/**
 * Custom hook for persisting state to localStorage
 * @param key - localStorage key
 * @param initialValue - initial value if no stored value exists
 * @returns [value, setValue] tuple similar to useState
 */
export function useLocalStorage<T>(
  key: string,
  initialValue: T
): [T, (value: T | ((val: T) => T)) => void] {
  // Get value from localStorage or use initial value
  const [storedValue, setStoredValue] = useState<T>(() => {
    try {
      const item = window.localStorage.getItem(key);
      return item ? JSON.parse(item) : initialValue;
    } catch (error) {
      console.warn(`Error reading localStorage key "${key}":`, error);
      return initialValue;
    }
  });

  // Return a wrapped version of useState's setter function that persists the new value to localStorage
  const setValue = (value: T | ((val: T) => T)) => {
    try {
      // Allow value to be a function so we have the same API as useState
      const valueToStore = value instanceof Function ? value(storedValue) : value;
      
      // Save state
      setStoredValue(valueToStore);
      
      // Save to localStorage
      window.localStorage.setItem(key, JSON.stringify(valueToStore));
    } catch (error) {
      console.warn(`Error setting localStorage key "${key}":`, error);
    }
  };

  return [storedValue, setValue];
}

/**
 * Hook specifically for view mode persistence
 */
export function useViewMode(defaultMode: 'card' | 'list' = 'card') {
  return useLocalStorage<'card' | 'list'>('nodes-view-mode', defaultMode);
}

/**
 * Hook for persisting search query
 */
export function usePersistedSearch(key: string = 'nodes-search') {
  return useLocalStorage<string>(key, '');
}

/**
 * Hook for persisting filter state
 */
export function usePersistedFilters<T extends Record<string, any>>(
  key: string,
  initialFilters: T
) {
  return useLocalStorage<T>(key, initialFilters);
}

export default useLocalStorage;