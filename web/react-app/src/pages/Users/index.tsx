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
  const { user: currentUser } = useStore();
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
      
      message.success('Password changed successfully');
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
      
      message.success('Notification settings updated');
    } catch (error: any) {
      message.error(error.response?.data?.message || 'Failed to update notifications');
    } finally {
      setSaving(false);
    }
  };

  const columns: ColumnsType<User> = [
    {
      title: 'User',
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
      title: 'Full Name',
      dataIndex: 'full_name',
      key: 'full_name',
      render: (name) => name || '-',
    },
    {
      title: 'Role',
      dataIndex: 'is_admin',
      key: 'role',
      render: (isAdmin: boolean) => (
        <Tag color={isAdmin ? 'red' : 'blue'}>
          {isAdmin ? 'Admin' : 'User'}
        </Tag>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'is_active',
      key: 'status',
      render: (isActive: boolean) => (
        <Tag color={isActive ? 'green' : 'default'}>
          {isActive ? 'Active' : 'Inactive'}
        </Tag>
      ),
    },
    {
      title: 'Last Login',
      dataIndex: 'last_login',
      key: 'last_login',
      render: (date) => date ? new Date(date).toLocaleString() : 'Never',
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            icon={<SettingOutlined />}
            onClick={() => handleUserSettings(record)}
          >
            Settings
          </Button>
          {isAdmin && (
            <Popconfirm
              title="Delete user"
              description="Are you sure you want to delete this user?"
              onConfirm={() => handleDelete(record.id)}
              okText="Yes"
              cancelText="No"
              disabled={record.is_admin}
            >
              <Button type="link" danger icon={<DeleteOutlined />} disabled={record.is_admin}>
                Delete
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
        title="User Management"
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={loadUsers} loading={loading}>
              Refresh
            </Button>
            {isAdmin && (
              <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
                Add User
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
            showTotal: (t) => `Total ${t} users`,
            onChange: (p, ps) => { setPage(p); setPageSize(ps); },
          }}
        />
      </Card>

      {/* Create User Modal */}
      <Modal
        title="Create User"
        open={createModalVisible}
        onOk={handleCreateSubmit}
        onCancel={() => setCreateModalVisible(false)}
        width={500}
      >
        <Form form={createForm} layout="vertical" initialValues={{ is_admin: false }}>
          <Form.Item
            name="username"
            label="Username"
            rules={[
              { required: true, message: 'Please enter username' },
              { min: 3, max: 50, message: 'Username must be 3-50 characters' },
            ]}
          >
            <Input prefix={<UserOutlined />} placeholder="Username" />
          </Form.Item>
          <Form.Item
            name="email"
            label="Email"
            rules={[
              { required: true, message: 'Please enter email' },
              { type: 'email', message: 'Please enter valid email' },
            ]}
          >
            <Input prefix={<MailOutlined />} placeholder="Email" />
          </Form.Item>
          <Form.Item name="full_name" label="Full Name">
            <Input placeholder="Full Name (optional)" />
          </Form.Item>
          <Form.Item
            name="password"
            label="Password"
            rules={[
              { required: true, message: 'Please enter password' },
              { min: 6, message: 'Password must be at least 6 characters' },
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="Password" />
          </Form.Item>
          <Form.Item name="is_admin" label="Admin Role" valuePropName="checked">
            <Switch checkedChildren="Admin" unCheckedChildren="User" />
          </Form.Item>
        </Form>
      </Modal>

      {/* User Settings Drawer */}
      <Drawer
        title={`User Settings: ${selectedUser?.username || ''}`}
        placement="right"
        width={520}
        onClose={() => setDrawerVisible(false)}
        open={drawerVisible}
      >
        <Tabs
          items={[
            {
              key: 'profile',
              label: <span><UserOutlined /> Profile</span>,
              children: (
                <Form form={profileForm} layout="vertical">
                  <Form.Item name="username" label="Username">
                    <Input prefix={<UserOutlined />} disabled />
                  </Form.Item>
                  <Form.Item
                    name="email"
                    label="Email"
                    rules={[
                      { required: true, message: 'Please enter email' },
                      { type: 'email', message: 'Please enter valid email' },
                    ]}
                  >
                    <Input prefix={<MailOutlined />} />
                  </Form.Item>
                  <Form.Item name="full_name" label="Full Name">
                    <Input placeholder="Enter full name" />
                  </Form.Item>
                  <Form.Item name="timezone" label="Timezone">
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
                      <Form.Item name="is_admin" label="Admin Role" valuePropName="checked">
                        <Switch checkedChildren="Admin" unCheckedChildren="User" />
                      </Form.Item>
                      <Form.Item name="is_active" label="Account Status" valuePropName="checked">
                        <Switch checkedChildren="Active" unCheckedChildren="Inactive" />
                      </Form.Item>
                    </>
                  )}
                  <Form.Item>
                    <Button type="primary" icon={<SaveOutlined />} onClick={handleProfileUpdate} loading={saving}>
                      Save Profile
                    </Button>
                  </Form.Item>
                </Form>
              ),
            },
            {
              key: 'security',
              label: <span><LockOutlined /> Security</span>,
              children: (
                <Form form={passwordForm} layout="vertical">
                  <Form.Item
                    name="new_password"
                    label="New Password"
                    rules={[
                      { required: true, message: 'Please enter new password' },
                      { min: 6, message: 'Password must be at least 6 characters' },
                    ]}
                  >
                    <Input.Password prefix={<LockOutlined />} placeholder="Enter new password" />
                  </Form.Item>
                  <Form.Item
                    name="confirm_password"
                    label="Confirm Password"
                    dependencies={['new_password']}
                    rules={[
                      { required: true, message: 'Please confirm password' },
                      ({ getFieldValue }) => ({
                        validator(_, value) {
                          if (!value || getFieldValue('new_password') === value) {
                            return Promise.resolve();
                          }
                          return Promise.reject(new Error('Passwords do not match'));
                        },
                      }),
                    ]}
                  >
                    <Input.Password prefix={<LockOutlined />} placeholder="Confirm new password" />
                  </Form.Item>
                  <Form.Item>
                    <Button type="primary" icon={<SaveOutlined />} onClick={handlePasswordChange} loading={saving}>
                      Reset Password
                    </Button>
                  </Form.Item>
                </Form>
              ),
            },
            {
              key: 'notifications',
              label: <span><BellOutlined /> Notifications</span>,
              children: (
                <Form form={notificationForm} layout="vertical">
                  <Form.Item name="email_notifications" label="Email Notifications" valuePropName="checked">
                    <Switch />
                  </Form.Item>
                  <Divider />
                  <Form.Item name="process_alerts" label="Process Alerts" valuePropName="checked">
                    <Switch />
                  </Form.Item>
                  <Form.Item name="system_alerts" label="System Alerts" valuePropName="checked">
                    <Switch />
                  </Form.Item>
                  <Form.Item name="node_status_changes" label="Node Status Changes" valuePropName="checked">
                    <Switch />
                  </Form.Item>
                  <Divider />
                  <Form.Item name="weekly_report" label="Weekly Summary Report" valuePropName="checked">
                    <Switch />
                  </Form.Item>
                  <Form.Item>
                    <Button type="primary" icon={<SaveOutlined />} onClick={handleNotificationUpdate} loading={saving}>
                      Save Preferences
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
