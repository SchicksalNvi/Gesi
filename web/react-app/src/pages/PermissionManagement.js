import React, { useState, useEffect } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Badge } from '../components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../components/ui/tabs';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '../components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../components/ui/select';
import { Checkbox } from '../components/ui/checkbox';
import { toast } from 'sonner';
import {
  Users,
  Shield,
  Plus,
  Edit,
  Trash2,
  Search,
  UserCheck,
  Settings,
  Eye,
  Lock,
  Unlock
} from 'lucide-react';

const PermissionManagement = () => {
  const [roles, setRoles] = useState([]);
  const [permissions, setPermissions] = useState([]);
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedRole, setSelectedRole] = useState(null);
  const [showCreateRoleDialog, setShowCreateRoleDialog] = useState(false);
  const [showAssignPermissionsDialog, setShowAssignPermissionsDialog] = useState(false);
  const [showAssignRoleDialog, setShowAssignRoleDialog] = useState(false);
  const [newRole, setNewRole] = useState({ name: '', displayName: '', description: '' });
  const [selectedPermissions, setSelectedPermissions] = useState([]);
  const [selectedUser, setSelectedUser] = useState('');

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [rolesRes, permissionsRes, usersRes] = await Promise.all([
        fetch('/api/roles', {
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        }),
        fetch('/api/permissions', {
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        }),
        fetch('/api/users', {
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
      ]);

      if (rolesRes.ok && permissionsRes.ok && usersRes.ok) {
        const rolesData = await rolesRes.json();
        const permissionsData = await permissionsRes.json();
        const usersData = await usersRes.json();
        
        setRoles(rolesData.data || rolesData);
        setPermissions(permissionsData);
        setUsers(usersData.data || usersData);
      }
    } catch (error) {
      console.error('Error fetching data:', error);
      toast.error('获取数据失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateRole = async () => {
    try {
      const response = await fetch('/api/roles', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify(newRole)
      });

      if (response.ok) {
        toast.success('角色创建成功');
        setShowCreateRoleDialog(false);
        setNewRole({ name: '', displayName: '', description: '' });
        fetchData();
      } else {
        const error = await response.json();
        toast.error(error.error || '创建角色失败');
      }
    } catch (error) {
      console.error('Error creating role:', error);
      toast.error('创建角色失败');
    }
  };

  const handleDeleteRole = async (roleId) => {
    if (!window.confirm('确定要删除这个角色吗？')) return;

    try {
      const response = await fetch(`/api/roles/${roleId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (response.ok) {
        toast.success('角色删除成功');
        fetchData();
      } else {
        const error = await response.json();
        toast.error(error.error || '删除角色失败');
      }
    } catch (error) {
      console.error('Error deleting role:', error);
      toast.error('删除角色失败');
    }
  };

  const handleAssignPermissions = async () => {
    try {
      const response = await fetch(`/api/roles/${selectedRole.id}/permissions`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify(selectedPermissions)
      });

      if (response.ok) {
        toast.success('权限分配成功');
        setShowAssignPermissionsDialog(false);
        setSelectedPermissions([]);
        fetchData();
      } else {
        const error = await response.json();
        toast.error(error.error || '权限分配失败');
      }
    } catch (error) {
      console.error('Error assigning permissions:', error);
      toast.error('权限分配失败');
    }
  };

  const handleAssignRole = async () => {
    try {
      const response = await fetch(`/api/roles/${selectedRole.id}/users/${selectedUser}`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (response.ok) {
        toast.success('角色分配成功');
        setShowAssignRoleDialog(false);
        setSelectedUser('');
        fetchData();
      } else {
        const error = await response.json();
        toast.error(error.error || '角色分配失败');
      }
    } catch (error) {
      console.error('Error assigning role:', error);
      toast.error('角色分配失败');
    }
  };

  const filteredRoles = roles.filter(role =>
    role.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    role.display_name?.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const getResourceIcon = (resource) => {
    switch (resource) {
      case 'system': return <Settings className="w-4 h-4" />;
      case 'user': return <Users className="w-4 h-4" />;
      case 'node': return <Shield className="w-4 h-4" />;
      case 'process': return <Settings className="w-4 h-4" />;
      case 'log': return <Eye className="w-4 h-4" />;
      case 'config': return <Settings className="w-4 h-4" />;
      default: return <Lock className="w-4 h-4" />;
    }
  };

  const getActionColor = (action) => {
    switch (action) {
      case 'read': return 'bg-blue-100 text-blue-800';
      case 'write': return 'bg-green-100 text-green-800';
      case 'delete': return 'bg-red-100 text-red-800';
      case 'execute': return 'bg-purple-100 text-purple-800';
      case 'manage': return 'bg-orange-100 text-orange-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">权限管理</h1>
          <p className="text-gray-600">管理系统角色和权限分配</p>
        </div>
        <Dialog open={showCreateRoleDialog} onOpenChange={setShowCreateRoleDialog}>
          <DialogTrigger asChild>
            <Button className="bg-blue-600 hover:bg-blue-700">
              <Plus className="w-4 h-4 mr-2" />
              创建角色
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>创建新角色</DialogTitle>
            </DialogHeader>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">角色名称</label>
                <Input
                  value={newRole.name}
                  onChange={(e) => setNewRole({ ...newRole, name: e.target.value })}
                  placeholder="输入角色名称"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">显示名称</label>
                <Input
                  value={newRole.displayName}
                  onChange={(e) => setNewRole({ ...newRole, displayName: e.target.value })}
                  placeholder="输入显示名称"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">描述</label>
                <Input
                  value={newRole.description}
                  onChange={(e) => setNewRole({ ...newRole, description: e.target.value })}
                  placeholder="输入角色描述"
                />
              </div>
              <div className="flex justify-end space-x-2">
                <Button variant="outline" onClick={() => setShowCreateRoleDialog(false)}>
                  取消
                </Button>
                <Button onClick={handleCreateRole}>
                  创建
                </Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>
      </div>

      <Tabs defaultValue="roles" className="space-y-4">
        <TabsList>
          <TabsTrigger value="roles">角色管理</TabsTrigger>
          <TabsTrigger value="permissions">权限列表</TabsTrigger>
          <TabsTrigger value="audit">审计日志</TabsTrigger>
        </TabsList>

        <TabsContent value="roles" className="space-y-4">
          <Card>
            <CardHeader>
              <div className="flex justify-between items-center">
                <CardTitle>角色列表</CardTitle>
                <div className="flex items-center space-x-2">
                  <div className="relative">
                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                    <Input
                      placeholder="搜索角色..."
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      className="pl-10 w-64"
                    />
                  </div>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {filteredRoles.map((role) => (
                  <Card key={role.id} className="border border-gray-200">
                    <CardContent className="p-4">
                      <div className="flex justify-between items-start mb-3">
                        <div>
                          <h3 className="font-semibold text-lg">{role.display_name || role.name}</h3>
                          <p className="text-sm text-gray-600">{role.name}</p>
                        </div>
                        {role.is_system && (
                          <Badge variant="secondary" className="text-xs">
                            系统角色
                          </Badge>
                        )}
                      </div>
                      
                      <p className="text-sm text-gray-700 mb-4">{role.description}</p>
                      
                      <div className="flex flex-wrap gap-1 mb-4">
                        {role.permissions?.slice(0, 3).map((permission) => (
                          <Badge key={permission.id} variant="outline" className="text-xs">
                            {permission.display_name}
                          </Badge>
                        ))}
                        {role.permissions?.length > 3 && (
                          <Badge variant="outline" className="text-xs">
                            +{role.permissions.length - 3} 更多
                          </Badge>
                        )}
                      </div>
                      
                      <div className="flex justify-between items-center">
                        <span className="text-sm text-gray-500">
                          {role.users?.length || 0} 个用户
                        </span>
                        <div className="flex space-x-1">
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => {
                              setSelectedRole(role);
                              setSelectedPermissions(role.permissions?.map(p => p.id) || []);
                              setShowAssignPermissionsDialog(true);
                            }}
                          >
                            <Shield className="w-3 h-3" />
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => {
                              setSelectedRole(role);
                              setShowAssignRoleDialog(true);
                            }}
                          >
                            <UserCheck className="w-3 h-3" />
                          </Button>
                          {!role.is_system && (
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => handleDeleteRole(role.id)}
                              className="text-red-600 hover:text-red-700"
                            >
                              <Trash2 className="w-3 h-3" />
                            </Button>
                          )}
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="permissions" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>权限列表</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {Object.entries(
                  permissions.reduce((acc, permission) => {
                    const resource = permission.resource || 'other';
                    if (!acc[resource]) acc[resource] = [];
                    acc[resource].push(permission);
                    return acc;
                  }, {})
                ).map(([resource, resourcePermissions]) => (
                  <div key={resource} className="border rounded-lg p-4">
                    <div className="flex items-center mb-3">
                      {getResourceIcon(resource)}
                      <h3 className="ml-2 font-semibold capitalize">{resource}</h3>
                    </div>
                    <div className="grid gap-2 md:grid-cols-2 lg:grid-cols-3">
                      {resourcePermissions.map((permission) => (
                        <div key={permission.id} className="flex items-center justify-between p-2 border rounded">
                          <div>
                            <span className="font-medium text-sm">{permission.display_name}</span>
                            <p className="text-xs text-gray-600">{permission.description}</p>
                          </div>
                          <Badge className={getActionColor(permission.action)}>
                            {permission.action}
                          </Badge>
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="audit" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>审计日志</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-center py-8 text-gray-500">
                审计日志功能开发中...
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* 权限分配对话框 */}
      <Dialog open={showAssignPermissionsDialog} onOpenChange={setShowAssignPermissionsDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>为角色分配权限</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 max-h-96 overflow-y-auto">
            {Object.entries(
              permissions.reduce((acc, permission) => {
                const resource = permission.resource || 'other';
                if (!acc[resource]) acc[resource] = [];
                acc[resource].push(permission);
                return acc;
              }, {})
            ).map(([resource, resourcePermissions]) => (
              <div key={resource} className="border rounded-lg p-4">
                <div className="flex items-center mb-3">
                  {getResourceIcon(resource)}
                  <h3 className="ml-2 font-semibold capitalize">{resource}</h3>
                </div>
                <div className="space-y-2">
                  {resourcePermissions.map((permission) => (
                    <div key={permission.id} className="flex items-center space-x-2">
                      <Checkbox
                        checked={selectedPermissions.includes(permission.id)}
                        onCheckedChange={(checked) => {
                          if (checked) {
                            setSelectedPermissions([...selectedPermissions, permission.id]);
                          } else {
                            setSelectedPermissions(selectedPermissions.filter(id => id !== permission.id));
                          }
                        }}
                      />
                      <div className="flex-1">
                        <span className="font-medium text-sm">{permission.display_name}</span>
                        <p className="text-xs text-gray-600">{permission.description}</p>
                      </div>
                      <Badge className={getActionColor(permission.action)}>
                        {permission.action}
                      </Badge>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
          <div className="flex justify-end space-x-2">
            <Button variant="outline" onClick={() => setShowAssignPermissionsDialog(false)}>
              取消
            </Button>
            <Button onClick={handleAssignPermissions}>
              确认分配
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* 用户角色分配对话框 */}
      <Dialog open={showAssignRoleDialog} onOpenChange={setShowAssignRoleDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>为用户分配角色</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">选择用户</label>
              <Select value={selectedUser} onValueChange={setSelectedUser}>
                <SelectTrigger>
                  <SelectValue placeholder="选择用户" />
                </SelectTrigger>
                <SelectContent>
                  {users.map((user) => (
                    <SelectItem key={user.id} value={user.id}>
                      {user.username} ({user.email})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="flex justify-end space-x-2">
              <Button variant="outline" onClick={() => setShowAssignRoleDialog(false)}>
                取消
              </Button>
              <Button onClick={handleAssignRole} disabled={!selectedUser}>
                分配角色
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default PermissionManagement;