import { create } from 'zustand';
import { User, Node, SystemStats } from '@/types';
import { UserPreferences } from '@/api/settings';

interface AppState {
  // Auth state
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  setUser: (user: User | null) => void;
  setToken: (token: string | null) => void;
  logout: () => void;

  // User preferences state
  userPreferences: UserPreferences | null;
  setUserPreferences: (preferences: UserPreferences | null) => void;

  // System settings state
  websocketEnabled: boolean;
  setWebsocketEnabled: (enabled: boolean) => void;

  // Nodes state
  nodes: Node[];
  selectedNode: Node | null;
  setNodes: (nodes: Node[]) => void;
  setSelectedNode: (node: Node | null) => void;

  // System stats
  systemStats: SystemStats | null;
  setSystemStats: (stats: SystemStats) => void;

  // UI state
  sidebarCollapsed: boolean;
  toggleSidebar: () => void;
}

export const useStore = create<AppState>((set, get) => ({
  // Auth state - Initialize based on localStorage
  user: JSON.parse(localStorage.getItem('user') || 'null'),
  token: localStorage.getItem('token'),
  // Initialize as authenticated if both user and token exist
  isAuthenticated: !!(localStorage.getItem('token') && localStorage.getItem('user')),
  setUser: (user) => {
    if (user) {
      localStorage.setItem('user', JSON.stringify(user));
    } else {
      localStorage.removeItem('user');
    }
    set({ user, isAuthenticated: !!user && !!get().token });
  },
  setToken: (token) => {
    if (token) {
      localStorage.setItem('token', token);
    } else {
      localStorage.removeItem('token');
    }
    set({ token, isAuthenticated: !!token && !!get().user });
  },
  logout: () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    localStorage.removeItem('userPreferences');
    set({ user: null, token: null, isAuthenticated: false, userPreferences: null });
  },

  // User preferences state
  userPreferences: JSON.parse(localStorage.getItem('userPreferences') || 'null'),
  setUserPreferences: (preferences) => {
    if (preferences) {
      localStorage.setItem('userPreferences', JSON.stringify(preferences));
    } else {
      localStorage.removeItem('userPreferences');
    }
    set({ userPreferences: preferences });
  },

  // System settings state - default to true
  websocketEnabled: localStorage.getItem('websocketEnabled') !== 'false',
  setWebsocketEnabled: (enabled) => {
    localStorage.setItem('websocketEnabled', String(enabled));
    set({ websocketEnabled: enabled });
  },

  // Nodes state
  nodes: [],
  selectedNode: null,
  setNodes: (nodes) => set({ nodes }),
  setSelectedNode: (node) => set({ selectedNode: node }),

  // System stats
  systemStats: null,
  setSystemStats: (stats) => set({ systemStats: stats }),

  // UI state
  sidebarCollapsed: false,
  toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
}));
