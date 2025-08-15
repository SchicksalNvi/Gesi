class Api {
    constructor() {
        this.baseUrl = '';
    }

    async getVersion() {
        const response = await fetch(`${this.baseUrl}/api/version`);
        if (!response.ok) {
            throw new Error(`Failed to get version: ${response.statusText}`);
        }
        return await response.json();
    }

    async getNodes() {
        const response = await fetch(`${this.baseUrl}/api/nodes`);
        if (!response.ok) {
            throw new Error(`Failed to get nodes: ${response.statusText}`);
        }
        return await response.json();
    }

    async getNodeProcesses(nodeName) {
        const response = await fetch(`${this.baseUrl}/api/nodes/${nodeName}/processes`);
        if (!response.ok) {
            throw new Error(`Failed to get node processes: ${response.statusText}`);
        }
        return await response.json();
    }

    async startProcess(nodeName, processName) {
        const response = await fetch(`${this.baseUrl}/api/nodes/${nodeName}/processes/${processName}/start`, {
            method: 'POST'
        });
        if (!response.ok) {
            throw new Error(`Failed to start process: ${response.statusText}`);
        }
        return await response.json();
    }

    async stopProcess(nodeName, processName) {
        const response = await fetch(`${this.baseUrl}/api/nodes/${nodeName}/processes/${processName}/stop`, {
            method: 'POST'
        });
        if (!response.ok) {
            throw new Error(`Failed to stop process: ${response.statusText}`);
        }
        return await response.json();
    }

    async restartProcess(nodeName, processName) {
        const response = await fetch(`${this.baseUrl}/api/nodes/${nodeName}/processes/${processName}/restart`, {
            method: 'POST'
        });
        if (!response.ok) {
            throw new Error(`Failed to restart process: ${response.statusText}`);
        }
        return await response.json();
    }

    async getLogs() {
        const response = await fetch(`${this.baseUrl}/api/logs`);
        if (!response.ok) {
            throw new Error(`Failed to get logs: ${response.statusText}`);
        }
        return await response.json();
    }

    async login(username, password) {
        const response = await fetch(`${this.baseUrl}/api/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ username, password })
        });
        if (!response.ok) {
            throw new Error(`Failed to login: ${response.statusText}`);
        }
        return await response.json();
    }

    async checkAuth() {
        const response = await fetch(`${this.baseUrl}/api/auth/user`);
        if (!response.ok) {
            throw new Error(`Failed to check auth: ${response.statusText}`);
        }
        return await response.json();
    }

    async getCurrentUser() {
        const response = await fetch(`${this.baseUrl}/api/auth/user`);
        if (!response.ok) {
            throw new Error(`Failed to get current user: ${response.statusText}`);
        }
        return await response.json();
    }

    // 用户管理 API
    async getUsers() {
        const response = await fetch(`${this.baseUrl}/api/users`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        if (!response.ok) {
            throw new Error(`Failed to get users: ${response.statusText}`);
        }
        return await response.json();
    }

    async createUser(username, password, isAdmin) {
        const userData = {
            username: username,
            password: password,
            is_admin: isAdmin
        };
        const response = await fetch(`${this.baseUrl}/api/users`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(userData)
        });
        if (!response.ok) {
            throw new Error(`Failed to create user: ${response.statusText}`);
        }
        return await response.json();
    }

    async deleteUser(username) {
        const response = await fetch(`${this.baseUrl}/api/users/${username}`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        if (!response.ok) {
            throw new Error(`Failed to delete user: ${response.statusText}`);
        }
        return await response.json();
    }

    async changePassword(username, oldPassword, newPassword) {
        const passwordData = {
            old_password: oldPassword,
            new_password: newPassword
        };
        const response = await fetch(`${this.baseUrl}/api/users/${username}/password`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(passwordData)
        });
        if (!response.ok) {
            throw new Error(`Failed to change password: ${response.statusText}`);
        }
        return await response.json();
    }

    async logout() {
        const response = await fetch(`${this.baseUrl}/api/auth/logout`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        if (!response.ok) {
            throw new Error(`Failed to logout: ${response.statusText}`);
        }
        return await response.json();
    }

    async getProcessLogs(nodeName, processName) {
        const response = await fetch(`${this.baseUrl}/api/nodes/${nodeName}/processes/${processName}/logs`);
        if (!response.ok) {
            throw new Error(`Failed to get process logs: ${response.statusText}`);
        }
        return await response.json();
    }

    // 环境管理 API
    async getEnvironments() {
        const response = await fetch(`${this.baseUrl}/api/environments`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        if (!response.ok) {
            throw new Error(`Failed to get environments: ${response.statusText}`);
        }
        return await response.json();
    }

    async getEnvironmentDetails(environmentName) {
        const response = await fetch(`${this.baseUrl}/api/environments/${environmentName}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        if (!response.ok) {
            throw new Error(`Failed to get environment details: ${response.statusText}`);
        }
        return await response.json();
    }

    // 分组管理 API
    async getGroups() {
        const response = await fetch(`${this.baseUrl}/api/groups`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        if (!response.ok) {
            throw new Error(`Failed to get groups: ${response.statusText}`);
        }
        return await response.json();
    }

    async getGroupDetails(groupName) {
        const response = await fetch(`${this.baseUrl}/api/groups/${groupName}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        if (!response.ok) {
            throw new Error(`Failed to get group details: ${response.statusText}`);
        }
        return await response.json();
    }

    async startGroupProcesses(groupName, environment = '') {
        const url = environment ? 
            `${this.baseUrl}/api/groups/${groupName}/start?environment=${environment}` : 
            `${this.baseUrl}/api/groups/${groupName}/start`;
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        if (!response.ok) {
            throw new Error(`Failed to start group processes: ${response.statusText}`);
        }
        return await response.json();
    }

    async stopGroupProcesses(groupName, environment = '') {
        const url = environment ? 
            `${this.baseUrl}/api/groups/${groupName}/stop?environment=${environment}` : 
            `${this.baseUrl}/api/groups/${groupName}/stop`;
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        if (!response.ok) {
            throw new Error(`Failed to stop group processes: ${response.statusText}`);
        }
        return await response.json();
    }

    async restartGroupProcesses(groupName, environment = '') {
        const url = environment ? 
            `${this.baseUrl}/api/groups/${groupName}/restart?environment=${environment}` : 
            `${this.baseUrl}/api/groups/${groupName}/restart`;
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        if (!response.ok) {
            throw new Error(`Failed to restart group processes: ${response.statusText}`);
        }
        return await response.json();
    }

    // Activity Logs API
    async getActivityLogs(params = {}) {
        const queryString = new URLSearchParams(params).toString();
        const url = queryString ? `${this.baseUrl}/api/activity-logs?${queryString}` : `${this.baseUrl}/api/activity-logs`;
        const response = await fetch(url, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        if (!response.ok) {
            throw new Error(`Failed to get activity logs: ${response.statusText}`);
        }
        return await response.json();
    }

    async getRecentLogs(limit = 20) {
        const response = await fetch(`${this.baseUrl}/api/activity-logs/recent?limit=${limit}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        if (!response.ok) {
            throw new Error(`Failed to get recent logs: ${response.statusText}`);
        }
        return await response.json();
    }

    async getLogStatistics(days = 7) {
        const response = await fetch(`${this.baseUrl}/api/activity-logs/statistics?days=${days}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        if (!response.ok) {
            throw new Error(`Failed to get log statistics: ${response.statusText}`);
        }
        return await response.json();
    }

    async cleanOldLogs(days = 90) {
        const response = await fetch(`${this.baseUrl}/api/activity-logs/clean?days=${days}`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        if (!response.ok) {
            throw new Error(`Failed to clean old logs: ${response.statusText}`);
        }
        return await response.json();
    }
}

// 创建全局实例
window.Api = Api;
window.api = new Api();

// 刷新数据函数
window.refreshData = async function() {
    try {
        const nodes = await window.api.getNodes();
        const logsResponse = await window.api.getLogs();
        const logs = logsResponse.logs || [];
        return { nodes, logs };
    } catch (error) {
        console.error('Error refreshing data:', error);
        throw error;
    }
};