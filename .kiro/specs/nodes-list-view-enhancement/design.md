# Design Document

## Overview

This design enhances the nodes page with dual view modes, advanced filtering, and search capabilities optimized for managing hundreds of supervisor nodes. The solution prioritizes performance, usability, and scalability while maintaining the existing real-time update functionality.

## Architecture

### Component Structure

```
NodesPage
├── NodesHeader (title, refresh, view toggle)
├── NodesToolbar (search, filters, bulk actions)
├── NodesCardView (existing enhanced)
└── NodesListView (new table component)
```

### State Management

The component will manage:
- `viewMode`: 'card' | 'list' (persisted to localStorage)
- `searchQuery`: string for text search
- `filters`: object containing active filters
- `selectedNodes`: array of selected node IDs (list view only)
- `nodes`: filtered and sorted node array

## Components and Interfaces

### NodesToolbar Component

```typescript
interface NodesToolbarProps {
  searchQuery: string;
  onSearchChange: (query: string) => void;
  filters: NodeFilters;
  onFiltersChange: (filters: NodeFilters) => void;
  selectedNodes: string[];
  onBulkAction: (action: BulkAction) => void;
  totalNodes: number;
  filteredNodes: number;
}

interface NodeFilters {
  status?: 'online' | 'offline';
  environment?: string;
  processCountRange?: [number, number];
}
```

### NodesListView Component

```typescript
interface NodesListViewProps {
  nodes: Node[];
  loading: boolean;
  selectedNodes: string[];
  onSelectionChange: (nodeIds: string[]) => void;
  onNodeClick: (nodeName: string) => void;
}

interface ListViewColumn {
  key: string;
  title: string;
  dataIndex: string;
  width?: number;
  render?: (value: any, record: Node) => ReactNode;
  sorter?: boolean;
  responsive?: string[];
}
```

### ViewToggle Component

```typescript
interface ViewToggleProps {
  value: 'card' | 'list';
  onChange: (mode: 'card' | 'list') => void;
  disabled?: boolean;
}
```

## Data Models

### Enhanced Node Interface

```typescript
interface Node {
  name: string;
  host: string;
  port: number;
  username?: string;
  environment: string;
  is_connected: boolean;
  process_count: number;
  last_updated: string;
  supervisor_version?: string;
  uptime?: number;
}

interface NodeListItem extends Node {
  key: string; // for table row key
  statusColor: string; // computed status color
  displayName: string; // formatted display name
}
```

## Performance Optimizations

### Virtual Scrolling Implementation

For list view with 100+ nodes:

```typescript
import { FixedSizeList as List } from 'react-window';

const VirtualizedTable = ({ nodes, height = 400 }) => {
  const Row = ({ index, style }) => (
    <div style={style}>
      <NodeRow node={nodes[index]} />
    </div>
  );

  return (
    <List
      height={height}
      itemCount={nodes.length}
      itemSize={48} // row height
    >
      {Row}
    </List>
  );
};
```

### Search and Filter Optimization

```typescript
// Debounced search with useMemo for performance
const useNodeFiltering = (nodes: Node[], searchQuery: string, filters: NodeFilters) => {
  const debouncedSearch = useDebounce(searchQuery, 300);
  
  return useMemo(() => {
    return nodes.filter(node => {
      // Search logic
      if (debouncedSearch) {
        const searchLower = debouncedSearch.toLowerCase();
        if (!node.name.toLowerCase().includes(searchLower) &&
            !node.host.toLowerCase().includes(searchLower) &&
            !node.environment.toLowerCase().includes(searchLower)) {
          return false;
        }
      }
      
      // Filter logic
      if (filters.status && 
          ((filters.status === 'online') !== node.is_connected)) {
        return false;
      }
      
      if (filters.environment && 
          node.environment !== filters.environment) {
        return false;
      }
      
      return true;
    });
  }, [nodes, debouncedSearch, filters]);
};
```

### Pagination for Card View

```typescript
const CARDS_PER_PAGE = 50;

const useCardPagination = (nodes: Node[]) => {
  const [currentPage, setCurrentPage] = useState(1);
  
  const paginatedNodes = useMemo(() => {
    const start = (currentPage - 1) * CARDS_PER_PAGE;
    return nodes.slice(start, start + CARDS_PER_PAGE);
  }, [nodes, currentPage]);
  
  return { paginatedNodes, currentPage, setCurrentPage };
};
```

## User Interface Design

### Layout Breakpoints

- **Desktop (≥1200px)**: Full toolbar, both view modes available
- **Tablet (768px-1199px)**: Collapsed filters, both view modes
- **Mobile (<768px)**: List view only, minimal toolbar

### List View Columns

| Column | Width | Responsive | Sortable |
|--------|-------|------------|----------|
| Status | 60px | Always | Yes |
| Name | 200px | Always | Yes |
| Environment | 120px | Desktop+ | Yes |
| Host:Port | 180px | Tablet+ | No |
| Processes | 100px | Desktop+ | Yes |
| Actions | 80px | Always | No |

### Color Coding

- **Online**: Green (#52c41a)
- **Offline**: Red (#ff4d4f)
- **Unknown**: Gray (#d9d9d9)
- **Selected Row**: Blue background (#e6f7ff)

## Error Handling

### Search Performance

```typescript
const handleSearchError = (error: Error) => {
  console.warn('Search filtering error:', error);
  // Fallback to simple string matching
  return nodes.filter(node => 
    node.name.includes(searchQuery)
  );
};
```

### Virtual Scrolling Fallback

```typescript
const ListViewWithFallback = ({ nodes, ...props }) => {
  const [useVirtual, setUseVirtual] = useState(nodes.length > 100);
  
  if (useVirtual) {
    try {
      return <VirtualizedTable nodes={nodes} {...props} />;
    } catch (error) {
      console.warn('Virtual scrolling failed, using standard table');
      setUseVirtual(false);
    }
  }
  
  return <StandardTable nodes={nodes} {...props} />;
};
```

## Testing Strategy

### Unit Tests
- Component rendering with different props
- Search and filter logic correctness
- View mode persistence
- Bulk selection functionality

### Property-Based Tests
- **Property 1: Search consistency** - For any search query, results should contain the query string
- **Property 2: Filter preservation** - Applying and removing filters should return to original state
- **Property 3: Selection integrity** - Selected nodes should remain valid after filtering

### Performance Tests
- Render time with 500+ nodes
- Search response time with large datasets
- Memory usage during virtual scrolling
- Filter application speed

## Accessibility

- **Keyboard Navigation**: Full keyboard support for all interactions
- **Screen Readers**: Proper ARIA labels and roles
- **High Contrast**: Support for high contrast themes
- **Focus Management**: Clear focus indicators and logical tab order

## Browser Compatibility

- **Modern Browsers**: Chrome 90+, Firefox 88+, Safari 14+, Edge 90+
- **Graceful Degradation**: Basic functionality on older browsers
- **Mobile Support**: Touch-friendly interactions on mobile devices