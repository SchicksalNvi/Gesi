import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Form, Input, Button, Card, message, Space } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { authApi } from '@/api/auth';
import { useStore } from '@/store';

export default function Login() {
  const navigate = useNavigate();
  const { setUser, setToken } = useStore();
  const [loading, setLoading] = useState(false);

  const onFinish = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      const response = await authApi.login(values);
      
      if (response.status === 'success' && response.data) {
        const { token, user } = response.data;
        
        // 设置 token
        setToken(token);
        
        // 设置用户信息
        setUser(user);
        
        message.success('Login successful!');
        navigate('/dashboard');
      } else {
        message.error(response.message || 'Login failed');
      }
    } catch (error: any) {
      message.error(error.response?.data?.message || 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      }}
    >
      <Card
        style={{
          width: 400,
          boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
        }}
      >
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div style={{ textAlign: 'center' }}>
            <h1 style={{ fontSize: 32, fontWeight: 'bold', margin: 0 }}>
              Go-CESI
            </h1>
            <p style={{ color: '#666', marginTop: 8 }}>
              Supervisor Management Platform
            </p>
          </div>

          <Form
            name="login"
            onFinish={onFinish}
            autoComplete="off"
            size="large"
          >
            <Form.Item
              name="username"
              rules={[{ required: true, message: 'Please input your username!' }]}
            >
              <Input
                prefix={<UserOutlined />}
                placeholder="Username"
              />
            </Form.Item>

            <Form.Item
              name="password"
              rules={[{ required: true, message: 'Please input your password!' }]}
            >
              <Input.Password
                prefix={<LockOutlined />}
                placeholder="Password"
              />
            </Form.Item>

            <Form.Item>
              <Button
                type="primary"
                htmlType="submit"
                loading={loading}
                block
              >
                Log in
              </Button>
            </Form.Item>
          </Form>

          <div style={{ textAlign: 'center', color: '#999', fontSize: 12 }}>
            Default: admin / 123456
          </div>
        </Space>
      </Card>
    </div>
  );
}
