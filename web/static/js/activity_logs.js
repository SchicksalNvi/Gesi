class ActivityLogManager {
    constructor() {
        this.currentPage = 1;
        this.pageSize = 50;
        this.filters = {};
        this.cleanLogsModal = null;
        this.init();
    }

    init() {
        // 初始化模态框
        this.cleanLogsModal = new bootstrap.Modal(document.getElementById('cleanLogsModal'));
        
        // 绑定事件
        this.bindEvents();
        
        // 加载初始数据
        this.loadStatistics();
        this.loadLogs();
        
        // 检查管理员权限
        this.checkAdminPermissions();
    }

    bindEvents() {
        // 页面大小选择
        document.getElementById('pageSizeSelect').addEventListener('change', (e) => {
            this.pageSize = parseInt(e.target.value);
            this.currentPage = 1;
            this.loadLogs();
        });

        // 过滤器输入事件
        const filterInputs = ['levelFilter', 'actionFilter', 'resourceFilter', 'usernameFilter', 'startTimeFilter', 'endTimeFilter'];
        filterInputs.forEach(id => {
            const element = document.getElementById(id);
            if (element) {
                element.addEventListener('keypress', (e) => {
                    if (e.key === 'Enter') {
                        this.applyFilters();
                    }
                });
            }
        });
        
        // 退出登录事件
        if (this.logoutBtn) {
            this.logoutBtn.addEventListener('click', () => this.handleLogout());
        }
    }

    async loadStatistics() {
        try {
            const response = await api.getLogStatistics();
            if (response.status === 'success') {
                this.renderStatistics(response.data);
            }
        } catch (error) {
            console.error('Failed to load statistics:', error);
        }
    }

    renderStatistics(stats) {
        const container = document.getElementById('statsContainer');
        
        const statsHtml = `
            <div class="col-md-3">
                <div class="card stats-card">
                    <div class="card-body text-center">
                        <i class="fas fa-list-alt fa-2x mb-2"></i>
                        <h4>${stats.total_logs || 0}</h4>
                        <p class="mb-0">Total Logs</p>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card bg-info text-white">
                    <div class="card-body text-center">
                        <i class="fas fa-info-circle fa-2x mb-2"></i>
                        <h4>${stats.info_count || 0}</h4>
                        <p class="mb-0">Info</p>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card bg-warning text-white">
                    <div class="card-body text-center">
                        <i class="fas fa-exclamation-triangle fa-2x mb-2"></i>
                        <h4>${stats.warning_count || 0}</h4>
                        <p class="mb-0">Warnings</p>
                    </div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="card bg-danger text-white">
                    <div class="card-body text-center">
                        <i class="fas fa-times-circle fa-2x mb-2"></i>
                        <h4>${stats.error_count || 0}</h4>
                        <p class="mb-0">Errors</p>
                    </div>
                </div>
            </div>
        `;
        
        container.innerHTML = statsHtml;
    }

    async loadLogs() {
        try {
            this.showLoading();
            
            const params = {
                page: this.currentPage,
                page_size: this.pageSize,
                ...this.filters
            };
            
            const response = await api.getActivityLogs(params);
            if (response.status === 'success') {
                this.renderLogs(response.data.logs);
                this.renderPagination(response.data.pagination);
            } else {
                this.showAlert('Failed to load logs: ' + response.message, 'danger');
            }
        } catch (error) {
            console.error('Failed to load logs:', error);
            this.showAlert('Failed to load logs', 'danger');
        } finally {
            this.hideLoading();
        }
    }

    renderLogs(logs) {
        const container = document.getElementById('logsContainer');
        
        if (!logs || logs.length === 0) {
            container.innerHTML = `
                <div class="text-center py-4">
                    <i class="fas fa-inbox fa-3x text-muted mb-3"></i>
                    <p class="text-muted">No logs found</p>
                </div>
            `;
            return;
        }

        const logsHtml = logs.map(log => this.createLogItem(log)).join('');
        container.innerHTML = logsHtml;
    }

    createLogItem(log) {
        const levelClass = `level-${log.level.toLowerCase()}`;
        const levelIcon = this.getLevelIcon(log.level);
        const levelColor = this.getLevelColor(log.level);
        
        const createdAt = new Date(log.created_at).toLocaleString();
        
        return `
            <div class="log-item ${levelClass} p-3 mb-2 border rounded">
                <div class="row align-items-start">
                    <div class="col-auto">
                        <i class="${levelIcon}" style="color: ${levelColor};"></i>
                    </div>
                    <div class="col">
                        <div class="d-flex justify-content-between align-items-start mb-1">
                            <div>
                                <span class="badge bg-secondary me-2">${log.level}</span>
                                <span class="badge bg-primary me-2">${log.action}</span>
                                ${log.resource ? `<span class="badge bg-info">${log.resource}</span>` : ''}
                            </div>
                            <small class="text-muted">${createdAt}</small>
                        </div>
                        <div class="mb-2">
                            <strong>${log.message}</strong>
                        </div>
                        <div class="row text-sm">
                            ${log.username ? `<div class="col-auto"><i class="fas fa-user me-1"></i>${log.username}</div>` : ''}
                            ${log.target ? `<div class="col-auto"><i class="fas fa-bullseye me-1"></i>${log.target}</div>` : ''}
                            ${log.ip_address ? `<div class="col-auto"><i class="fas fa-globe me-1"></i>${log.ip_address}</div>` : ''}
                            ${log.user_agent ? `<div class="col-12 mt-1"><i class="fas fa-desktop me-1"></i><small class="text-muted">${log.user_agent}</small></div>` : ''}
                        </div>
                        ${log.extra_info ? `
                            <div class="mt-2">
                                <button class="btn btn-sm btn-outline-secondary" type="button" data-bs-toggle="collapse" data-bs-target="#extra-${log.id}">
                                    <i class="fas fa-info-circle me-1"></i>Details
                                </button>
                                <div class="collapse mt-2" id="extra-${log.id}">
                                    <pre class="bg-light p-2 rounded"><code>${JSON.stringify(log.extra_info, null, 2)}</code></pre>
                                </div>
                            </div>
                        ` : ''}
                    </div>
                </div>
            </div>
        `;
    }

    getLevelIcon(level) {
        const icons = {
            'INFO': 'fas fa-info-circle',
            'WARNING': 'fas fa-exclamation-triangle',
            'ERROR': 'fas fa-times-circle',
            'DEBUG': 'fas fa-bug'
        };
        return icons[level] || 'fas fa-circle';
    }

    getLevelColor(level) {
        const colors = {
            'INFO': '#0dcaf0',
            'WARNING': '#ffc107',
            'ERROR': '#dc3545',
            'DEBUG': '#6c757d'
        };
        return colors[level] || '#6c757d';
    }

    renderPagination(pagination) {
        const container = document.getElementById('pagination');
        
        if (pagination.total_pages <= 1) {
            container.innerHTML = '';
            return;
        }

        let paginationHtml = '';
        
        // Previous button
        if (pagination.has_prev) {
            paginationHtml += `
                <li class="page-item">
                    <a class="page-link" href="#" onclick="activityLogManager.goToPage(${pagination.page - 1})">
                        <i class="fas fa-chevron-left"></i>
                    </a>
                </li>
            `;
        }
        
        // Page numbers
        const startPage = Math.max(1, pagination.page - 2);
        const endPage = Math.min(pagination.total_pages, pagination.page + 2);
        
        if (startPage > 1) {
            paginationHtml += `
                <li class="page-item">
                    <a class="page-link" href="#" onclick="activityLogManager.goToPage(1)">1</a>
                </li>
            `;
            if (startPage > 2) {
                paginationHtml += '<li class="page-item disabled"><span class="page-link">...</span></li>';
            }
        }
        
        for (let i = startPage; i <= endPage; i++) {
            const activeClass = i === pagination.page ? 'active' : '';
            paginationHtml += `
                <li class="page-item ${activeClass}">
                    <a class="page-link" href="#" onclick="activityLogManager.goToPage(${i})">${i}</a>
                </li>
            `;
        }
        
        if (endPage < pagination.total_pages) {
            if (endPage < pagination.total_pages - 1) {
                paginationHtml += '<li class="page-item disabled"><span class="page-link">...</span></li>';
            }
            paginationHtml += `
                <li class="page-item">
                    <a class="page-link" href="#" onclick="activityLogManager.goToPage(${pagination.total_pages})">${pagination.total_pages}</a>
                </li>
            `;
        }
        
        // Next button
        if (pagination.has_next) {
            paginationHtml += `
                <li class="page-item">
                    <a class="page-link" href="#" onclick="activityLogManager.goToPage(${pagination.page + 1})">
                        <i class="fas fa-chevron-right"></i>
                    </a>
                </li>
            `;
        }
        
        container.innerHTML = paginationHtml;
    }

    goToPage(page) {
        this.currentPage = page;
        this.loadLogs();
    }

    applyFilters() {
        this.filters = {
            level: document.getElementById('levelFilter').value,
            action: document.getElementById('actionFilter').value,
            resource: document.getElementById('resourceFilter').value,
            username: document.getElementById('usernameFilter').value,
            start_time: document.getElementById('startTimeFilter').value,
            end_time: document.getElementById('endTimeFilter').value
        };
        
        // 移除空值
        Object.keys(this.filters).forEach(key => {
            if (!this.filters[key]) {
                delete this.filters[key];
            }
        });
        
        this.currentPage = 1;
        this.loadLogs();
    }

    clearFilters() {
        document.getElementById('levelFilter').value = '';
        document.getElementById('actionFilter').value = '';
        document.getElementById('resourceFilter').value = '';
        document.getElementById('usernameFilter').value = '';
        document.getElementById('startTimeFilter').value = '';
        document.getElementById('endTimeFilter').value = '';
        
        this.filters = {};
        this.currentPage = 1;
        this.loadLogs();
    }

    refreshLogs() {
        this.loadLogs();
        this.loadStatistics();
    }

    async checkAdminPermissions() {
        try {
            const user = await api.getCurrentUser();
            if (!user.is_admin) {
                document.getElementById('cleanLogsBtn').style.display = 'none';
            }
        } catch (error) {
            console.error('Failed to check permissions:', error);
            document.getElementById('cleanLogsBtn').style.display = 'none';
        }
    }

    showCleanModal() {
        this.cleanLogsModal.show();
    }

    async cleanOldLogs() {
        const days = parseInt(document.getElementById('cleanDays').value);
        
        if (days < 7 || days > 365) {
            this.showAlert('Days must be between 7 and 365', 'warning');
            return;
        }
        
        try {
            const response = await api.cleanOldLogs(days);
            if (response.status === 'success') {
                this.showAlert('Old logs cleaned successfully', 'success');
                this.cleanLogsModal.hide();
                this.loadLogs();
                this.loadStatistics();
            } else {
                this.showAlert('Failed to clean logs: ' + response.message, 'danger');
            }
        } catch (error) {
            console.error('Failed to clean logs:', error);
            this.showAlert('Failed to clean logs', 'danger');
        }
    }

    showLoading() {
        const container = document.getElementById('logsContainer');
        container.innerHTML = `
            <div class="text-center py-4">
                <div class="spinner-border text-primary" role="status">
                    <span class="visually-hidden">Loading...</span>
                </div>
                <p class="mt-2 text-muted">Loading logs...</p>
            </div>
        `;
    }

    hideLoading() {
        // Loading will be hidden when logs are rendered
    }

    showAlert(message, type = 'info') {
        const alertContainer = document.getElementById('alertContainer');
        const alertId = 'alert-' + Date.now();
        
        const alertHtml = `
            <div id="${alertId}" class="alert alert-${type} alert-dismissible fade show" role="alert">
                ${message}
                <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
            </div>
        `;
        
        alertContainer.insertAdjacentHTML('beforeend', alertHtml);
        
        // Auto remove after 5 seconds
        setTimeout(() => {
            const alertElement = document.getElementById(alertId);
            if (alertElement) {
                const alert = bootstrap.Alert.getOrCreateInstance(alertElement);
                alert.close();
            }
        }, 5000);
    }

    async handleLogout() {
        try {
            await this.api.logout();
            window.location.href = '/login';
        } catch (error) {
            this.showAlert('退出登录失败', 'danger');
        }
    }
}

// Global functions
function applyFilters() {
    activityLogManager.applyFilters();
}

function clearFilters() {
    activityLogManager.clearFilters();
}

function refreshLogs() {
    activityLogManager.refreshLogs();
}

function showCleanModal() {
    activityLogManager.showCleanModal();
}

function cleanOldLogs() {
    activityLogManager.cleanOldLogs();
}

function logout() {
    activityLogManager.logout();
}

// Initialize when page loads
let activityLogManager;
document.addEventListener('DOMContentLoaded', function() {
    activityLogManager = new ActivityLogManager();
});