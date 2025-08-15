class UserManager {
    constructor() {
        // 确保 Api 类已加载
        if (!window.Api || !window.api) {
            throw new Error('Api class is not loaded. Please check if api.js is properly included.');
        }
        this.api = window.api;
        this.initializeElements();
        this.bindEvents();
        this.loadUsers();
    }

    initializeElements() {
        // 表格和按钮
        this.usersTableBody = document.getElementById('users-table-body');
        this.logoutBtn = document.getElementById('logout-btn');
        
        // 创建用户表单元素
        this.createUserForm = document.getElementById('create-user-form');
        this.newUsernameInput = document.getElementById('new-username');
        this.newPasswordInput = document.getElementById('new-password');
        this.newIsAdminInput = document.getElementById('new-is-admin');
        this.createUserBtn = document.getElementById('create-user-btn');

        // 修改密码表单元素
        this.changePasswordForm = document.getElementById('change-password-form');
        this.changePasswordUsername = document.getElementById('change-password-username');
        this.oldPasswordInput = document.getElementById('old-password');
        this.newPasswordChangeInput = this.changePasswordForm.querySelector('#new-password');
        this.changePasswordBtn = document.getElementById('change-password-btn');

        // Bootstrap 模态框
        this.createUserModal = new bootstrap.Modal(document.getElementById('createUserModal'));
        this.changePasswordModal = new bootstrap.Modal(document.getElementById('changePasswordModal'));
    }

    bindEvents() {
        // 退出登录
        this.logoutBtn.addEventListener('click', () => this.handleLogout());

        // 创建用户
        this.createUserBtn.addEventListener('click', () => this.handleCreateUser());

        // 修改密码
        this.changePasswordBtn.addEventListener('click', () => this.handleChangePassword());
    }

    async loadUsers() {
        try {
            const response = await this.api.getUsers();
            if (response.status === 'success') {
                this.renderUsers(response.data);
            } else {
                this.showAlert(response.message || '加载用户列表失败', 'danger');
            }
        } catch (error) {
            this.showAlert(error.message || '加载用户列表失败', 'danger');
        }
    }

    renderUsers(users) {
        this.usersTableBody.innerHTML = '';
        users.forEach(user => {
            const tr = document.createElement('tr');
            tr.innerHTML = `
                <td>${user.username}</td>
                <td>${user.is_admin ? '管理员' : '普通用户'}</td>
                <td>
                    <button class="btn btn-sm btn-outline-primary me-2" onclick="userManager.openChangePasswordModal('${user.username}')">
                        <i class="bi bi-key"></i> 修改密码
                    </button>
                    <button class="btn btn-sm btn-outline-danger" onclick="userManager.handleDeleteUser('${user.username}')">
                        <i class="bi bi-trash"></i> 删除
                    </button>
                </td>
            `;
            this.usersTableBody.appendChild(tr);
        });
    }

    async handleCreateUser() {
        const username = this.newUsernameInput.value.trim();
        const password = this.newPasswordInput.value;
        const isAdmin = this.newIsAdminInput.checked;

        if (!username || !password) {
            this.showAlert('请填写用户名和密码', 'warning');
            return;
        }

        try {
            this.createUserBtn.disabled = true;
            this.createUserBtn.innerHTML = '<i class="bi bi-hourglass-split"></i> 创建中...';

            const response = await this.api.createUser(username, password, isAdmin);
            if (response.status === 'success') {
                this.showAlert('用户创建成功', 'success');
                this.createUserModal.hide();
                this.createUserForm.reset();
                await this.loadUsers();
            } else {
                this.showAlert(response.message || '创建用户失败', 'danger');
            }
        } catch (error) {
            this.showAlert(error.message || '创建用户失败', 'danger');
        } finally {
            this.createUserBtn.disabled = false;
            this.createUserBtn.innerHTML = '创建';
        }
    }

    openChangePasswordModal(username) {
        this.changePasswordUsername.value = username;
        this.changePasswordModal.show();
    }

    async handleChangePassword() {
        const username = this.changePasswordUsername.value;
        const oldPassword = this.oldPasswordInput.value;
        const newPassword = this.newPasswordChangeInput.value;

        if (!oldPassword || !newPassword) {
            this.showAlert('请填写密码', 'warning');
            return;
        }

        try {
            this.changePasswordBtn.disabled = true;
            this.changePasswordBtn.innerHTML = '<i class="bi bi-hourglass-split"></i> 保存中...';

            const response = await this.api.changePassword(username, oldPassword, newPassword);
            if (response.status === 'success') {
                this.showAlert('密码修改成功', 'success');
                this.changePasswordModal.hide();
                this.changePasswordForm.reset();
            } else {
                this.showAlert(response.message || '修改密码失败', 'danger');
            }
        } catch (error) {
            this.showAlert(error.message || '修改密码失败', 'danger');
        } finally {
            this.changePasswordBtn.disabled = false;
            this.changePasswordBtn.innerHTML = '保存';
        }
    }

    async handleDeleteUser(username) {
        if (!confirm(`确定要删除用户 ${username} 吗？此操作不可撤销。`)) {
            return;
        }

        try {
            const response = await this.api.deleteUser(username);
            if (response.status === 'success') {
                this.showAlert('用户删除成功', 'success');
                await this.loadUsers();
            } else {
                this.showAlert(response.message || '删除用户失败', 'danger');
            }
        } catch (error) {
            this.showAlert(error.message || '删除用户失败', 'danger');
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
        alertContainer.innerHTML = '';
        alertContainer.appendChild(alert);

        // 5秒后自动关闭提示
        setTimeout(() => {
            alert.remove();
        }, 5000);
    }
}

// 初始化用户管理器并暴露到全局作用域（用于事件处理）
window.userManager = new UserManager();