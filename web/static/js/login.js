class LoginManager {
    constructor() {
        // 确保 Api 类已加载
        if (!window.Api || !window.api) {
            throw new Error('Api class is not loaded. Please check if api.js is properly included.');
        }

        this.api = window.api;
        this.initializeElements();
        this.bindEvents();
        this.checkAuthStatus();
    }

    initializeElements() {
        this.loginForm = document.getElementById('login-form');
        this.usernameInput = document.getElementById('username');
        this.passwordInput = document.getElementById('password');
        this.submitButton = document.querySelector('button[type="submit"]');

        if (!this.loginForm || !this.usernameInput || !this.passwordInput || !this.submitButton) {
            throw new Error('Required login form elements not found');
        }
    }

    bindEvents() {
        this.loginForm.addEventListener('submit', (event) => {
            event.preventDefault();
            this.handleLogin();
        });
    }

    async checkAuthStatus() {
        try {
            const response = await this.api.checkAuth();
            if (response.status === 'success' && response.data) {
                // 如果已经登录，重定向到仪表板
                window.location.href = '/dashboard';
            }
        } catch (error) {
            // 未登录状态，保持在登录页面
            console.log('Not logged in');
        }
    }

    async handleLogin() {
        const username = this.usernameInput.value.trim();
        const password = this.passwordInput.value;

        if (!username || !password) {
            this.showAlert('请输入用户名和密码', 'warning');
            return;
        }

        try {
            // 禁用提交按钮
            this.submitButton.disabled = true;
            this.submitButton.innerHTML = '<i class="bi bi-hourglass-split"></i> 登录中...';

            const response = await this.api.login(username, password);
            if (response.status === 'success') {
                window.location.href = '/dashboard';
            } else {
                this.showAlert(response.message || '登录失败', 'danger');
            }
        } catch (error) {
            this.showAlert(error.message || '登录失败，请检查用户名和密码', 'danger');
        } finally {
            // 恢复提交按钮
            this.submitButton.disabled = false;
            this.submitButton.innerHTML = '<i class="bi bi-box-arrow-in-right"></i> 登录';
        }
    }

    showAlert(message, type = 'info') {
        const alertContainer = document.getElementById('alert-container');
        if (!alertContainer) {
            console.error('Alert container not found');
            return;
        }

        const alert = document.createElement('div');
        alert.className = `alert alert-${type} alert-dismissible fade show`;
        alert.innerHTML = `
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        `;
        alertContainer.innerHTML = '';
        alertContainer.appendChild(alert);

        // 5秒后自动关闭提示
        setTimeout(() => {
            alert.remove();
        }, 5000);
    }
}

// 等待 DOM 加载完成
document.addEventListener('DOMContentLoaded', () => {
    if (!window.Api || !window.api) {
        console.error('Api class is not loaded. Please check if api.js is properly included.');
        return;
    }
    window.loginManager = new LoginManager();
});