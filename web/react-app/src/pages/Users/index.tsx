import { useEffect, useState } from 'react';
import {
  Table,
  Tag,
  Button,
  Space,
  Card,
  message,
  Modal,
  Form,
  Input,
  Switch,
  Popconfirm,
  Avatar,
  Drawer,
  Tabs,
  Divider,
  Select,
} from 'antd';
import {
  PlusOutlined,
  DeleteOutlined,
  ReloadOutlined,
  UserOutlined,
  LockOutlined,
  MailOutlined,
  SettingOutlined,
  BellOutlined,
  SaveOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { usersApi, User, CreateUserRequest } from '../../api/users';
import { useStore } from '../../store';

const Users: React.FC = () => {
  const { user: currentUser, t } = useStore();
  const isAdmin = currentUser?.is_admin ?? false;
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [saving, setSaving] = useState(false);
  const [createForm] = Form.useForm();
  const [profileForm] = Form.useForm();
  const [passwordForm] = Form.useForm();
  const [notificationForm] = Form.useForm();

  useEffect(() => {
    loadUsers();
  }, [page, pageSize]);

  const loadUsers = async () => {
    setLoading(true);
    try {
      if (isAdmin) {
        // Admin sees all users
        const response = await usersApi.getUsers(page, pageSize);
        if (response?.data) {
          setUsers(response.data.users || []);
          setTotal(response.data.total || 0);
        }
      } else {
        // Non-admin only sees themselves
        if (currentUser) {
          setUsers([{
            id: currentUser.id,
            username: currentUser.username,
            email: currentUser.email || '',
            full_name: currentUser.full_name,
            is_admin: currentUser.is_admin,
            is_active: true,
            created_at: '',
            updated_at: '',
          }]);
          setTotal(1);
        }
      }
    } catch (error) {
      console.error('Failed to load users:', error);
      message.error('Failed to load users');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    createForm.resetFields();
    setCreateModalVisible(true);
  };

  const handleUserSettings = async (user: User) => {
    setSelectedUser(user);
    setDrawerVisible(true);
    
    // Set basic user info first
    profileForm.setFieldsValue({
      username: user.username,
      email: user.email,
      full_name: user.full_name || '',
      is_admin: user.is_admin,
      is_active: user.is_active,
      timezone: 'UTC', // Default, will be overwritten
    });
    passwordForm.resetFields();
    
    // Load user preferences from backend
    try {
      const prefs = await usersApi.getUserPreferences(user.id);
      // Update profile form with timezone
      profileForm.setFieldsValue({
        timezone: prefs.timezone || 'UTC',
      });
      // Update notification form
      notificationForm.setFieldsValue({
        email_notifications: prefs.email_notifications ?? true,
        process_alerts: prefs.process_alerts ?? true,
        system_alerts: prefs.system_alerts ?? true,
        node_status_changes: prefs.node_status_changes ?? false,
        weekly_report: prefs.weekly_report ?? false,
      });
    } catch (error) {
      console.error('Failed to load user preferences:', error);
      // Set defaults if loading fails
      notificationForm.setFieldsValue({
        email_notifications: true,
        process_alerts: true,
        system_alerts: true,
        node_status_changes: false,
        weekly_report: false,
      });
    }
  };

  const handleDelete = async (userId: string) => {
    try {
      await usersApi.deleteUser(userId);
      message.success('User deleted successfully');
      loadUsers();
    } catch (error: any) {
      message.error(error.response?.data?.message || 'Failed to delete user');
    }
  };

  const handleCreateSubmit = async () => {
    try {
      const values = await createForm.validateFields();
      const createData: CreateUserRequest = {
        username: values.username,
        email: values.email,
        password: values.password,
        full_name: values.full_name,
        is_admin: values.is_admin || false,
      };
      await usersApi.createUser(createData);
      message.success('User created successfully');
      setCreateModalVisible(false);
      loadUsers();
    } catch (error: any) {
      message.error(error.response?.data?.message || 'Failed to create user');
    }
  };

  const handleProfileUpdate = async () => {
    if (!selectedUser) return;
    try {
      const values = await profileForm.validateFields();
      setSaving(true);
      
      // Update user basic info
      await usersApi.updateUser(selectedUser.id, {
        email: values.email,
        full_name: values.full_name,
        is_admin: values.is_admin,
        is_active: values.is_active,
      });
      
      // Update timezone in user preferences
      await usersApi.updateUserPreferences(selectedUser.id, {
        timezone: values.timezone,
      });
      
      message.success('User updated successfully');
      loadUsers();
      // Update selectedUser to reflect changes
      setSelectedUser({ ...selectedUser, ...values });
    } catch (error: any) {
      message.error(error.response?.data?.message || 'Failed to update user');
    } finally {
      setSaving(false);
    }
  };

  const handlePasswordChange = async () => {
    if (!selectedUser) return;
    try {
      const values = await passwordForm.validateFields();
      setSaving(true);
      
      await usersApi.resetPassword(selectedUser.id, values.new_password);
      
      message.success(t.users.passwordChanged);
      passwordForm.resetFields();
    } catch (error: any) {
      message.error(error.response?.data?.message || 'Failed to change password');
    } finally {
      setSaving(false);
    }
  };

  const handleNotificationUpdate = async () => {
    if (!selectedUser) return;
    try {
      const values = await notificationForm.validateFields();
      setSaving(true);
      
      await usersApi.updateUserPreferences(selectedUser.id, {
        email_notifications: values.email_notifications,
        process_alerts: values.process_alerts,
        system_alerts: values.system_alerts,
        node_status_changes: values.node_status_changes,
        weekly_report: values.weekly_report,
      });
      
      message.success(t.users.notificationUpdated);
    } catch (error: any) {
      message.error(error.response?.data?.message || 'Failed to update notifications');
    } finally {
      setSaving(false);
    }
  };

  const columns: ColumnsType<User> = [
    {
      title: t.users.username,
      key: 'user',
      render: (_, record) => (
        <Space>
          <Avatar icon={<UserOutlined />} />
          <div>
            <div style={{ fontWeight: 500 }}>{record.username}</div>
            <div style={{ fontSize: '12px', color: '#999' }}>{record.email}</div>
          </div>
        </Space>
      ),
    },
    {
      title: t.common.name,
      dataIndex: 'full_name',
      key: 'full_name',
      render: (name) => name || '-',
    },
    {
      title: t.users.role,
      dataIndex: 'is_admin',
      key: 'role',
      render: (isAdmin: boolean) => (
        <Tag color={isAdmin ? 'red' : 'blue'}>
          {isAdmin ? t.users.admin : t.users.user}
        </Tag>
      ),
    },
    {
      title: t.common.status,
      dataIndex: 'is_active',
      key: 'status',
      render: (isActive: boolean) => (
        <Tag color={isActive ? 'green' : 'default'}>
          {isActive ? t.common.enabled : t.common.disabled}
        </Tag>
      ),
    },
    {
      title: t.users.lastLogin,
      dataIndex: 'last_login',
      key: 'last_login',
      render: (date) => date ? new Date(date).toLocaleString() : t.users.never,
    },
    {
      title: t.common.actions,
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            icon={<SettingOutlined />}
            onClick={() => handleUserSettings(record)}
          >
            {t.nav.settings}
          </Button>
          {isAdmin && (
            <Popconfirm
              title={t.users.deleteUser}
              description={t.users.confirmDelete}
              onConfirm={() => handleDelete(record.id)}
              okText={t.common.yes}
              cancelText={t.common.no}
              disabled={record.is_admin}
            >
              <Button type="link" danger icon={<DeleteOutlined />} disabled={record.is_admin}>
                {t.common.delete}
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Card
        title={t.users.title}
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={loadUsers} loading={loading}>
              {t.common.refresh}
            </Button>
            {isAdmin && (
              <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
                {t.users.addUser}
              </Button>
            )}
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={users}
          rowKey="id"
          loading={loading}
          pagination={{
            current: page,
            pageSize: pageSize,
            total: total,
            showSizeChanger: true,
            showTotal: (total) => `${t.common.total} ${total}`,
            onChange: (p, ps) => { setPage(p); setPageSize(ps); },
          }}
        />
      </Card>

      {/* Create User Modal */}
      <Modal
        title={t.users.addUser}
        open={createModalVisible}
        onOk={handleCreateSubmit}
        onCancel={() => setCreateModalVisible(false)}
        width={500}
      >
        <Form form={createForm} layout="vertical" initialValues={{ is_admin: false }}>
          <Form.Item
            name="username"
            label={t.users.username}
            rules={[
              { required: true, message: t.login.usernameRequired },
              { min: 3, max: 50, message: t.users.usernameLength },
            ]}
          >
            <Input prefix={<UserOutlined />} placeholder={t.users.username} />
          </Form.Item>
          <Form.Item
            name="email"
            label={t.users.email}
            rules={[
              { required: true, message: t.users.pleaseEnterEmail },
              { type: 'email', message: t.users.pleaseEnterValidEmail },
            ]}
          >
            <Input prefix={<MailOutlined />} placeholder={t.users.email} />
          </Form.Item>
          <Form.Item name="full_name" label={t.users.fullName}>
            <Input placeholder={t.users.fullName} />
          </Form.Item>
          <Form.Item
            name="password"
            label={t.users.password}
            rules={[
              { required: true, message: t.login.passwordRequired },
              { min: 6, message: t.users.passwordMinLength },
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder={t.users.password} />
          </Form.Item>
          <Form.Item name="is_admin" label={t.users.role} valuePropName="checked">
            <Switch checkedChildren={t.users.admin} unCheckedChildren={t.users.user} />
          </Form.Item>
        </Form>
      </Modal>

      {/* User Settings Drawer */}
      <Drawer
        title={`${t.users.userSettings}: ${selectedUser?.username || ''}`}
        placement="right"
        width={520}
        onClose={() => setDrawerVisible(false)}
        open={drawerVisible}
      >
        <Tabs
          items={[
            {
              key: 'profile',
              label: <span><UserOutlined /> {t.users.profile}</span>,
              children: (
                <Form form={profileForm} layout="vertical">
                  <Form.Item name="username" label={t.users.username}>
                    <Input prefix={<UserOutlined />} disabled />
                  </Form.Item>
                  <Form.Item
                    name="email"
                    label={t.users.email}
                    rules={[
                      { required: true, message: t.users.pleaseEnterEmail },
                      { type: 'email', message: t.users.pleaseEnterValidEmail },
                    ]}
                  >
                    <Input prefix={<MailOutlined />} />
                  </Form.Item>
                  <Form.Item name="full_name" label={t.users.fullName}>
                    <Input placeholder={t.users.enterFullName} />
                  </Form.Item>
                  <Form.Item name="timezone" label={t.settings.timezone}>
                    <Select
                      options={[
                        { label: 'UTC', value: 'UTC' },
                        { label: 'America/New_York', value: 'America/New_York' },
                        { label: 'America/Los_Angeles', value: 'America/Los_Angeles' },
                        { label: 'Europe/London', value: 'Europe/London' },
                        { label: 'Asia/Shanghai', value: 'Asia/Shanghai' },
                        { label: 'Asia/Tokyo', value: 'Asia/Tokyo' },
                      ]}
                    />
                  </Form.Item>
                  {isAdmin && (
                    <>
                      <Divider />
                      <Form.Item name="is_admin" label={t.users.adminRole} valuePropName="checked">
                        <Switch checkedChildren={t.users.admin} unCheckedChildren={t.users.user} />
                      </Form.Item>
                      <Form.Item name="is_active" label={t.users.accountStatus} valuePropName="checked">
                        <Switch checkedChildren={t.users.active} unCheckedChildren={t.users.inactive} />
                      </Form.Item>
                    </>
                  )}
                  <Form.Item>
                    <Button type="primary" icon={<SaveOutlined />} onClick={handleProfileUpdate} loading={saving}>
                      {t.users.saveProfile}
                    </Button>
                  </Form.Item>
                </Form>
              ),
            },
            {
              key: 'security',
              label: <span><LockOutlined /> {t.users.security}</span>,
              children: (
                <Form form={passwordForm} layout="vertical">
                  <Form.Item
                    name="new_password"
                    label={t.users.newPassword}
                    rules={[
                      { required: true, message: t.users.pleaseEnterNewPassword },
                      { min: 6, message: t.users.passwordMinLength },
                    ]}
                  >
                    <Input.Password prefix={<LockOutlined />} placeholder={t.users.enterNewPassword} />
                  </Form.Item>
                  <Form.Item
                    name="confirm_password"
                    label={t.users.confirmPassword}
                    dependencies={['new_password']}
                    rules={[
                      { required: true, message: t.users.pleaseConfirmPassword },
                      ({ getFieldValue }) => ({
                        validator(_, value) {
                          if (!value || getFieldValue('new_password') === value) {
                            return Promise.resolve();
                          }
                          return Promise.reject(new Error(t.users.passwordMismatch));
                        },
                      }),
                    ]}
                  >
                    <Input.Password prefix={<LockOutlined />} placeholder={t.users.confirmNewPassword} />
                  </Form.Item>
                  <Form.Item>
                    <Button type="primary" icon={<SaveOutlined />} onClick={handlePasswordChange} loading={saving}>
                      {t.users.resetPassword}
                    </Button>
                  </Form.Item>
                </Form>
              ),
            },
            {
              key: 'notifications',
              label: <span><BellOutlined /> {t.users.notifications}</span>,
              children: (
                <Form form={notificationForm} layout="vertical">
                  <Form.Item name="email_notifications" label={t.users.emailNotifications} valuePropName="checked">
                    <Switch />
                  </Form.Item>
                  <Divider />
                  <Form.Item name="process_alerts" label={t.users.processAlerts} valuePropName="checked">
                    <Switch />
                  </Form.Item>
                  <Form.Item name="system_alerts" label={t.users.systemAlerts} valuePropName="checked">
                    <Switch />
                  </Form.Item>
                  <Form.Item name="node_status_changes" label={t.users.nodeStatusChanges} valuePropName="checked">
                    <Switch />
                  </Form.Item>
                  <Divider />
                  <Form.Item name="weekly_report" label={t.users.weeklyReport} valuePropName="checked">
                    <Switch />
                  </Form.Item>
                  <Form.Item>
                    <Button type="primary" icon={<SaveOutlined />} onClick={handleNotificationUpdate} loading={saving}>
                      {t.users.savePreferences}
                    </Button>
                  </Form.Item>
                </Form>
              ),
            },
          ]}
        />
      </Drawer>
    </div>
  );
};

export default Users;
