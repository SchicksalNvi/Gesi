class NodesManager {
    constructor() {
        // 确保 Api 类已加载
        if (!window.Api || !window.api) {
            throw new Error('Api class is not loaded. Please check if api.js is properly included.');
        }
        this.api = window.api;
        this.nodes = [];
        this.selectedNodes = new Set();
        this.logoutBtn = document.getElementById('logout-btn');
        this.init();
    }

    async init() {
        await this.loadNodes();
        this.setupEventListeners();
    }

    async loadNodes() {
        try {
            const result = await this.api.getNodes();
            this.nodes = result.nodes || [];
            this.renderNodes();
            this.renderNodesList();
        } catch (err) {
            console.error('Failed to load nodes:', err);
            this.showError('Failed to load nodes');
        }
    }

    renderNodes() {
        const nodesContainer = document.getElementById('nodes-container');
        nodesContainer.innerHTML = '';

        this.nodes.forEach(node => {
            if (this.selectedNodes.has(node.name)) {
                this.renderNodeProcesses(node);
            }
        });
    }

    renderNodesList() {
        const nodesList = document.getElementById('nodes-list');
        nodesList.innerHTML = this.nodes.map(node => `
            <div class="form-check">
                <input class="form-check-input" type="checkbox" 
                       id="node-${node.name}" name="${node.name}" 
                       ${this.selectedNodes.has(node.name) ? 'checked' : ''}>
                <label class="form-check-label" for="node-${node.name}">
                    ${node.name} (${node.environment})
                </label>
            </div>
        `).join('');
    }

    async renderNodeProcesses(node) {
        try {
            const result = await this.api.getNodeProcesses(node.name);
            const processes = result.processes || [];

            const processesHtml = `
                <div class="card mb-3" id="node-${node.name}-processes">
                    <div class="card-header">
                        <h5 class="card-title mb-0">
                            ${node.name} Processes
                            <span class="badge bg-secondary">${processes.length}</span>
                        </h5>
                    </div>
                    <div class="card-body">
                        <div class="btn-group mb-3">
                            <button class="btn btn-success btn-sm" onclick="nodesManager.startAllProcesses('${node.name}')">
                                Start All
                            </button>
                            <button class="btn btn-danger btn-sm" onclick="nodesManager.stopAllProcesses('${node.name}')">
                                Stop All
                            </button>
                            <button class="btn btn-warning btn-sm" onclick="nodesManager.restartAllProcesses('${node.name}')">
                                Restart All
                            </button>
                        </div>
                        <div class="table-responsive">
                            <table class="table table-hover">
                                <thead>
                                    <tr>
                                        <th>Name</th>
                                        <th>Group</th>
                                        <th>PID</th>
                                        <th>Uptime</th>
                                        <th>State</th>
                                        <th>Actions</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    ${processes.map(process => this.renderProcess(node.name, process)).join('')}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            `;

            const nodesContainer = document.getElementById('nodes-container');
            const existingNode = document.getElementById(`node-${node.name}-processes`);
            if (existingNode) {
                existingNode.outerHTML = processesHtml;
            } else {
                nodesContainer.insertAdjacentHTML('beforeend', processesHtml);
            }
        } catch (err) {
            console.error(`Failed to load processes for node ${node.name}:`, err);
            this.showError(`Failed to load processes for node ${node.name}`);
        }
    }

    renderProcess(nodeName, process) {
        const stateClass = this.getStateClass(process.state);
        return `
            <tr class="${stateClass}">
                <td>${process.name}</td>
                <td>${process.group}</td>
                <td>${process.pid || '-'}</td>
                <td>${process.uptime || '-'}</td>
                <td>${process.state}</td>
                <td>
                    <div class="btn-group btn-group-sm">
                        <button class="btn btn-success" onclick="nodesManager.startProcess('${nodeName}', '${process.name}')">
                            Start
                        </button>
                        <button class="btn btn-danger" onclick="nodesManager.stopProcess('${nodeName}', '${process.name}')">
                            Stop
                        </button>
                        <button class="btn btn-warning" onclick="nodesManager.restartProcess('${nodeName}', '${process.name}')">
                            Restart
                        </button>
                        <button class="btn btn-info" onclick="nodesManager.showLogs('${nodeName}', '${process.name}')">
                            Logs
                        </button>
                    </div>
                </td>
            </tr>
        `;
    }

    getStateClass(state) {
        switch (state?.toUpperCase()) {
            case 'RUNNING': return 'table-success';
            case 'STOPPED': return 'table-warning';
            case 'FATAL': return 'table-danger';
            case 'STARTING': return 'table-info';
            default: return '';
        }
    }

    setupEventListeners() {
        document.getElementById('nodes-list').addEventListener('change', (e) => {
            if (e.target.type === 'checkbox') {
                const nodeName = e.target.name;
                if (e.target.checked) {
                    this.selectedNodes.add(nodeName);
                } else {
                    this.selectedNodes.delete(nodeName);
                }
                this.renderNodes();
            }
        });

        document.getElementById('refresh-btn').addEventListener('click', () => {
            this.loadNodes();
        });

        this.logoutBtn.addEventListener('click', () => this.handleLogout());
    }

    async startProcess(nodeName, processName) {
        try {
            await this.api.startProcess(nodeName, processName);
            await this.renderNodeProcesses(this.nodes.find(n => n.name === nodeName));
        } catch (err) {
            this.showError(`Failed to start process ${processName}`);
        }
    }

    async stopProcess(nodeName, processName) {
        try {
            await this.api.stopProcess(nodeName, processName);
            await this.renderNodeProcesses(this.nodes.find(n => n.name === nodeName));
        } catch (err) {
            this.showError(`Failed to stop process ${processName}`);
        }
    }

    async restartProcess(nodeName, processName) {
        try {
            await this.api.restartProcess(nodeName, processName);
            await this.renderNodeProcesses(this.nodes.find(n => n.name === nodeName));
        } catch (err) {
            this.showError(`Failed to restart process ${processName}`);
        }
    }

    async showLogs(nodeName, processName) {
        try {
            const result = await this.api.getProcessLogs(nodeName, processName);
            const logs = result.logs || { stdout: [], stderr: [] };
            
            const modal = document.getElementById('logs-modal');
            const modalTitle = modal.querySelector('.modal-title');
            const modalBody = modal.querySelector('.modal-body');
            
            modalTitle.textContent = `Logs: ${nodeName} - ${processName}`;
            modalBody.innerHTML = `
                <div class="logs-container">
                    <h6>Stdout:</h6>
                    <pre class="logs-stdout">${logs.stdout.join('\n')}</pre>
                    <h6>Stderr:</h6>
                    <pre class="logs-stderr">${logs.stderr.join('\n')}</pre>
                </div>
            `;
            
            new bootstrap.Modal(modal).show();
        } catch (err) {
            this.showError(`Failed to load logs for process ${processName}`);
        }
    }

    async handleLogout() {
        try {
            await this.api.logout();
            window.location.href = '/login';
        } catch (error) {
            this.showError('退出登录失败');
        }
    }

    showError(message) {
        const alertContainer = document.getElementById('alert-container');
        const alert = `
            <div class="alert alert-danger alert-dismissible fade show" role="alert">
                ${message}
                <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
            </div>
        `;
        alertContainer.insertAdjacentHTML('beforeend', alert);
    }
}

const nodesManager = new NodesManager();
window.nodesManager = nodesManager; // 使其在全局可用