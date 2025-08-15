class LogsManager {
    constructor() {
        // 确保 Api 类已加载
        if (!window.Api || !window.api) {
            throw new Error('Api class is not loaded. Please check if api.js is properly included.');
        }
        this.api = window.api;
        this.currentPage = 1;
        this.pageSize = 20;
        this.hasMoreLogs = true;
        this.autoRefreshInterval = null;
        this.filters = {
            node: '',
            process: '',
            level: ''
        };

        this.initializeElements();
        this.bindEvents();
        this.initializeFilters();
        this.loadLogs();
        this.startAutoRefresh();
    }

    initializeElements() {
        // 过滤器元素
        this.nodeFilter = document.getElementById('node-filter');
        this.processFilter = document.getElementById('process-filter');
        this.levelFilter = document.getElementById('level-filter');
        this.autoRefreshCheckbox = document.getElementById('auto-refresh');
        this.refreshButton = document.getElementById('refresh-btn');
        this.clearFiltersButton = document.getElementById('clear-filters-btn');
        this.logoutBtn = document.getElementById('logout-btn');

        // 日志容器
        this.logsContainer = document.getElementById('logs-container');
        this.loadMoreContainer = document.getElementById('load-more');
        this.loadMoreButton = this.loadMoreContainer.querySelector('button');
    }

    bindEvents() {
        // 过滤器事件
        this.nodeFilter.addEventListener('change', () => this.handleFilterChange());
        this.processFilter.addEventListener('change', () => this.handleFilterChange());
        this.levelFilter.addEventListener('change', () => this.handleFilterChange());
        
        // 自动刷新事件
        this.autoRefreshCheckbox.addEventListener('change', () => {
            if (this.autoRefreshCheckbox.checked) {
                this.startAutoRefresh();
            } else {
                this.stopAutoRefresh();
            }
        });

        // 刷新按钮事件
        this.refreshButton.addEventListener('click', () => this.loadLogs());

        // 清除过滤器事件
        this.clearFiltersButton.addEventListener('click', () => this.clearFilters());

        // 加载更多事件
        this.loadMoreButton.addEventListener('click', () => this.loadMoreLogs());

        // 滚动加载
        window.addEventListener('scroll', () => {
            if (this.isNearBottom() && this.hasMoreLogs) {
                this.loadMoreLogs();
            }
        });

        // 退出登录事件
        if (this.logoutBtn) {
            this.logoutBtn.addEventListener('click', () => this.handleLogout());
        }
    }

    async initializeFilters() {
        try {
            // 获取所有节点
            const nodes = await this.api.getNodes();
            
            // 填充节点过滤器
            nodes.forEach(node => {
                const option = document.createElement('option');
                option.value = node.name;
                option.textContent = node.name;
                this.nodeFilter.appendChild(option);
            });

            // 获取所有进程并填充进程过滤器
            for (const node of nodes) {
                const processes = await this.api.getNodeProcesses(node.name);
                processes.forEach(process => {
                    // 检查是否已存在相同的进程选项
                    if (!this.processFilter.querySelector(`option[value="${process.name}"]`)) {
                        const option = document.createElement('option');
                        option.value = process.name;
                        option.textContent = process.name;
                        this.processFilter.appendChild(option);
                    }
                });
            }
        } catch (error) {
            this.showAlert('Error initializing filters: ' + error.message, 'danger');
        }
    }

    async loadLogs(reset = true) {
        try {
            if (reset) {
                this.currentPage = 1;
                this.logsContainer.innerHTML = '';
                this.hasMoreLogs = true;
            }

            const response = await this.api.getLogs({
                page: this.currentPage,
                pageSize: this.pageSize,
                node: this.filters.node,
                process: this.filters.process,
                level: this.filters.level
            });

            const logs = response.logs || [];

            if (logs.length < this.pageSize) {
                this.hasMoreLogs = false;
                this.loadMoreContainer.style.display = 'none';
            } else {
                this.loadMoreContainer.style.display = 'block';
            }

            this.renderLogs(logs, reset);
        } catch (error) {
            this.showAlert('Error loading logs: ' + error.message, 'danger');
        }
    }

    async loadMoreLogs() {
        if (!this.hasMoreLogs) return;
        
        this.currentPage++;
        await this.loadLogs(false);
    }

    renderLogs(logs, reset = true) {
        if (reset) {
            this.logsContainer.innerHTML = '';
        }

        logs.forEach(log => {
            const logEntry = document.createElement('div');
            logEntry.className = 'log-entry';
            logEntry.innerHTML = `
                <div class="d-flex justify-content-between align-items-start">
                    <div>
                        <span class="log-level log-level-${log.level.toLowerCase()}">${log.level}</span>
                        <span class="ms-2">${log.node} / ${log.process}</span>
                    </div>
                    <span class="log-time">${this.formatTimestamp(log.timestamp)}</span>
                </div>
                <div class="log-message">${log.message}</div>
                ${log.details ? `<div class="log-details text-muted">${log.details}</div>` : ''}
            `;
            this.logsContainer.appendChild(logEntry);
        });
    }

    handleFilterChange() {
        this.filters = {
            node: this.nodeFilter.value,
            process: this.processFilter.value,
            level: this.levelFilter.value
        };
        this.loadLogs();
    }

    clearFilters() {
        this.nodeFilter.value = '';
        this.processFilter.value = '';
        this.levelFilter.value = '';
        this.filters = {
            node: '',
            process: '',
            level: ''
        };
        this.loadLogs();
    }

    startAutoRefresh() {
        if (this.autoRefreshInterval) {
            clearInterval(this.autoRefreshInterval);
        }
        this.autoRefreshInterval = setInterval(() => this.loadLogs(), 10000); // 每10秒刷新一次
    }

    stopAutoRefresh() {
        if (this.autoRefreshInterval) {
            clearInterval(this.autoRefreshInterval);
            this.autoRefreshInterval = null;
        }
    }

    isNearBottom() {
        return window.innerHeight + window.scrollY >= document.documentElement.scrollHeight - 100;
    }

    formatTimestamp(timestamp) {
        const date = new Date(timestamp);
        return date.toLocaleString();
    }

    async handleLogout() {
        try {
            await this.api.logout();
            window.location.href = '/login';
        } catch (error) {
            this.showAlert('退出登录失败', 'danger');
        }
    }

    showAlert(message, type = 'info') {
        const alertContainer = document.getElementById('alert-container');
        const alert = document.createElement('div');
        alert.className = `alert alert-${type} alert-dismissible fade show`;
        alert.innerHTML = `
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        `;
        alertContainer.appendChild(alert);

        // 5秒后自动关闭提示
        setTimeout(() => {
            alert.remove();
        }, 5000);
    }
}

// 初始化日志管理器
document.addEventListener('DOMContentLoaded', () => {
    new LogsManager();
});