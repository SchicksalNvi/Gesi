import { useState, useEffect } from 'react';

interface BreakpointMap {
  xs: boolean; // < 576px
  sm: boolean; // >= 576px
  md: boolean; // >= 768px
  lg: boolean; // >= 992px
  xl: boolean; // >= 1200px
  xxl: boolean; // >= 1600px
}

const breakpoints = {
  xs: 0,
  sm: 576,
  md: 768,
  lg: 992,
  xl: 1200,
  xxl: 1600,
};

/**
 * Hook to detect responsive breakpoints
 * @returns Object with boolean values for each breakpoint
 */
export function useResponsive(): BreakpointMap {
  const [screenMap, setScreenMap] = useState<BreakpointMap>({
    xs: false,
    sm: false,
    md: false,
    lg: false,
    xl: false,
    xxl: false,
  });

  useEffect(() => {
    const updateScreenMap = () => {
      const width = window.innerWidth;
      
      setScreenMap({
        xs: width < breakpoints.sm,
        sm: width >= breakpoints.sm && width < breakpoints.md,
        md: width >= breakpoints.md && width < breakpoints.lg,
        lg: width >= breakpoints.lg && width < breakpoints.xl,
        xl: width >= breakpoints.xl && width < breakpoints.xxl,
        xxl: width >= breakpoints.xxl,
      });
    };

    // Initial check
    updateScreenMap();

    // Add event listener
    window.addEventListener('resize', updateScreenMap);

    // Cleanup
    return () => {
      window.removeEventListener('resize', updateScreenMap);
    };
  }, []);

  return screenMap;
}

/**
 * Hook to check if screen is mobile size
 * @returns boolean indicating if screen is mobile (< 768px)
 */
export function useIsMobile(): boolean {
  const [isMobile, setIsMobile] = useState(false);

  useEffect(() => {
    const checkIsMobile = () => {
      setIsMobile(window.innerWidth < breakpoints.md);
    };

    // Initial check
    checkIsMobile();

    // Add event listener
    window.addEventListener('resize', checkIsMobile);

    // Cleanup
    return () => {
      window.removeEventListener('resize', checkIsMobile);
    };
  }, []);

  return isMobile;
}

export default useResponsive;