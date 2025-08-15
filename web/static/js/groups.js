class GroupManager {
    constructor() {
        this.groups = [];
        this.currentGroup = null;
        this.confirmActionModal = null;
        this.groupDetailsModal = null;
        this.logoutBtn = document.getElementById('logout-btn');
        this.init();
    }

    init() {
        // 初始化模态框
        this.confirmActionModal = new bootstrap.Modal(document.getElementById('confirmActionModal'));
        this.groupDetailsModal = new bootstrap.Modal(document.getElementById('groupDetailsModal'));
        
        // 绑定确认按钮事件
        document.getElementById('confirmActionBtn').addEventListener('click', () => {
            this.executeConfirmedAction();
        });
        
        // 绑定退出登录按钮事件
        if (this.logoutBtn) {
            this.logoutBtn.addEventListener('click', () => this.handleLogout());
        }
        
        // 加载分组数据
        this.loadGroups();
    }

    async loadGroups() {
        try {
            this.showLoading(true);
            const response = await api.getGroups();
            
            if (response.status === 'success') {
                this.groups = response.groups || [];
                this.renderGroups();
            } else {
                this.showAlert('Error loading groups: ' + (response.message || 'Unknown error'), 'danger');
            }
        } catch (error) {
            console.error('Error loading groups:', error);
            this.showAlert('Failed to load groups. Please try again.', 'danger');
        } finally {
            this.showLoading(false);
        }
    }

    renderGroups() {
        const container = document.getElementById('groupsContainer');
        
        if (this.groups.length === 0) {
            container.innerHTML = `
                <div class="col-12">
                    <div class="alert alert-info text-center">
                        <i class="fas fa-info-circle me-2"></i>
                        No process groups found.
                    </div>
                </div>
            `;
            return;
        }

        container.innerHTML = this.groups.map(group => this.createGroupCard(group)).join('');
    }

    createGroupCard(group) {
        const totalProcesses = this.getTotalProcesses(group);
        const runningProcesses = this.getRunningProcesses(group);
        const environments = group.environments || [];
        
        return `
            <div class="col-md-6 col-lg-4 mb-4">
                <div class="card group-card h-100">
                    <div class="card-header d-flex justify-content-between align-items-center">
                        <h5 class="card-title mb-0">
                            <i class="fas fa-object-group me-2"></i>${group.name}
                        </h5>
                        <div class="dropdown">
                            <button class="btn btn-sm btn-outline-secondary dropdown-toggle" type="button" data-bs-toggle="dropdown">
                                <i class="fas fa-cog"></i>
                            </button>
                            <ul class="dropdown-menu">
                                <li><a class="dropdown-item" href="#" onclick="groupManager.showGroupDetails('${group.name}')">
                                    <i class="fas fa-info-circle me-1"></i>Details
                                </a></li>
                                <li><hr class="dropdown-divider"></li>
                                <li><a class="dropdown-item text-success" href="#" onclick="groupManager.confirmGroupAction('${group.name}', 'start')">
                                    <i class="fas fa-play me-1"></i>Start All
                                </a></li>
                                <li><a class="dropdown-item text-warning" href="#" onclick="groupManager.confirmGroupAction('${group.name}', 'restart')">
                                    <i class="fas fa-redo me-1"></i>Restart All
                                </a></li>
                                <li><a class="dropdown-item text-danger" href="#" onclick="groupManager.confirmGroupAction('${group.name}', 'stop')">
                                    <i class="fas fa-stop me-1"></i>Stop All
                                </a></li>
                            </ul>
                        </div>
                    </div>
                    <div class="card-body">
                        <div class="row mb-3">
                            <div class="col-6">
                                <div class="text-center">
                                    <div class="h4 mb-0 text-primary">${totalProcesses}</div>
                                    <small class="text-muted">Total Processes</small>
                                </div>
                            </div>
                            <div class="col-6">
                                <div class="text-center">
                                    <div class="h4 mb-0 text-success">${runningProcesses}</div>
                                    <small class="text-muted">Running</small>
                                </div>
                            </div>
                        </div>
                        
                        <div class="mb-3">
                            <h6 class="text-muted mb-2">Environments:</h6>
                            ${environments.map(env => `
                                <div class="d-flex justify-content-between align-items-center mb-2">
                                    <span class="badge bg-secondary">${env.name}</span>
                                    <div>
                                        <span class="badge bg-info me-1">${env.processes.length} processes</span>
                                        <span class="badge bg-light text-dark">${env.members.length} nodes</span>
                                    </div>
                                </div>
                            `).join('')}
                        </div>
                        
                        <div class="d-grid gap-2">
                            <button class="btn btn-outline-primary btn-sm" onclick="groupManager.showGroupDetails('${group.name}')">
                                <i class="fas fa-eye me-1"></i>View Details
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }

    getTotalProcesses(group) {
        return (group.environments || []).reduce((total, env) => total + (env.processes || []).length, 0);
    }

    getRunningProcesses(group) {
        return (group.environments || []).reduce((total, env) => {
            return total + (env.processes || []).filter(p => p.state === 'RUNNING').length;
        }, 0);
    }

    async showGroupDetails(groupName) {
        try {
            const response = await api.getGroupDetails(groupName);
            
            if (response.status === 'success' && response.group) {
                this.renderGroupDetails(response.group);
                document.getElementById('modalGroupName').textContent = groupName;
                this.groupDetailsModal.show();
            } else {
                this.showAlert('Error loading group details: ' + (response.message || 'Unknown error'), 'danger');
            }
        } catch (error) {
            console.error('Error loading group details:', error);
            this.showAlert('Failed to load group details. Please try again.', 'danger');
        }
    }

    renderGroupDetails(group) {
        const container = document.getElementById('groupDetailsContent');
        const environments = group.environments || [];
        
        container.innerHTML = `
            <div class="row mb-4">
                <div class="col-md-4">
                    <div class="card bg-primary text-white">
                        <div class="card-body text-center">
                            <h3>${this.getTotalProcesses(group)}</h3>
                            <p class="mb-0">Total Processes</p>
                        </div>
                    </div>
                </div>
                <div class="col-md-4">
                    <div class="card bg-success text-white">
                        <div class="card-body text-center">
                            <h3>${this.getRunningProcesses(group)}</h3>
                            <p class="mb-0">Running</p>
                        </div>
                    </div>
                </div>
                <div class="col-md-4">
                    <div class="card bg-info text-white">
                        <div class="card-body text-center">
                            <h3>${environments.length}</h3>
                            <p class="mb-0">Environments</p>
                        </div>
                    </div>
                </div>
            </div>
            
            ${environments.map(env => `
                <div class="environment-section">
                    <div class="d-flex justify-content-between align-items-center mb-3">
                        <h5 class="mb-0">
                            <i class="fas fa-layer-group me-2"></i>${env.name}
                            <span class="badge bg-secondary ms-2">${env.processes.length} processes</span>
                        </h5>
                        <div class="btn-group btn-group-sm">
                            <button class="btn btn-success" onclick="groupManager.confirmGroupAction('${group.name}', 'start', '${env.name}')">
                                <i class="fas fa-play"></i>
                            </button>
                            <button class="btn btn-warning" onclick="groupManager.confirmGroupAction('${group.name}', 'restart', '${env.name}')">
                                <i class="fas fa-redo"></i>
                            </button>
                            <button class="btn btn-danger" onclick="groupManager.confirmGroupAction('${group.name}', 'stop', '${env.name}')">
                                <i class="fas fa-stop"></i>
                            </button>
                        </div>
                    </div>
                    
                    <div class="row">
                        ${(env.processes || []).map(process => `
                            <div class="col-md-6 col-lg-4 mb-2">
                                <div class="process-item">
                                    <div class="d-flex justify-content-between align-items-center">
                                        <div>
                                            <strong>${process.name}</strong>
                                            <br>
                                            <small class="text-muted">Node: ${process.node}</small>
                                        </div>
                                        <div class="text-end">
                                            <span class="badge process-badge ${this.getStatusClass(process.state)}">
                                                ${process.state}
                                            </span>
                                            ${process.pid ? `<br><small class="text-muted">PID: ${process.pid}</small>` : ''}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        `).join('')}
                    </div>
                </div>
            `).join('')}
        `;
    }

    getStatusClass(state) {
        switch (state) {
            case 'RUNNING':
                return 'bg-success';
            case 'STOPPED':
                return 'bg-secondary';
            case 'STARTING':
                return 'bg-warning';
            case 'FATAL':
                return 'bg-danger';
            default:
                return 'bg-secondary';
        }
    }

    confirmGroupAction(groupName, action, environment = '') {
        this.pendingAction = { groupName, action, environment };
        
        const actionText = action.charAt(0).toUpperCase() + action.slice(1);
        const envText = environment ? ` in environment "${environment}"` : '';
        
        document.getElementById('confirmMessage').textContent = 
            `Are you sure you want to ${actionText.toLowerCase()} all processes in group "${groupName}"${envText}?`;
        
        this.confirmActionModal.show();
    }

    async executeConfirmedAction() {
        if (!this.pendingAction) return;
        
        const { groupName, action, environment } = this.pendingAction;
        
        try {
            this.confirmActionModal.hide();
            
            let response;
            switch (action) {
                case 'start':
                    response = await api.startGroupProcesses(groupName, environment);
                    break;
                case 'stop':
                    response = await api.stopGroupProcesses(groupName, environment);
                    break;
                case 'restart':
                    response = await api.restartGroupProcesses(groupName, environment);
                    break;
                default:
                    throw new Error('Unknown action: ' + action);
            }
            
            if (response.status === 'success') {
                this.showAlert(`Group processes ${action}ed successfully!`, 'success');
                // 刷新数据
                setTimeout(() => {
                    this.loadGroups();
                    if (this.groupDetailsModal._isShown) {
                        this.showGroupDetails(groupName);
                    }
                }, 1000);
            } else {
                this.showAlert('Error: ' + (response.message || 'Unknown error'), 'danger');
            }
        } catch (error) {
            console.error('Error executing group action:', error);
            this.showAlert('Failed to execute action. Please try again.', 'danger');
        } finally {
            this.pendingAction = null;
        }
    }

    showLoading(show) {
        const spinner = document.getElementById('loadingSpinner');
        const container = document.getElementById('groupsContainer');
        
        if (show) {
            spinner.style.display = 'block';
            container.style.display = 'none';
        } else {
            spinner.style.display = 'none';
            container.style.display = 'block';
        }
    }

    async handleLogout() {
        try {
            await api.logout();
            window.location.href = '/login';
        } catch (error) {
            this.showAlert('退出登录失败', 'danger');
        }
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
        
        // 自动移除警告
        setTimeout(() => {
            const alertElement = document.getElementById(alertId);
            if (alertElement) {
                const alert = new bootstrap.Alert(alertElement);
                alert.close();
            }
        }, 5000);
    }
}

// 全局函数
function refreshGroups() {
    groupManager.loadGroups();
}

function logout() {
    api.logout().then(() => {
        window.location.href = '/login';
    }).catch(error => {
        console.error('Logout error:', error);
        window.location.href = '/login';
    });
}

// 初始化
let groupManager;
document.addEventListener('DOMContentLoaded', function() {
    groupManager = new GroupManager();
});