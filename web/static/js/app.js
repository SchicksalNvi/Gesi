class PageManager {
    constructor() {
        if (!window.Api || !window.api) {
            throw new Error('Api class is not loaded. Please check if api.js is properly included.');
        }
        this.api = window.api;
        this.initializeElements();
        this.bindEvents();
        this.loadInitialContent();
        this.setupRefreshInterval();
    }

    initializeElements() {
        this.appContent = document.getElementById('app-content');
        this.refreshButton = document.getElementById('refresh-button');
        this.navLinks = document.querySelectorAll('.nav-link');
    }

    bindEvents() {
        // 导航事件监听
        this.navLinks.forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const href = link.getAttribute('href');
                this.loadContent(href);
            });
        });

        // 刷新按钮事件监听
        if (this.refreshButton) {
            this.refreshButton.addEventListener('click', () => this.refreshContent());
        }
    }

    setupRefreshInterval() {
        // 每60秒自动刷新一次
        setInterval(() => this.refreshContent(), 60000);
    }

    async loadInitialContent() {
        const path = window.location.pathname;
        await this.loadContent(path);
    }

    async loadContent(path) {
        try {
            const data = await window.refreshData();
            switch (path) {
                case '/dashboard':
                    this.loadDashboard(data);
                    break;
                case '/nodes':
                    this.loadNodes(data.nodes);
                    break;
                case '/logs':
                    this.loadLogs(data.logs);
                    break;
                default:
                    this.loadDashboard(data);
            }
        } catch (error) {
            this.showError('加载内容失败: ' + error.message);
        }
    }

    async refreshContent() {
        const path = window.location.pathname;
        await this.loadContent(path);
    }

    loadDashboard(data) {
        const { nodes, logs } = data;
        const totalNodes = nodes.length;
        const totalProcesses = nodes.reduce((sum, node) => sum + node.processes.length, 0);
        const runningProcesses = nodes.reduce((sum, node) => {
            return sum + node.processes.filter(p => p.statename === 'RUNNING').length;
        }, 0);

        const html = `
            <div class="container mt-4">
                <div class="row mb-4">
                    <div class="col">
                        <h2>仪表盘</h2>
                        <button id="refresh-button" class="btn btn-primary float-end">
                            <i class="bi bi-arrow-clockwise"></i> 刷新
                        </button>
                    </div>
                </div>
                <div class="row g-4">
                    <div class="col-md-4">
                        <div class="card h-100">
                            <div class="card-body text-center">
                                <h5 class="card-title">节点数</h5>
                                <p class="card-text display-4">${totalNodes}</p>
                            </div>
                        </div>
                    </div>
                    <div class="col-md-4">
                        <div class="card h-100">
                            <div class="card-body text-center">
                                <h5 class="card-title">总进程数</h5>
                                <p class="card-text display-4">${totalProcesses}</p>
                            </div>
                        </div>
                    </div>
                    <div class="col-md-4">
                        <div class="card h-100">
                            <div class="card-body text-center">
                                <h5 class="card-title">运行中进程</h5>
                                <p class="card-text display-4">${runningProcesses}</p>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="row mt-4">
                    <div class="col-md-6">
                        <div class="card">
                            <div class="card-header">
                                <h5 class="card-title mb-0">最新日志</h5>
                            </div>
                            <div class="card-body">
                                <div class="list-group">
                                    ${logs.slice(0, 5).map(log => `
                                        <div class="list-group-item">
                                            <div class="d-flex w-100 justify-content-between">
                                                <h6 class="mb-1">${log.node_name} - ${log.process_name}</h6>
                                                <small>${new Date(log.timestamp).toLocaleString()}</small>
                                            </div>
                                            <p class="mb-1">${log.message}</p>
                                        </div>
                                    `).join('')}
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="col-md-6">
                        <div class="card">
                            <div class="card-header">
                                <h5 class="card-title mb-0">节点状态</h5>
                            </div>
                            <div class="card-body">
                                <div class="list-group">
                                    ${nodes.map(node => `
                                        <div class="list-group-item">
                                            <div class="d-flex w-100 justify-content-between">
                                                <h6 class="mb-1">${node.name}</h6>
                                                <span class="badge bg-${node.connected ? 'success' : 'danger'}">
                                                    ${node.connected ? '在线' : '离线'}
                                                </span>
                                            </div>
                                            <p class="mb-1">进程数: ${node.processes.length}</p>
                                        </div>
                                    `).join('')}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;

        if (this.appContent) {
            this.appContent.innerHTML = html;
            this.refreshButton = document.getElementById('refresh-button');
            if (this.refreshButton) {
                this.refreshButton.addEventListener('click', () => this.refreshContent());
            }
        }
    }

    loadNodes(nodes) {
        const html = `
            <div class="container mt-4">
                <div class="row mb-4">
                    <div class="col">
                        <h2>节点管理</h2>
                        <button id="refresh-button" class="btn btn-primary float-end">
                            <i class="bi bi-arrow-clockwise"></i> 刷新
                        </button>
                    </div>
                </div>
                ${nodes.map(node => `
                    <div class="card mb-4">
                        <div class="card-header d-flex justify-content-between align-items-center">
                            <h5 class="mb-0">${node.name}</h5>
                            <span class="badge bg-${node.connected ? 'success' : 'danger'}">
                                ${node.connected ? '在线' : '离线'}
                            </span>
                        </div>
                        <div class="card-body">
                            <div class="table-responsive">
                                <table class="table">
                                    <thead>
                                        <tr>
                                            <th>进程名</th>
                                            <th>状态</th>
                                            <th>描述</th>
                                            <th>操作</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        ${node.processes.map(process => `
                                            <tr>
                                                <td>${process.name}</td>
                                                <td>
                                                    <span class="badge bg-${process.statename === 'RUNNING' ? 'success' : 'warning'}">
                                                        ${process.statename}
                                                    </span>
                                                </td>
                                                <td>${process.description || '-'}</td>
                                                <td>
                                                    <div class="btn-group btn-group-sm">
                                                        <button class="btn btn-success" onclick="window.pageManager.startProcess('${node.name}', '${process.name}')">
                                                            <i class="bi bi-play-fill"></i>
                                                        </button>
                                                        <button class="btn btn-warning" onclick="window.pageManager.stopProcess('${node.name}', '${process.name}')">
                                                            <i class="bi bi-stop-fill"></i>
                                                        </button>
                                                        <button class="btn btn-info" onclick="window.pageManager.restartProcess('${node.name}', '${process.name}')">
                                                            <i class="bi bi-arrow-repeat"></i>
                                                        </button>
                                                    </div>
                                                </td>
                                            </tr>
                                        `).join('')}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;

        if (this.appContent) {
            this.appContent.innerHTML = html;
            this.refreshButton = document.getElementById('refresh-button');
            if (this.refreshButton) {
                this.refreshButton.addEventListener('click', () => this.refreshContent());
            }
        }
    }

    loadLogs(logs) {
        const html = `
            <div class="container mt-4">
                <div class="row mb-4">
                    <div class="col">
                        <h2>系统日志</h2>
                        <button id="refresh-button" class="btn btn-primary float-end">
                            <i class="bi bi-arrow-clockwise"></i> 刷新
                        </button>
                    </div>
                </div>
                <div class="card">
                    <div class="card-body">
                        <div class="table-responsive">
                            <table class="table">
                                <thead>
                                    <tr>
                                        <th>时间</th>
                                        <th>节点</th>
                                        <th>进程</th>
                                        <th>消息</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    ${logs.map(log => `
                                        <tr>
                                            <td>${new Date(log.timestamp).toLocaleString()}</td>
                                            <td>${log.node_name}</td>
                                            <td>${log.process_name}</td>
                                            <td>${log.message}</td>
                                        </tr>
                                    `).join('')}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>
        `;

        if (this.appContent) {
            this.appContent.innerHTML = html;
            this.refreshButton = document.getElementById('refresh-button');
            if (this.refreshButton) {
                this.refreshButton.addEventListener('click', () => this.refreshContent());
            }
        }
    }

    async startProcess(nodeName, processName) {
        try {
            await this.api.startProcess(nodeName, processName);
            await this.refreshContent();
        } catch (error) {
            this.showError('启动进程失败: ' + error.message);
        }
    }

    async stopProcess(nodeName, processName) {
        try {
            await this.api.stopProcess(nodeName, processName);
            await this.refreshContent();
        } catch (error) {
            this.showError('停止进程失败: ' + error.message);
        }
    }

    async restartProcess(nodeName, processName) {
        try {
            await this.api.restartProcess(nodeName, processName);
            await this.refreshContent();
        } catch (error) {
            this.showError('重启进程失败: ' + error.message);
        }
    }

    showError(message) {
        // 创建错误提示元素
        const alertDiv = document.createElement('div');
        alertDiv.className = 'alert alert-danger alert-dismissible fade show position-fixed top-0 start-50 translate-middle-x mt-3';
        alertDiv.setAttribute('role', 'alert');
        alertDiv.innerHTML = `
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
        `;

        // 添加到页面
        document.body.appendChild(alertDiv);

        // 3秒后自动移除
        setTimeout(() => {
            alertDiv.remove();
        }, 3000);
    }
}

// 创建全局页面管理器实例
window.pageManager = new PageManager();
