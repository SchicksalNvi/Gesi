import React, { useState, useEffect } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Textarea } from '../components/ui/textarea';
import { Badge } from '../components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../components/ui/tabs';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '../components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../components/ui/select';
import { Switch } from '../components/ui/switch';
import { Checkbox } from '../components/ui/checkbox';
import { toast } from 'sonner';
import {
  Settings,
  Plus,
  Edit,
  Trash2,
  Save,
  RefreshCw,
  Download,
  Upload,
  Copy,
  Eye,
  EyeOff,
  History,
  GitBranch,
  FileText,
  Lock,
  Unlock,
  Search,
  Filter,
  AlertTriangle,
  CheckCircle,
  Clock,
  User,
  Calendar,
  Database,
  Server,
  Globe,
  Zap
} from 'lucide-react';

const ConfigurationManagement = () => {
  const [configurations, setConfigurations] = useState([]);
  const [environments, setEnvironments] = useState([]);
  const [configHistory, setConfigHistory] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedEnv, setSelectedEnv] = useState('production');
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedCategory, setSelectedCategory] = useState('all');
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [showHistoryDialog, setShowHistoryDialog] = useState(false);
  const [showImportDialog, setShowImportDialog] = useState(false);
  const [selectedConfig, setSelectedConfig] = useState(null);
  const [editingConfig, setEditingConfig] = useState(null);
  const [showSecrets, setShowSecrets] = useState({});
  const [newConfig, setNewConfig] = useState({
    key: '',
    value: '',
    description: '',
    category: 'application',
    type: 'string',
    required: false,
    secret: false,
    environment: 'production'
  });
  const [importData, setImportData] = useState('');
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);

  useEffect(() => {
    fetchData();
  }, [selectedEnv]);

  const fetchData = async () => {
    try {
      setLoading(true);
      
      // 获取配置数据
      const configResponse = await fetch('/api/configurations', {
        credentials: 'include'
      });
      if (configResponse.ok) {
        const configs = await configResponse.json();
        setConfigurations(configs);
      }
      
      // 获取环境数据
      const envResponse = await fetch('/api/environments', {
        credentials: 'include'
      });
      if (envResponse.ok) {
        const envs = await envResponse.json();
        setEnvironments(envs);
      }
      
      // 获取配置历史
      const historyResponse = await fetch('/api/configurations/history', {
        credentials: 'include'
      });
      if (historyResponse.ok) {
        const history = await historyResponse.json();
        setConfigHistory(history);
      }
    } catch (error) {
      console.error('Error fetching configuration data:', error);
      toast.error('获取配置数据失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateConfig = async () => {
    try {
      const response = await fetch('/api/configurations', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        credentials: 'include',
        body: JSON.stringify(newConfig)
      });
      
      if (response.ok) {
        const createdConfig = await response.json();
        setConfigurations([...configurations, createdConfig]);
        setShowCreateDialog(false);
        setNewConfig({
          key: '',
          value: '',
          description: '',
          category: 'application',
          type: 'string',
          required: false,
          secret: false,
          environment: 'production'
        });
        toast.success('配置项创建成功');
      } else {
        throw new Error('创建配置失败');
      }
    } catch (error) {
      console.error('Error creating configuration:', error);
      toast.error('创建配置项失败');
    }
  };

  const handleUpdateConfig = async (configId, updates) => {
    try {
      const response = await fetch(`/api/configurations/${configId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        },
        credentials: 'include',
        body: JSON.stringify(updates)
      });
      
      if (response.ok) {
        const updatedConfig = await response.json();
        setConfigurations(configurations.map(config => 
          config.id === configId ? updatedConfig : config
        ));
        setEditingConfig(null);
        setHasUnsavedChanges(false);
        toast.success('配置更新成功');
      } else {
        throw new Error('更新配置失败');
      }
    } catch (error) {
      console.error('Error updating configuration:', error);
      toast.error('更新配置失败');
    }
  };

  const handleDeleteConfig = async (configId) => {
    try {
      const response = await fetch(`/api/configurations/${configId}`, {
        method: 'DELETE',
        credentials: 'include'
      });
      
      if (response.ok) {
        setConfigurations(configurations.filter(config => config.id !== configId));
        toast.success('配置项删除成功');
      } else {
        throw new Error('删除配置失败');
      }
    } catch (error) {
      console.error('Error deleting configuration:', error);
      toast.error('删除配置项失败');
    }
  };

  const handleExportConfig = () => {
    try {
      const exportData = configurations
        .filter(config => config.environment === selectedEnv)
        .reduce((acc, config) => {
          acc[config.key] = config.secret ? '[HIDDEN]' : config.value;
          return acc;
        }, {});
      
      const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `config-${selectedEnv}-${new Date().toISOString().split('T')[0]}.json`;
      a.click();
      URL.revokeObjectURL(url);
      
      toast.success('配置导出成功');
    } catch (error) {
      console.error('Error exporting configuration:', error);
      toast.error('配置导出失败');
    }
  };

  const handleImportConfig = async () => {
    try {
      const response = await fetch('/api/configurations/import', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        credentials: 'include',
        body: JSON.stringify({
          data: importData,
          environment: selectedEnv
        })
      });
      
      if (response.ok) {
        const result = await response.json();
        await fetchData(); // 重新获取数据
        setShowImportDialog(false);
        setImportData('');
        toast.success(`成功导入 ${result.count} 个配置项`);
      } else {
        throw new Error('导入配置失败');
      }
    } catch (error) {
      console.error('Error importing configuration:', error);
      toast.error('配置导入失败，请检查JSON格式');
    }
  };

  const toggleSecretVisibility = (configId) => {
    setShowSecrets(prev => ({
      ...prev,
      [configId]: !prev[configId]
    }));
  };

  const getCategoryIcon = (category) => {
    switch (category) {
      case 'database': return <Database className="w-4 h-4" />;
      case 'cache': return <Zap className="w-4 h-4" />;
      case 'security': return <Lock className="w-4 h-4" />;
      case 'logging': return <FileText className="w-4 h-4" />;
      case 'monitoring': return <Eye className="w-4 h-4" />;
      case 'performance': return <Server className="w-4 h-4" />;
      case 'application': return <Globe className="w-4 h-4" />;
      default: return <Settings className="w-4 h-4" />;
    }
  };

  const getCategoryColor = (category) => {
    switch (category) {
      case 'database': return 'bg-blue-100 text-blue-800';
      case 'cache': return 'bg-yellow-100 text-yellow-800';
      case 'security': return 'bg-red-100 text-red-800';
      case 'logging': return 'bg-green-100 text-green-800';
      case 'monitoring': return 'bg-purple-100 text-purple-800';
      case 'performance': return 'bg-orange-100 text-orange-800';
      case 'application': return 'bg-gray-100 text-gray-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getTypeColor = (type) => {
    switch (type) {
      case 'string': return 'bg-blue-50 text-blue-700';
      case 'number': return 'bg-green-50 text-green-700';
      case 'boolean': return 'bg-purple-50 text-purple-700';
      default: return 'bg-gray-50 text-gray-700';
    }
  };

  const filteredConfigurations = configurations
    .filter(config => config.environment === selectedEnv)
    .filter(config => 
      config.key.toLowerCase().includes(searchTerm.toLowerCase()) ||
      config.description.toLowerCase().includes(searchTerm.toLowerCase())
    )
    .filter(config => selectedCategory === 'all' || config.category === selectedCategory);

  const categories = [...new Set(configurations.map(config => config.category))];

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
          <h1 className="text-2xl font-bold text-gray-900">配置管理</h1>
          <p className="text-gray-600">系统配置和环境变量管理</p>
        </div>
        <div className="flex items-center space-x-2">
          <Button onClick={() => setShowHistoryDialog(true)} variant="outline">
            <History className="w-4 h-4 mr-2" />
            变更历史
          </Button>
          <Button onClick={() => setShowImportDialog(true)} variant="outline">
            <Upload className="w-4 h-4 mr-2" />
            导入
          </Button>
          <Button onClick={handleExportConfig} variant="outline">
            <Download className="w-4 h-4 mr-2" />
            导出
          </Button>
          <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
            <DialogTrigger asChild>
              <Button className="bg-blue-600 hover:bg-blue-700">
                <Plus className="w-4 h-4 mr-2" />
                新建配置
              </Button>
            </DialogTrigger>
          </Dialog>
        </div>
      </div>

      <Tabs defaultValue="configurations" className="space-y-4">
        <TabsList>
          <TabsTrigger value="configurations">配置管理</TabsTrigger>
          <TabsTrigger value="environments">环境管理</TabsTrigger>
          <TabsTrigger value="templates">配置模板</TabsTrigger>
        </TabsList>

        <TabsContent value="configurations" className="space-y-4">
          {/* 过滤和搜索 */}
          <Card>
            <CardContent className="p-4">
              <div className="flex flex-wrap items-center gap-4">
                <div className="flex-1 min-w-64">
                  <div className="relative">
                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                    <Input
                      placeholder="搜索配置项..."
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      className="pl-10"
                    />
                  </div>
                </div>
                
                <Select value={selectedEnv} onValueChange={setSelectedEnv}>
                  <SelectTrigger className="w-40">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {environments.map((env) => (
                      <SelectItem key={env.id} value={env.id}>
                        {env.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                
                <Select value={selectedCategory} onValueChange={setSelectedCategory}>
                  <SelectTrigger className="w-32">
                    <SelectValue placeholder="分类" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">所有分类</SelectItem>
                    {categories.map((category) => (
                      <SelectItem key={category} value={category}>
                        {category}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                
                <Button onClick={fetchData} variant="outline">
                  <RefreshCw className="w-4 h-4 mr-2" />
                  刷新
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* 配置列表 */}
          <Card>
            <CardHeader>
              <div className="flex justify-between items-center">
                <CardTitle>配置项 ({filteredConfigurations.length})</CardTitle>
                {hasUnsavedChanges && (
                  <Badge variant="destructive">
                    <AlertTriangle className="w-3 h-3 mr-1" />
                    有未保存的更改
                  </Badge>
                )}
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {filteredConfigurations.map((config) => (
                  <Card key={config.id} className="border border-gray-200">
                    <CardContent className="p-4">
                      {editingConfig === config.id ? (
                        // 编辑模式
                        <div className="space-y-4">
                          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div>
                              <label className="block text-sm font-medium mb-1">配置键</label>
                              <Input
                                value={config.key}
                                onChange={(e) => {
                                  const updatedConfigs = configurations.map(c => 
                                    c.id === config.id ? { ...c, key: e.target.value } : c
                                  );
                                  setConfigurations(updatedConfigs);
                                  setHasUnsavedChanges(true);
                                }}
                              />
                            </div>
                            <div>
                              <label className="block text-sm font-medium mb-1">分类</label>
                              <Select 
                                value={config.category} 
                                onValueChange={(value) => {
                                  const updatedConfigs = configurations.map(c => 
                                    c.id === config.id ? { ...c, category: value } : c
                                  );
                                  setConfigurations(updatedConfigs);
                                  setHasUnsavedChanges(true);
                                }}
                              >
                                <SelectTrigger>
                                  <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                  <SelectItem value="application">应用</SelectItem>
                                  <SelectItem value="database">数据库</SelectItem>
                                  <SelectItem value="cache">缓存</SelectItem>
                                  <SelectItem value="security">安全</SelectItem>
                                  <SelectItem value="logging">日志</SelectItem>
                                  <SelectItem value="monitoring">监控</SelectItem>
                                  <SelectItem value="performance">性能</SelectItem>
                                </SelectContent>
                              </Select>
                            </div>
                          </div>
                          
                          <div>
                            <label className="block text-sm font-medium mb-1">配置值</label>
                            <Textarea
                              value={config.value}
                              onChange={(e) => {
                                const updatedConfigs = configurations.map(c => 
                                  c.id === config.id ? { ...c, value: e.target.value } : c
                                );
                                setConfigurations(updatedConfigs);
                                setHasUnsavedChanges(true);
                              }}
                              rows={3}
                            />
                          </div>
                          
                          <div>
                            <label className="block text-sm font-medium mb-1">描述</label>
                            <Input
                              value={config.description}
                              onChange={(e) => {
                                const updatedConfigs = configurations.map(c => 
                                  c.id === config.id ? { ...c, description: e.target.value } : c
                                );
                                setConfigurations(updatedConfigs);
                                setHasUnsavedChanges(true);
                              }}
                            />
                          </div>
                          
                          <div className="flex items-center space-x-6">
                            <div className="flex items-center space-x-2">
                              <Checkbox 
                                checked={config.required}
                                onCheckedChange={(checked) => {
                                  const updatedConfigs = configurations.map(c => 
                                    c.id === config.id ? { ...c, required: checked } : c
                                  );
                                  setConfigurations(updatedConfigs);
                                  setHasUnsavedChanges(true);
                                }}
                              />
                              <span className="text-sm">必需</span>
                            </div>
                            <div className="flex items-center space-x-2">
                              <Checkbox 
                                checked={config.secret}
                                onCheckedChange={(checked) => {
                                  const updatedConfigs = configurations.map(c => 
                                    c.id === config.id ? { ...c, secret: checked } : c
                                  );
                                  setConfigurations(updatedConfigs);
                                  setHasUnsavedChanges(true);
                                }}
                              />
                              <span className="text-sm">敏感信息</span>
                            </div>
                          </div>
                          
                          <div className="flex justify-end space-x-2">
                            <Button
                              variant="outline"
                              onClick={() => {
                                setEditingConfig(null);
                                setHasUnsavedChanges(false);
                                fetchData(); // 重新获取数据以撤销更改
                              }}
                            >
                              取消
                            </Button>
                            <Button
                              onClick={() => handleUpdateConfig(config.id, config)}
                              className="bg-green-600 hover:bg-green-700"
                            >
                              <Save className="w-3 h-3 mr-1" />
                              保存
                            </Button>
                          </div>
                        </div>
                      ) : (
                        // 查看模式
                        <div>
                          <div className="flex justify-between items-start mb-3">
                            <div className="flex-1">
                              <div className="flex items-center space-x-3 mb-2">
                                <h3 className="font-semibold text-lg">{config.key}</h3>
                                <Badge className={getCategoryColor(config.category)}>
                                  <div className="flex items-center space-x-1">
                                    {getCategoryIcon(config.category)}
                                    <span>{config.category}</span>
                                  </div>
                                </Badge>
                                <Badge className={getTypeColor(config.type)}>
                                  {config.type}
                                </Badge>
                                {config.required && (
                                  <Badge variant="destructive">必需</Badge>
                                )}
                                {config.secret && (
                                  <Badge className="bg-red-100 text-red-800">
                                    <Lock className="w-3 h-3 mr-1" />
                                    敏感
                                  </Badge>
                                )}
                              </div>
                              <p className="text-sm text-gray-600 mb-2">{config.description}</p>
                              <div className="flex items-center space-x-4 text-xs text-gray-500">
                                <span className="flex items-center">
                                  <User className="w-3 h-3 mr-1" />
                                  {config.modifiedBy}
                                </span>
                                <span className="flex items-center">
                                  <Clock className="w-3 h-3 mr-1" />
                                  {new Date(config.lastModified).toLocaleString()}
                                </span>
                                <span className="flex items-center">
                                  <GitBranch className="w-3 h-3 mr-1" />
                                  v{config.version}
                                </span>
                              </div>
                            </div>
                            <div className="flex items-center space-x-2">
                              <Button
                                size="sm"
                                variant="outline"
                                onClick={() => setEditingConfig(config.id)}
                              >
                                <Edit className="w-3 h-3" />
                              </Button>
                              <Button
                                size="sm"
                                variant="outline"
                                onClick={() => {
                                  navigator.clipboard.writeText(config.value);
                                  toast.success('配置值已复制到剪贴板');
                                }}
                              >
                                <Copy className="w-3 h-3" />
                              </Button>
                              <Button
                                size="sm"
                                variant="outline"
                                className="text-red-600"
                                onClick={() => handleDeleteConfig(config.id)}
                              >
                                <Trash2 className="w-3 h-3" />
                              </Button>
                            </div>
                          </div>
                          
                          <div className="bg-gray-50 p-3 rounded border">
                            <div className="flex items-center justify-between">
                              <span className="text-sm font-medium">配置值:</span>
                              {config.secret && (
                                <Button
                                  size="sm"
                                  variant="ghost"
                                  onClick={() => toggleSecretVisibility(config.id)}
                                >
                                  {showSecrets[config.id] ? (
                                    <EyeOff className="w-3 h-3" />
                                  ) : (
                                    <Eye className="w-3 h-3" />
                                  )}
                                </Button>
                              )}
                            </div>
                            <div className="mt-1 font-mono text-sm break-all">
                              {config.secret && !showSecrets[config.id] 
                                ? '••••••••••••••••' 
                                : config.value
                              }
                            </div>
                          </div>
                        </div>
                      )}
                    </CardContent>
                  </Card>
                ))}
                
                {filteredConfigurations.length === 0 && (
                  <div className="text-center text-gray-500 py-8">
                    <Settings className="w-8 h-8 mx-auto mb-2 opacity-50" />
                    <p>暂无配置项</p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="environments" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>环境管理</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {environments.map((env) => (
                  <Card key={env.id} className="border border-gray-200">
                    <CardContent className="p-4">
                      <div className="flex justify-between items-start mb-3">
                        <div>
                          <h3 className="font-semibold text-lg">{env.name}</h3>
                          <p className="text-sm text-gray-600">{env.description}</p>
                        </div>
                        <Badge className={env.active ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}>
                          {env.active ? '活跃' : '停用'}
                        </Badge>
                      </div>
                      
                      <div className="space-y-2">
                        <div className="flex justify-between text-sm">
                          <span>配置项数量:</span>
                          <span className="font-medium">
                            {configurations.filter(config => config.environment === env.id).length}
                          </span>
                        </div>
                        <div className="flex justify-between text-sm">
                          <span>敏感配置:</span>
                          <span className="font-medium">
                            {configurations.filter(config => config.environment === env.id && config.secret).length}
                          </span>
                        </div>
                      </div>
                      
                      <div className="flex justify-end space-x-2 mt-4">
                        <Button size="sm" variant="outline">
                          <Edit className="w-3 h-3" />
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => setSelectedEnv(env.id)}
                        >
                          <Eye className="w-3 h-3" />
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="templates" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>配置模板</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-center text-gray-500 py-8">
                <FileText className="w-8 h-8 mx-auto mb-2 opacity-50" />
                <p>配置模板功能开发中...</p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* 创建配置对话框 */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>创建配置项</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">配置键</label>
                <Input
                  value={newConfig.key}
                  onChange={(e) => setNewConfig({ ...newConfig, key: e.target.value })}
                  placeholder="例如: DATABASE_URL"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">分类</label>
                <Select value={newConfig.category} onValueChange={(value) => setNewConfig({ ...newConfig, category: value })}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="application">应用</SelectItem>
                    <SelectItem value="database">数据库</SelectItem>
                    <SelectItem value="cache">缓存</SelectItem>
                    <SelectItem value="security">安全</SelectItem>
                    <SelectItem value="logging">日志</SelectItem>
                    <SelectItem value="monitoring">监控</SelectItem>
                    <SelectItem value="performance">性能</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            
            <div>
              <label className="block text-sm font-medium mb-1">配置值</label>
              <Textarea
                value={newConfig.value}
                onChange={(e) => setNewConfig({ ...newConfig, value: e.target.value })}
                placeholder="输入配置值"
                rows={3}
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium mb-1">描述</label>
              <Input
                value={newConfig.description}
                onChange={(e) => setNewConfig({ ...newConfig, description: e.target.value })}
                placeholder="配置项描述"
              />
            </div>
            
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">类型</label>
                <Select value={newConfig.type} onValueChange={(value) => setNewConfig({ ...newConfig, type: value })}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="string">字符串</SelectItem>
                    <SelectItem value="number">数字</SelectItem>
                    <SelectItem value="boolean">布尔值</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">环境</label>
                <Select value={newConfig.environment} onValueChange={(value) => setNewConfig({ ...newConfig, environment: value })}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {environments.map((env) => (
                      <SelectItem key={env.id} value={env.id}>
                        {env.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            
            <div className="flex items-center space-x-6">
              <div className="flex items-center space-x-2">
                <Checkbox 
                  checked={newConfig.required}
                  onCheckedChange={(checked) => setNewConfig({ ...newConfig, required: checked })}
                />
                <span className="text-sm">必需配置</span>
              </div>
              <div className="flex items-center space-x-2">
                <Checkbox 
                  checked={newConfig.secret}
                  onCheckedChange={(checked) => setNewConfig({ ...newConfig, secret: checked })}
                />
                <span className="text-sm">敏感信息</span>
              </div>
            </div>
            
            <div className="flex justify-end space-x-2">
              <Button variant="outline" onClick={() => setShowCreateDialog(false)}>
                取消
              </Button>
              <Button onClick={handleCreateConfig}>
                创建
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* 导入配置对话框 */}
      <Dialog open={showImportDialog} onOpenChange={setShowImportDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>导入配置</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">JSON配置数据</label>
              <Textarea
                value={importData}
                onChange={(e) => setImportData(e.target.value)}
                placeholder='{
  "DATABASE_URL": "postgresql://...",
  "REDIS_HOST": "localhost",
  "LOG_LEVEL": "info"
}'
                rows={10}
                className="font-mono"
              />
            </div>
            
            <div className="flex justify-end space-x-2">
              <Button variant="outline" onClick={() => setShowImportDialog(false)}>
                取消
              </Button>
              <Button onClick={handleImportConfig}>
                导入
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* 变更历史对话框 */}
      <Dialog open={showHistoryDialog} onOpenChange={setShowHistoryDialog}>
        <DialogContent className="max-w-4xl">
          <DialogHeader>
            <DialogTitle>配置变更历史</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            {configHistory.map((history) => (
              <Card key={history.id} className="border border-gray-200">
                <CardContent className="p-4">
                  <div className="flex justify-between items-start mb-2">
                    <div>
                      <h4 className="font-medium">{history.key}</h4>
                      <p className="text-sm text-gray-600">{history.reason}</p>
                    </div>
                    <div className="text-right">
                      <Badge className={
                        history.action === 'create' ? 'bg-green-100 text-green-800' :
                        history.action === 'update' ? 'bg-blue-100 text-blue-800' :
                        'bg-red-100 text-red-800'
                      }>
                        {history.action === 'create' ? '创建' : 
                         history.action === 'update' ? '更新' : '删除'}
                      </Badge>
                      <p className="text-xs text-gray-500 mt-1">{history.timestamp}</p>
                      <p className="text-xs text-gray-500">by {history.user}</p>
                    </div>
                  </div>
                  
                  {history.action === 'update' && (
                    <div className="grid grid-cols-2 gap-4 mt-3">
                      <div>
                        <label className="block text-xs font-medium text-gray-600 mb-1">旧值</label>
                        <div className="bg-red-50 p-2 rounded text-sm font-mono break-all">
                          {history.oldValue || 'N/A'}
                        </div>
                      </div>
                      <div>
                        <label className="block text-xs font-medium text-gray-600 mb-1">新值</label>
                        <div className="bg-green-50 p-2 rounded text-sm font-mono break-all">
                          {history.newValue}
                        </div>
                      </div>
                    </div>
                  )}
                  
                  {history.action === 'create' && (
                    <div className="mt-3">
                      <label className="block text-xs font-medium text-gray-600 mb-1">初始值</label>
                      <div className="bg-green-50 p-2 rounded text-sm font-mono break-all">
                        {history.newValue}
                      </div>
                    </div>
                  )}
                </CardContent>
              </Card>
            ))}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default ConfigurationManagement;