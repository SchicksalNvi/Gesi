class SupervisorManager {
    constructor() {
        this.nodes = [];
        this.currentNode = null;
        this.init();
    }

    async init() {
        await this.loadNodes();
        this.setupEventListeners();
    }

    async loadNodes() {
        try {
            const response = await fetch('/api/supervisor/nodes');
            this.nodes = await response.json();
            this.renderNodes();
        } catch (err) {
            console.error('Failed to load nodes:', err);
        }
    }

    renderNodes() {
        const nodesList = document.getElementById('nodes-list');
        nodesList.innerHTML = this.nodes.map(node => `
            <div class="node-item" data-node-id="${node.id}">
                <h3>${node.name}</h3>
                <p>Status: <span class="status">${node.status || 'unknown'}</span></p>
                <button class="btn-view-processes">View Processes</button>
            </div>
        `).join('');
    }

    async loadProcesses(nodeId) {
        try {
            const response = await fetch(`/api/supervisor/nodes/${nodeId}/processes`);
            const processes = await response.json();
            this.renderProcesses(processes);
            this.currentNode = nodeId;
        } catch (err) {
            console.error('Failed to load processes:', err);
        }
    }

    renderProcesses(processes) {
        const processesList = document.getElementById('processes-list');
        processesList.innerHTML = processes.map(process => `
            <div class="process-item">
                <h3>${process.name}</h3>
                <p>Status: <span class="status ${process.status}">${process.status}</span></p>
                <div class="process-actions">
                    <button class="btn-start" data-process="${process.name}">Start</button>
                    <button class="btn-stop" data-process="${process.name}">Stop</button>
                    <button class="btn-restart" data-process="${process.name}">Restart</button>
                </div>
            </div>
        `).join('');
    }

    setupEventListeners() {
        document.getElementById('nodes-list').addEventListener('click', async (e) => {
            if (e.target.classList.contains('btn-view-processes')) {
                const nodeItem = e.target.closest('.node-item');
                const nodeId = nodeItem.dataset.nodeId;
                await this.loadProcesses(nodeId);
            }
        });

        document.getElementById('processes-list').addEventListener('click', async (e) => {
            const processName = e.target.dataset.process;
            if (!processName || !this.currentNode) return;

            let endpoint = '';
            if (e.target.classList.contains('btn-start')) {
                endpoint = 'start';
            } else if (e.target.classList.contains('btn-stop')) {
                endpoint = 'stop';
            } else if (e.target.classList.contains('btn-restart')) {
                endpoint = 'restart';
            }

            if (endpoint) {
                try {
                    const response = await fetch(
                        `/api/supervisor/nodes/${this.currentNode}/processes/${processName}/${endpoint}`,
                        { method: 'POST' }
                    );
                    if (response.ok) {
                        await this.loadProcesses(this.currentNode);
                    }
                } catch (err) {
                    console.error(`Failed to ${endpoint} process:`, err);
                }
            }
        });
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new SupervisorManager();
});