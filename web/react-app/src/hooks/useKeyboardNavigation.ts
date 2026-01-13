import { useEffect, useCallback } from 'react';

interface KeyboardNavigationOptions {
  onEnter?: () => void;
  onEscape?: () => void;
  onArrowUp?: () => void;
  onArrowDown?: () => void;
  onArrowLeft?: () => void;
  onArrowRight?: () => void;
  onTab?: () => void;
  onShiftTab?: () => void;
  enabled?: boolean;
}

/**
 * Hook for handling keyboard navigation
 * @param options - Keyboard event handlers
 */
export function useKeyboardNavigation(options: KeyboardNavigationOptions) {
  const {
    onEnter,
    onEscape,
    onArrowUp,
    onArrowDown,
    onArrowLeft,
    onArrowRight,
    onTab,
    onShiftTab,
    enabled = true,
  } = options;

  const handleKeyDown = useCallback((event: KeyboardEvent) => {
    if (!enabled) return;

    switch (event.key) {
      case 'Enter':
        if (onEnter) {
          event.preventDefault();
          onEnter();
        }
        break;
      case 'Escape':
        if (onEscape) {
          event.preventDefault();
          onEscape();
        }
        break;
      case 'ArrowUp':
        if (onArrowUp) {
          event.preventDefault();
          onArrowUp();
        }
        break;
      case 'ArrowDown':
        if (onArrowDown) {
          event.preventDefault();
          onArrowDown();
        }
        break;
      case 'ArrowLeft':
        if (onArrowLeft) {
          event.preventDefault();
          onArrowLeft();
        }
        break;
      case 'ArrowRight':
        if (onArrowRight) {
          event.preventDefault();
          onArrowRight();
        }
        break;
      case 'Tab':
        if (event.shiftKey && onShiftTab) {
          event.preventDefault();
          onShiftTab();
        } else if (!event.shiftKey && onTab) {
          event.preventDefault();
          onTab();
        }
        break;
    }
  }, [enabled, onEnter, onEscape, onArrowUp, onArrowDown, onArrowLeft, onArrowRight, onTab, onShiftTab]);

  useEffect(() => {
    if (enabled) {
      document.addEventListener('keydown', handleKeyDown);
      return () => {
        document.removeEventListener('keydown', handleKeyDown);
      };
    }
  }, [handleKeyDown, enabled]);
}

/**
 * Hook for managing focus within a container
 * @param containerRef - Ref to the container element
 * @param itemSelector - CSS selector for focusable items
 */
export function useFocusManagement(
  containerRef: React.RefObject<HTMLElement>,
  itemSelector: string = '[tabindex="0"], button, input, select, textarea, [href]'
) {
  const focusFirst = useCallback(() => {
    if (containerRef.current) {
      const firstItem = containerRef.current.querySelector(itemSelector) as HTMLElement;
      if (firstItem) {
        firstItem.focus();
      }
    }
  }, [containerRef, itemSelector]);

  const focusLast = useCallback(() => {
    if (containerRef.current) {
      const items = containerRef.current.querySelectorAll(itemSelector);
      const lastItem = items[items.length - 1] as HTMLElement;
      if (lastItem) {
        lastItem.focus();
      }
    }
  }, [containerRef, itemSelector]);

  const focusNext = useCallback(() => {
    if (containerRef.current) {
      const items = Array.from(containerRef.current.querySelectorAll(itemSelector)) as HTMLElement[];
      const currentIndex = items.findIndex(item => item === document.activeElement);
      
      if (currentIndex >= 0 && currentIndex < items.length - 1) {
        items[currentIndex + 1].focus();
      } else if (currentIndex === items.length - 1) {
        items[0].focus(); // Wrap to first
      } else {
        focusFirst();
      }
    }
  }, [containerRef, itemSelector, focusFirst]);

  const focusPrevious = useCallback(() => {
    if (containerRef.current) {
      const items = Array.from(containerRef.current.querySelectorAll(itemSelector)) as HTMLElement[];
      const currentIndex = items.findIndex(item => item === document.activeElement);
      
      if (currentIndex > 0) {
        items[currentIndex - 1].focus();
      } else if (currentIndex === 0) {
        items[items.length - 1].focus(); // Wrap to last
      } else {
        focusLast();
      }
    }
  }, [containerRef, itemSelector, focusLast]);

  return {
    focusFirst,
    focusLast,
    focusNext,
    focusPrevious,
  };
}

export default useKeyboardNavigation;