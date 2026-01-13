import { useMemo } from 'react';
import { useDebounce } from './useDebounce';
import { Node } from '@/types';
import { NodeFilters } from '@/components/FilterBar';

/**
 * Custom hook for filtering and searching nodes
 * @param nodes - Array of nodes to filter
 * @param searchQuery - Search query string
 * @param filters - Filter criteria object
 * @returns Filtered nodes array
 */
export function useNodeFiltering(
  nodes: Node[],
  searchQuery: string,
  filters: NodeFilters
) {
  // Debounce search query to avoid excessive filtering
  const debouncedSearch = useDebounce(searchQuery, 300);

  const filteredNodes = useMemo(() => {
    return nodes.filter(node => {
      // Search logic - check name, host, and environment
      if (debouncedSearch) {
        const searchLower = debouncedSearch.toLowerCase();
        const matchesSearch = 
          node.name.toLowerCase().includes(searchLower) ||
          node.host.toLowerCase().includes(searchLower) ||
          node.environment.toLowerCase().includes(searchLower) ||
          (node.username && node.username.toLowerCase().includes(searchLower));
        
        if (!matchesSearch) {
          return false;
        }
      }

      // Status filter
      if (filters.status) {
        const isOnline = node.is_connected;
        if (filters.status === 'online' && !isOnline) {
          return false;
        }
        if (filters.status === 'offline' && isOnline) {
          return false;
        }
      }

      // Environment filter
      if (filters.environment && node.environment !== filters.environment) {
        return false;
      }

      // Process count range filter
      if (filters.processCountRange) {
        const processCount = node.process_count || 0;
        const [min, max] = filters.processCountRange;
        
        if (min !== undefined && processCount < min) {
          return false;
        }
        if (max !== undefined && processCount > max) {
          return false;
        }
      }

      return true;
    });
  }, [nodes, debouncedSearch, filters]);

  // Extract unique environments for filter dropdown
  const availableEnvironments = useMemo(() => {
    const environments = new Set(nodes.map(node => node.environment));
    return Array.from(environments).sort();
  }, [nodes]);

  // Calculate statistics
  const stats = useMemo(() => {
    const total = nodes.length;
    const filtered = filteredNodes.length;
    const online = filteredNodes.filter(node => node.is_connected).length;
    const offline = filtered - online;
    
    return {
      total,
      filtered,
      online,
      offline,
      isFiltered: filtered !== total
    };
  }, [nodes.length, filteredNodes]);

  return {
    filteredNodes,
    availableEnvironments,
    stats,
    debouncedSearch
  };
}

/**
 * Hook for highlighting search matches in text
 * @param text - Text to highlight
 * @param searchQuery - Search query to highlight
 * @returns Text with highlighted matches (for use with dangerouslySetInnerHTML)
 */
export function useSearchHighlight(text: string, searchQuery: string) {
  return useMemo(() => {
    if (!searchQuery || !text) {
      return text;
    }

    const regex = new RegExp(`(${searchQuery.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
    return text.replace(regex, '<mark style="background-color: #fff566; padding: 0;">$1</mark>');
  }, [text, searchQuery]);
}

export default useNodeFiltering;