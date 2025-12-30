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

describe('MainLayout', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('User Menu Items', () => {
    it('should render user dropdown with correct menu items', () => {
      render(
        <TestWrapper>
          <MainLayout />
        </TestWrapper>
      );

      // Click on user avatar to open dropdown
      const userAvatar = screen.getByRole('img');
      fireEvent.click(userAvatar);

      // Check that Settings menu item exists
      expect(screen.getByText('Settings')).toBeInTheDocument();
      
      // Check that Logout menu item exists
      expect(screen.getByText('Logout')).toBeInTheDocument();
      
      // Check that Profile menu item does NOT exist
      expect(screen.queryByText('Profile')).not.toBeInTheDocument();
    });

    it('should navigate to settings when Settings is clicked', () => {
      render(
        <TestWrapper>
          <MainLayout />
        </TestWrapper>
      );

      // Click on user avatar to open dropdown
      const userAvatar = screen.getByRole('img');
      fireEvent.click(userAvatar);

      // Click on Settings
      const settingsItem = screen.getByText('Settings');
      fireEvent.click(settingsItem);

      expect(mockNavigate).toHaveBeenCalledWith('/settings');
    });

    it('should logout and navigate to login when Logout is clicked', () => {
      render(
        <TestWrapper>
          <MainLayout />
        </TestWrapper>
      );

      // Click on user avatar to open dropdown
      const userAvatar = screen.getByRole('img');
      fireEvent.click(userAvatar);

      // Click on Logout
      const logoutItem = screen.getByText('Logout');
      fireEvent.click(logoutItem);

      expect(mockLogout).toHaveBeenCalled();
      expect(mockNavigate).toHaveBeenCalledWith('/login');
    });
  });
});