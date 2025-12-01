import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { message } from 'antd';
import Logs from '../index';
import { activityLogsAPI } from '../../../api/activityLogs';
import type { ActivityLogsResponse } from '../../../types';

// Mock the API
jest.mock('../../../api/activityLogs');
const mockedActivityLogsAPI = activityLogsAPI as jest.Mocked<typeof activityLogsAPI>;

// Mock antd message
jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  message: {
    error: jest.fn(),
    success: jest.fn(),
    loading: jest.fn(),
  },
}));

describe('Logs Page', () => {
  const mockLogsResponse: ActivityLogsResponse = {
    status: 'success',
    data: {
      logs: [
        {
          id: 1,
          created_at: '2024-01-01T12:00:00Z',
          updated_at: '2024-01-01T12:00:00Z',
          level: 'INFO',
          message: 'Process started',
          action: 'start_process',
          resource: 'process',
          target: 'node1:web-app',
          user_id: '1',
          username: 'admin',
          ip_address: '192.168.1.100',
          user_agent: 'Mozilla/5.0',
          details: '',
          status: 'success',
          duration: 100,
        },
        {
          id: 2,
          created_at: '2024-01-01T11:00:00Z',
          updated_at: '2024-01-01T11:00:00Z',
          level: 'ERROR',
          message: 'Process failed',
          action: 'stop_process',
          resource: 'process',
          target: 'node2:worker',
          user_id: '2',
          username: 'john',
          ip_address: '192.168.1.101',
          user_agent: 'Mozilla/5.0',
          details: '',
          status: 'error',
          duration: 50,
        },
      ],
      pagination: {
        page: 1,
        page_size: 20,
        total: 2,
        total_pages: 1,
        has_next: false,
        has_prev: false,
      },
    },
  };

  beforeEach(() => {
    jest.clearAllMocks();
    mockedActivityLogsAPI.getActivityLogs.mockResolvedValue(mockLogsResponse);
  });

  describe('Component Rendering', () => {
    it('should render the component correctly', async () => {
      render(<Logs />);

      expect(screen.getByText('Activity Logs')).toBeInTheDocument();
      expect(screen.getByPlaceholderText('Search logs...')).toBeInTheDocument();
      expect(screen.getByText('Export')).toBeInTheDocument();
      expect(screen.getByText('Refresh')).toBeInTheDocument();
    });

    it('should display logs after loading', async () => {
      render(<Logs />);

      await waitFor(() => {
        expect(screen.getByText('Process started')).toBeInTheDocument();
        expect(screen.getByText('Process failed')).toBeInTheDocument();
      });
    });

    it('should display pagination information', async () => {
      render(<Logs />);

      await waitFor(() => {
        expect(screen.getByText(/1-2 of 2 logs/)).toBeInTheDocument();
      });
    });
  });

  describe('API Calls', () => {
    it('should call API on component mount', async () => {
      render(<Logs />);

      await waitFor(() => {
        expect(mockedActivityLogsAPI.getActivityLogs).toHaveBeenCalledTimes(1);
      });
    });

    it('should call API with correct default parameters', async () => {
      render(<Logs />);

      await waitFor(() => {
        expect(mockedActivityLogsAPI.getActivityLogs).toHaveBeenCalledWith({
          page: 1,
          page_size: 20,
        });
      });
    });

    it('should call API when refresh button is clicked', async () => {
      render(<Logs />);

      await waitFor(() => {
        expect(mockedActivityLogsAPI.getActivityLogs).toHaveBeenCalledTimes(1);
      });

      const refreshButton = screen.getByText('Refresh');
      fireEvent.click(refreshButton);

      await waitFor(() => {
        expect(mockedActivityLogsAPI.getActivityLogs).toHaveBeenCalledTimes(2);
      });
    });
  });

  describe('Filters', () => {
    it('should apply level filter when selected', async () => {
      render(<Logs />);

      await waitFor(() => {
        expect(screen.getByText('Process started')).toBeInTheDocument();
      });

      // Select level filter
      const levelSelect = screen.getByPlaceholderText('Level');
      fireEvent.mouseDown(levelSelect);
      
      await waitFor(() => {
        const errorOption = screen.getByText('ERROR');
        fireEvent.click(errorOption);
      });

      // Click search button
      const searchButton = screen.getByRole('button', { name: /search/i });
      fireEvent.click(searchButton);

      await waitFor(() => {
        expect(mockedActivityLogsAPI.getActivityLogs).toHaveBeenCalledWith(
          expect.objectContaining({
            level: 'ERROR',
            page: 1,
          })
        );
      });
    });

    it('should apply username filter when entered', async () => {
      render(<Logs />);

      await waitFor(() => {
        expect(screen.getByText('Process started')).toBeInTheDocument();
      });

      const usernameInput = screen.getByPlaceholderText('Username');
      fireEvent.change(usernameInput, { target: { value: 'admin' } });

      const searchButton = screen.getByRole('button', { name: /search/i });
      fireEvent.click(searchButton);

      await waitFor(() => {
        expect(mockedActivityLogsAPI.getActivityLogs).toHaveBeenCalledWith(
          expect.objectContaining({
            username: 'admin',
            page: 1,
          })
        );
      });
    });
  });

  describe('Pagination', () => {
    it('should handle page change', async () => {
      const multiPageResponse: ActivityLogsResponse = {
        ...mockLogsResponse,
        data: {
          ...mockLogsResponse.data,
          pagination: {
            page: 1,
            page_size: 20,
            total: 50,
            total_pages: 3,
            has_next: true,
            has_prev: false,
          },
        },
      };

      mockedActivityLogsAPI.getActivityLogs.mockResolvedValue(multiPageResponse);

      render(<Logs />);

      await waitFor(() => {
        expect(screen.getByText('Process started')).toBeInTheDocument();
      });

      // Find and click next page button
      const nextButton = screen.getByTitle('Next Page');
      fireEvent.click(nextButton);

      await waitFor(() => {
        expect(mockedActivityLogsAPI.getActivityLogs).toHaveBeenCalledWith(
          expect.objectContaining({
            page: 2,
          })
        );
      });
    });

    it('should handle page size change', async () => {
      render(<Logs />);

      await waitFor(() => {
        expect(screen.getByText('Process started')).toBeInTheDocument();
      });

      // Change page size
      const pageSizeSelect = screen.getByTitle('20 / page');
      fireEvent.mouseDown(pageSizeSelect);

      await waitFor(() => {
        const option50 = screen.getByTitle('50 / page');
        fireEvent.click(option50);
      });

      await waitFor(() => {
        expect(mockedActivityLogsAPI.getActivityLogs).toHaveBeenCalledWith(
          expect.objectContaining({
            page_size: 50,
          })
        );
      });
    });
  });

  describe('Error Handling', () => {
    it('should display error message when API call fails', async () => {
      const error = new Error('Network error');
      mockedActivityLogsAPI.getActivityLogs.mockRejectedValue(error);

      render(<Logs />);

      await waitFor(() => {
        expect(message.error).toHaveBeenCalledWith('Network error');
      });

      expect(screen.getByText('Error')).toBeInTheDocument();
      expect(screen.getByText('Network error')).toBeInTheDocument();
    });

    it('should allow closing error alert', async () => {
      const error = new Error('Test error');
      mockedActivityLogsAPI.getActivityLogs.mockRejectedValue(error);

      render(<Logs />);

      await waitFor(() => {
        expect(screen.getByText('Test error')).toBeInTheDocument();
      });

      const closeButton = screen.getByRole('button', { name: /close/i });
      fireEvent.click(closeButton);

      await waitFor(() => {
        expect(screen.queryByText('Test error')).not.toBeInTheDocument();
      });
    });
  });

  describe('Loading State', () => {
    it('should show loading state while fetching data', async () => {
      mockedActivityLogsAPI.getActivityLogs.mockImplementation(
        () => new Promise(resolve => setTimeout(() => resolve(mockLogsResponse), 100))
      );

      render(<Logs />);

      // Check for loading spinner in table
      expect(screen.getByRole('table')).toBeInTheDocument();
      
      await waitFor(() => {
        expect(screen.getByText('Process started')).toBeInTheDocument();
      });
    });

    it('should disable export button while loading', async () => {
      mockedActivityLogsAPI.getActivityLogs.mockImplementation(
        () => new Promise(resolve => setTimeout(() => resolve(mockLogsResponse), 100))
      );

      render(<Logs />);

      const exportButton = screen.getByText('Export');
      expect(exportButton).toBeDisabled();

      await waitFor(() => {
        expect(exportButton).not.toBeDisabled();
      });
    });
  });

  describe('Export Functionality', () => {
    it('should call export API when export button is clicked', async () => {
      const mockBlob = new Blob(['test'], { type: 'text/csv' });
      mockedActivityLogsAPI.exportLogs.mockResolvedValue(mockBlob);

      // Mock URL.createObjectURL
      global.URL.createObjectURL = jest.fn(() => 'blob:test');
      global.URL.revokeObjectURL = jest.fn();

      render(<Logs />);

      await waitFor(() => {
        expect(screen.getByText('Process started')).toBeInTheDocument();
      });

      const exportButton = screen.getByText('Export');
      fireEvent.click(exportButton);

      await waitFor(() => {
        expect(mockedActivityLogsAPI.exportLogs).toHaveBeenCalled();
      });
    });
  });
});
