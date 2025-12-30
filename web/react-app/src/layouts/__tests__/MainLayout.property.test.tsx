import { render, screen, fireEvent } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import MainLayout from '../MainLayout';

// Mock the store
const mockLogout = jest.fn();
const mockUser = { username: 'testuser' };

jest.mock('@/store', () => ({
  useStore: () => ({
    user: mockUser,
    logout: mockLogout,
  }),
}));

// Mock react-router-dom navigate
const mockNavigate = jest.fn();
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockNavigate,
  useLocation: () => ({ pathname: '/dashboard' }),
  Outlet: () => <div data-testid="outlet">Content</div>,
}));

// Test wrapper component
const TestWrapper = ({ children }: { children: React.ReactNode }) => (
  <BrowserRouter>
    <ConfigProvider>
      {children}
    </ConfigProvider>
  </BrowserRouter>
);

// Property-based test helper
const runPropertyTest = (testFn: () => void, iterations: number = 100) => {
  for (let i = 0; i < iterations; i++) {
    testFn();
  }
};

describe('MainLayout Property Tests', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  /**
   * Feature: remove-profile-menu, Property 1: Menu Item Count Consistency
   * For any user dropdown menu render, the menu should contain exactly 3 items: settings, divider, and logout
   * Validates: Requirements 1.1, 1.2
   */
  it('should always have exactly 3 menu items in user dropdown', () => {
    runPropertyTest(() => {
      const { unmount } = render(
        <TestWrapper>
          <MainLayout />
        </TestWrapper>
      );

      // Click on user avatar to open dropdown
      const userAvatar = screen.getByRole('img');
      fireEvent.click(userAvatar);

      // Get all menu items (including divider)
      const menuItems = document.querySelectorAll('.ant-dropdown-menu-item, .ant-dropdown-menu-divider');
      
      // Should have exactly 3 items: settings, divider, logout
      expect(menuItems).toHaveLength(3);
      
      // Verify the specific items exist
      expect(screen.getByText('Settings')).toBeInTheDocument();
      expect(screen.getByText('Logout')).toBeInTheDocument();
      expect(document.querySelector('.ant-dropdown-menu-divider')).toBeInTheDocument();
      
      // Verify Profile does not exist
      expect(screen.queryByText('Profile')).not.toBeInTheDocument();

      unmount();
    }, 100);
  });

  /**
   * Feature: remove-profile-menu, Property 2: Navigation Functionality Preservation
   * For any settings or logout menu item click, the navigation behavior should remain identical to the original implementation
   * Validates: Requirements 1.3, 1.4, 2.3
   */
  it('should preserve navigation functionality for settings and logout', () => {
    runPropertyTest(() => {
      const { unmount } = render(
        <TestWrapper>
          <MainLayout />
        </TestWrapper>
      );

      // Test Settings navigation
      const userAvatar = screen.getByRole('img');
      fireEvent.click(userAvatar);
      
      const settingsItem = screen.getByText('Settings');
      fireEvent.click(settingsItem);
      
      expect(mockNavigate).toHaveBeenCalledWith('/settings');
      
      // Reset mocks for logout test
      mockNavigate.mockClear();
      mockLogout.mockClear();
      
      // Re-open dropdown for logout test
      fireEvent.click(userAvatar);
      
      const logoutItem = screen.getByText('Logout');
      fireEvent.click(logoutItem);
      
      expect(mockLogout).toHaveBeenCalled();
      expect(mockNavigate).toHaveBeenCalledWith('/login');

      unmount();
    }, 100);
  });

  /**
   * Feature: remove-profile-menu, Property 3: Visual Structure Preservation
   * For any dropdown menu display, the divider should appear between settings and logout items maintaining visual separation
   * Validates: Requirements 1.5, 2.4
   */
  it('should maintain visual structure with divider between settings and logout', () => {
    runPropertyTest(() => {
      const { unmount } = render(
        <TestWrapper>
          <MainLayout />
        </TestWrapper>
      );

      // Click on user avatar to open dropdown
      const userAvatar = screen.getByRole('img');
      fireEvent.click(userAvatar);

      // Get all menu elements in order
      const menuContainer = document.querySelector('.ant-dropdown-menu');
      expect(menuContainer).toBeInTheDocument();
      
      const menuChildren = menuContainer?.children;
      expect(menuChildren).toHaveLength(3);
      
      // Verify order: Settings, Divider, Logout
      expect(menuChildren?.[0]).toHaveClass('ant-dropdown-menu-item');
      expect(menuChildren?.[0]?.textContent).toContain('Settings');
      
      expect(menuChildren?.[1]).toHaveClass('ant-dropdown-menu-divider');
      
      expect(menuChildren?.[2]).toHaveClass('ant-dropdown-menu-item');
      expect(menuChildren?.[2]?.textContent).toContain('Logout');

      unmount();
    }, 100);
  });
});