class EnvironmentManager {
    constructor() {
        // 确保 Api 类已加载
        if (!window.Api || !window.api) {
            throw new Error('Api class is not loaded. Please check if api.js is properly included.');
        }
        this.api = window.api;
        this.initializeElements();
        this.bindEvents();
        this.loadEnvironments();
    }

    initializeElements() {
        this.environmentsContainer = document.getElementById('environments-container');
        this.refreshBtn = document.getElementById('refresh-btn');
        this.logoutBtn = document.getElementById('logout-btn');
        this.environmentModal = new bootstrap.Modal(document.getElementById('environmentModal'));
        this.environmentModalTitle = document.getElementById('environmentModalTitle');
        this.environmentModalBody = document.getElementById('environmentModalBody');
    }

    bindEvents() {
        this.refreshBtn.addEventListener('click', () => this.loadEnvironments());
        this.logoutBtn.addEventListener('click', () => this.handleLogout());
    }

    async loadEnvironments() {
        try {
            this.refreshBtn.disabled = true;
            this.refreshBtn.innerHTML = '<i class="bi bi-hourglass-split"></i> 加载中...';

            const response = await this.api.getEnvironments();
            if (response.status === 'success') {
                this.renderEnvironments(response.environments);
            } else {
                this.showAlert(response.message || '加载环境列表失败', 'danger');
            }
        } catch (error) {
            this.showAlert(error.message || '加载环境列表失败', 'danger');
        } finally {
            this.refreshBtn.disabled = false;
            this.refreshBtn.innerHTML = '<i class="bi bi-arrow-clockwise"></i> 刷新';
        }
    }

    renderEnvironments(environments) {
        this.environmentsContainer.innerHTML = '';

        if (!environments || environments.length === 0) {
            this.environmentsContainer.innerHTML = `
                <div class="col-12">
                    <div class="alert alert-info text-center">
                        <i class="bi bi-info-circle"></i> 暂无环境数据
                    </div>
                </div>
            `;
            return;
        }

        environments.forEach(environment => {
            const environmentCard = this.createEnvironmentCard(environment);
            this.environmentsContainer.appendChild(environmentCard);
        });
    }

    createEnvironmentCard(environment) {
        const col = document.createElement('div');
        col.className = 'col-md-6 col-lg-4 mb-4';

        const connectedNodes = environment.members.filter(member => member.is_connected).length;
        const totalNodes = environment.members.length;
        const healthPercentage = totalNodes > 0 ? Math.round((connectedNodes / totalNodes) * 100) : 0;
        const healthClass = healthPercentage >= 80 ? 'success' : healthPercentage >= 50 ? 'warning' : 'danger';

        col.innerHTML = `
            <div class="card h-100">
                <div class="card-header d-flex justify-content-between align-items-center">
                    <h5 class="card-title mb-0">
                        <i class="bi bi-globe"></i> ${environment.name}
                    </h5>
                    <span class="badge bg-${healthClass}">${healthPercentage}% 健康</span>
                </div>
                <div class="card-body">
                    <div class="row">
                        <div class="col-6">
                            <div class="text-center">
                                <h4 class="text-primary">${totalNodes}</h4>
                                <small class="text-muted">总节点数</small>
                            </div>
                        </div>
                        <div class="col-6">
                            <div class="text-center">
                                <h4 class="text-success">${connectedNodes}</h4>
                                <small class="text-muted">在线节点</small>
                            </div>
                        </div>
                    </div>
                    
                    <div class="progress mt-3" style="height: 8px;">
                        <div class="progress-bar bg-${healthClass}" role="progressbar" 
                             style="width: ${healthPercentage}%" 
                             aria-valuenow="${healthPercentage}" 
                             aria-valuemin="0" 
                             aria-valuemax="100">
                        </div>
                    </div>
                    
                    <div class="mt-3">
                        <h6>节点列表:</h6>
                        <div class="node-list">
                            ${environment.members.slice(0, 3).map(member => `
                                <span class="badge bg-${member.is_connected ? 'success' : 'secondary'} me-1 mb-1">
                                    <i class="bi bi-${member.is_connected ? 'check-circle' : 'x-circle'}"></i>
                                    ${member.name}
                                </span>
                            `).join('')}
                            ${environment.members.length > 3 ? `<span class="text-muted">... 还有 ${environment.members.length - 3} 个节点</span>` : ''}
                        </div>
                    </div>
                </div>
                <div class="card-footer">
                    <button class="btn btn-outline-primary btn-sm w-100" 
                            onclick="environmentManager.showEnvironmentDetails('${environment.name}')">
                        <i class="bi bi-eye"></i> 查看详情
                    </button>
                </div>
            </div>
        `;

        return col;
    }

    async showEnvironmentDetails(environmentName) {
        try {
            const response = await this.api.getEnvironmentDetails(environmentName);
            if (response.status === 'success') {
                this.renderEnvironmentDetails(response.environment);
                this.environmentModal.show();
            } else {
                this.showAlert(response.message || '获取环境详情失败', 'danger');
            }
        } catch (error) {
            this.showAlert(error.message || '获取环境详情失败', 'danger');
        }
    }

    renderEnvironmentDetails(environment) {
        this.environmentModalTitle.textContent = `环境详情 - ${environment.name}`;
        
        const connectedNodes = environment.members.filter(member => member.is_connected).length;
        const totalNodes = environment.members.length;
        
        this.environmentModalBody.innerHTML = `
            <div class="row mb-3">
                <div class="col-md-4">
                    <div class="card text-center">
                        <div class="card-body">
                            <h5 class="card-title text-primary">${totalNodes}</h5>
                            <p class="card-text">总节点数</p>
                        </div>
                    </div>
                </div>
                <div class="col-md-4">
                    <div class="card text-center">
                        <div class="card-body">
                            <h5 class="card-title text-success">${connectedNodes}</h5>
                            <p class="card-text">在线节点</p>
                        </div>
                    </div>
                </div>
                <div class="col-md-4">
                    <div class="card text-center">
                        <div class="card-body">
                            <h5 class="card-title text-danger">${totalNodes - connectedNodes}</h5>
                            <p class="card-text">离线节点</p>
                        </div>
                    </div>
                </div>
            </div>
            
            <h6>节点详情:</h6>
            <div class="table-responsive">
                <table class="table table-striped">
                    <thead>
                        <tr>
                            <th>节点名称</th>
                            <th>主机地址</th>
                            <th>端口</th>
                            <th>状态</th>
                            <th>进程数</th>
                            <th>最后检查</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${environment.members.map(member => `
                            <tr>
                                <td>
                                    <i class="bi bi-hdd-network"></i> ${member.name}
                                </td>
                                <td>${member.host}</td>
                                <td>${member.port}</td>
                                <td>
                                    <span class="badge bg-${member.is_connected ? 'success' : 'danger'}">
                                        <i class="bi bi-${member.is_connected ? 'check-circle' : 'x-circle'}"></i>
                                        ${member.is_connected ? '在线' : '离线'}
                                    </span>
                                </td>
                                <td>${member.processes || 0}</td>
                                <td>${member.last_ping ? new Date(member.last_ping).toLocaleString() : '未知'}</td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
            </div>
        `;
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
        const alertId = 'alert-' + Date.now();
        
        const alertHTML = `
            <div id="${alertId}" class="alert alert-${type} alert-dismissible fade show" role="alert">
                ${message}
                <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
            </div>
        `;
        
        alertContainer.insertAdjacentHTML('beforeend', alertHTML);
        
        // 3秒后自动移除警告
        setTimeout(() => {
            const alertElement = document.getElementById(alertId);
            if (alertElement) {
                alertElement.remove();
            }
        }, 3000);
    }
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    window.environmentManager = new EnvironmentManager();
});