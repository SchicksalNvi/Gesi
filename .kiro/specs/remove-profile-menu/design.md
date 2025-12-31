# Design Document

## Overview

This design outlines the removal of the non-functional "Profile" menu item from the user dropdown in the MainLayout component. The change simplifies the UI by keeping only functional menu items: "Settings" and "Logout".

## Architecture

The change affects only the frontend React component:
- **Component**: `web/react-app/src/layouts/MainLayout.tsx`
- **Scope**: User dropdown menu configuration
- **Impact**: UI simplification, no backend changes required

## Components and Interfaces

### MainLayout Component
- **Location**: `web/react-app/src/layouts/MainLayout.tsx`
- **Change**: Modify `userMenuItems` array to remove profile entry
- **Dependencies**: No new dependencies, existing Ant Design components remain

### User Menu Items Structure
```typescript
const userMenuItems: MenuProps['items'] = [
  // Remove profile item
  {
    key: 'settings',
    icon: <SettingOutlined />,
    label: 'Settings',
    onClick: () => navigate('/settings'),
  },
  {
    type: 'divider',
  },
  {
    key: 'logout',
    icon: <LogoutOutlined />,
    label: 'Logout',
    onClick: () => {
      logout();
      navigate('/login');
    },
  },
];
```

## Data Models

No data model changes required. This is purely a UI modification.

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Menu Item Count Consistency
*For any* user dropdown menu render, the menu should contain exactly 3 items: settings, divider, and logout
**Validates: Requirements 1.1, 1.2**

### Property 2: Navigation Functionality Preservation  
*For any* settings or logout menu item click, the navigation behavior should remain identical to the original implementation
**Validates: Requirements 1.3, 1.4, 2.3**

### Property 3: Visual Structure Preservation
*For any* dropdown menu display, the divider should appear between settings and logout items maintaining visual separation
**Validates: Requirements 1.5, 2.4**

## Error Handling

No additional error handling required. Existing error handling for navigation and logout remains unchanged.

## Testing Strategy

### Unit Tests
- Test that userMenuItems array contains only expected items
- Test that settings navigation works correctly
- Test that logout functionality works correctly
- Test that divider is properly positioned

### Property Tests
- **Property 1**: Generate random component renders and verify menu item count is always 3
- **Property 2**: Generate random menu interactions and verify navigation behavior consistency
- **Property 3**: Generate random dropdown displays and verify divider positioning

**Dual Testing Approach**: Unit tests verify specific menu structure and interactions, while property tests ensure consistent behavior across all possible component states. Both are necessary for comprehensive coverage.

**Property Test Configuration**: Use React Testing Library with minimum 100 iterations per property test. Each test references its design document property with format: **Feature: remove-profile-menu, Property {number}: {property_text}**