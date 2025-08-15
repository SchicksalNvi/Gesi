class DashboardManager {
    constructor() {
        // 确保 Api 类已加载
        if (!window.Api || !window.api) {
            throw new Error('Api class is not loaded. Please check if api.js is properly included.');
        }
        this.api = window.api;
        this.initializeElements();
        this.bindEvents();
        this.loadDashboardData();
    }

    initializeElements() {
        // 统计卡片元素
        this.totalNodesElement = document.getElementById('total-nodes');
        this.runningProcessesElement = document.getElementById('running-processes');
        this.stoppedProcessesElement = document.getElementById('stopped-processes');
        this.totalProcessesElement = document.getElementById('total-processes');

        // 进程表格元素
        this.processesTableBody = document.getElementById('processes-table');
        this.refreshButton = document.getElementById('refresh-btn');
        this.logoutBtn = document.getElementById('logout-btn');

        // 日志模态框
        this.logsModal = new bootstrap.Modal(document.getElementById('logs-modal'));
        this.logsModalBody = document.querySelector('#logs-modal .modal-body');
    }

    bindEvents() {
        this.refreshButton.addEventListener('click', () => this.loadDashboardData());
        this.logoutBtn.addEventListener('click', () => this.handleLogout());

        // 为进程表格添加事件委托
        this.processesTableBody.addEventListener('click', (event) => {
            const target = event.target;
            if (target.classList.contains('action-btn')) {
                const action = target.dataset.action;
                const nodeId = target.dataset.node;
                const processName = target.dataset.process;

                switch (action) {
                    case 'start':
                        this.startProcess(nodeId, processName);
                        break;
                    case 'stop':
                        this.stopProcess(nodeId, processName);
                        break;
                    case 'restart':
                        this.restartProcess(nodeId, processName);
                        break;
                    case 'logs':
                        this.showProcessLogs(nodeId, processName);
                        break;
                }
            }
        });
    }

    async loadDashboardData() {
        try {
            const nodes = await this.api.getNodes();
            let totalProcesses = 0;
            let runningProcesses = 0;
            let stoppedProcesses = 0;
            const processesData = [];

            // 获取每个节点的进程信息
            for (const node of nodes) {
                const processes = await this.api.getNodeProcesses(node.name);
                processes.forEach(process => {
                    totalProcesses++;
                    if (process.statename === 'RUNNING') {
                        runningProcesses++;
                    } else if (process.statename === 'STOPPED') {
                        stoppedProcesses++;
                    }
                    processesData.push({ node: node.name, ...process });
                });
            }

            // 更新统计数据
            this.updateStatistics(nodes.length, runningProcesses, stoppedProcesses, totalProcesses);

            // 更新进程表格
            this.updateProcessesTable(processesData);

        } catch (error) {
            this.showAlert('Error loading dashboard data: ' + error.message, 'danger');
        }
    }

    updateStatistics(totalNodes, running, stopped, total) {
        this.totalNodesElement.textContent = totalNodes;
        this.runningProcessesElement.textContent = running;
        this.stoppedProcessesElement.textContent = stopped;
        this.totalProcessesElement.textContent = total;
    }

    updateProcessesTable(processes) {
        this.processesTableBody.innerHTML = '';
        processes.forEach(process => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>${process.node}</td>
                <td>${process.name}</td>
                <td>${process.group}</td>
                <td>
                    <span class="process-status status-${process.statename.toLowerCase()}"></span>
                    ${process.statename}
                </td>
                <td>${this.formatUptime(process.description)}</td>
                <td>
                    <div class="btn-group btn-group-sm">
                        ${process.statename === 'STOPPED' ? `
                            <button class="btn btn-success action-btn" data-action="start" data-node="${process.node}" data-process="${process.name}">
                                <i class="bi bi-play"></i>
                            </button>
                        ` : `
                            <button class="btn btn-danger action-btn" data-action="stop" data-node="${process.node}" data-process="${process.name}">
                                <i class="bi bi-stop"></i>
                            </button>
                        `}
                        <button class="btn btn-warning action-btn" data-action="restart" data-node="${process.node}" data-process="${process.name}">
                            <i class="bi bi-arrow-clockwise"></i>
                        </button>
                        <button class="btn btn-info action-btn" data-action="logs" data-node="${process.node}" data-process="${process.name}">
                            <i class="bi bi-journal-text"></i>
                        </button>
                    </div>
                </td>
            `;
            this.processesTableBody.appendChild(row);
        });
    }

    formatUptime(description) {
        // 处理进程描述中的运行时间信息
        const uptimeMatch = description.match(/uptime ([\d:]+)/);
        return uptimeMatch ? uptimeMatch[1] : 'N/A';
    }

    async startProcess(nodeId, processName) {
        try {
            await this.api.startProcess(nodeId, processName);
            this.showAlert(`Successfully started process ${processName} on node ${nodeId}`, 'success');
            await this.loadDashboardData();
        } catch (error) {
            this.showAlert(`Error starting process: ${error.message}`, 'danger');
        }
    }

    async stopProcess(nodeId, processName) {
        try {
            await this.api.stopProcess(nodeId, processName);
            this.showAlert(`Successfully stopped process ${processName} on node ${nodeId}`, 'success');
            await this.loadDashboardData();
        } catch (error) {
            this.showAlert(`Error stopping process: ${error.message}`, 'danger');
        }
    }

    async restartProcess(nodeId, processName) {
        try {
            await this.api.stopProcess(nodeId, processName);
            await this.api.startProcess(nodeId, processName);
            this.showAlert(`Successfully restarted process ${processName} on node ${nodeId}`, 'success');
            await this.loadDashboardData();
        } catch (error) {
            this.showAlert(`Error restarting process: ${error.message}`, 'danger');
        }
    }

    async showProcessLogs(nodeId, processName) {
        try {
            const response = await this.api.getProcessLogs(nodeId, processName);
            const logs = response.logs;
            let logText = '';
            if (logs.stdout && logs.stdout.length > 0) {
                logText += 'STDOUT:\n' + logs.stdout.join('\n') + '\n\n';
            }
            if (logs.stderr && logs.stderr.length > 0) {
                logText += 'STDERR:\n' + logs.stderr.join('\n');
            }
            if (!logText) {
                logText = 'No logs available';
            }
            this.logsModalBody.innerHTML = `<pre class="mb-0"><code>${logText}</code></pre>`;
            this.logsModal.show();
        } catch (error) {
            this.showAlert(`Error fetching logs: ${error.message}`, 'danger');
        }
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

// 初始化仪表板管理器
document.addEventListener('DOMContentLoaded', () => {
    new DashboardManager();
});