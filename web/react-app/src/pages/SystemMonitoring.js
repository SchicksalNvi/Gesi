import React, { useState, useEffect } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Badge } from '../components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../components/ui/tabs';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '../components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../components/ui/select';
import { Switch } from '../components/ui/switch';
import { Progress } from '../components/ui/progress';
import { toast } from 'sonner';
import {
  Activity,
  AlertTriangle,
  Bell,
  BellOff,
  Cpu,
  HardDrive,
  MemoryStick,
  Network,
  Plus,
  Settings,
  TrendingUp,
  TrendingDown,
  Zap,
  Eye,
  EyeOff,
  RefreshCw,
  Download,
  Calendar,
  Clock
} from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, AreaChart, Area, BarChart, Bar } from 'recharts';

const SystemMonitoring = () => {
  const [systemMetrics, setSystemMetrics] = useState({});
  const [processMetrics, setProcessMetrics] = useState([]);
  const [alerts, setAlerts] = useState([]);
  const [alertRules, setAlertRules] = useState([]);
  const [historicalData, setHistoricalData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [refreshInterval, setRefreshInterval] = useState(5);
  const [showCreateAlertDialog, setShowCreateAlertDialog] = useState(false);
  const [selectedTimeRange, setSelectedTimeRange] = useState('1h');
  const [newAlert, setNewAlert] = useState({
    name: '',
    metric: 'cpu',
    condition: 'greater_than',
    threshold: 80,
    duration: 5,
    enabled: true,
    notification: {
      email: true,
      webhook: false,
      sms: false
    }
  });

  useEffect(() => {
    fetchData();
    if (autoRefresh) {
      const interval = setInterval(fetchData, refreshInterval * 1000);
      return () => clearInterval(interval);
    }
  }, [autoRefresh, refreshInterval]);

  const fetchData = async () => {
    try {
      setLoading(true);
      
      // 获取系统指标数据
      const metricsResponse = await fetch('/api/system/metrics', {
        credentials: 'include'
      });
      if (metricsResponse.ok) {
        const metrics = await metricsResponse.json();
        setSystemMetrics(metrics);
      }
      
      // 获取进程数据
      const processesResponse = await fetch('/api/system/processes', {
        credentials: 'include'
      });
      if (processesResponse.ok) {
        const processes = await processesResponse.json();
        setProcessMetrics(processes);
      }
      
      // 获取告警数据
      const alertsResponse = await fetch('/api/system/alerts', {
        credentials: 'include'
      });
      if (alertsResponse.ok) {
        const alerts = await alertsResponse.json();
        setAlerts(alerts);
      }
      
      // 获取告警规则
      const rulesResponse = await fetch('/api/system/alert-rules', {
        credentials: 'include'
      });
      if (rulesResponse.ok) {
        const rules = await rulesResponse.json();
        setAlertRules(rules);
      }
      
      // 获取历史数据
      const historyResponse = await fetch(`/api/system/history?range=${selectedTimeRange}`, {
        credentials: 'include'
      });
      if (historyResponse.ok) {
        const history = await historyResponse.json();
        setHistoricalData(history);
      }
    } catch (error) {
      console.error('Error fetching monitoring data:', error);
      toast.error('获取监控数据失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateAlert = async () => {
    try {
      const response = await fetch('/api/system/alert-rules', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        credentials: 'include',
        body: JSON.stringify(newAlert)
      });
      
      if (response.ok) {
        toast.success('告警规则创建成功');
        setShowCreateAlertDialog(false);
        setNewAlert({
          name: '',
          metric: 'cpu',
          condition: 'greater_than',
          threshold: 80,
          duration: 5,
          enabled: true,
          notification: {
            email: true,
            webhook: false,
            sms: false
          }
        });
        fetchData();
      } else {
        throw new Error('Failed to create alert rule');
      }
    } catch (error) {
      console.error('Error creating alert rule:', error);
      toast.error('创建告警规则失败');
    }
  };

  const handleToggleAlert = async (alertId, enabled) => {
    try {
      const response = await fetch(`/api/system/alert-rules/${alertId}`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json'
        },
        credentials: 'include',
        body: JSON.stringify({ enabled })
      });
      
      if (response.ok) {
        toast.success(`告警规则已${enabled ? '启用' : '禁用'}`);
        fetchData();
      } else {
        throw new Error('Failed to toggle alert rule');
      }
    } catch (error) {
      console.error('Error toggling alert rule:', error);
      toast.error('切换告警规则状态失败');
    }
  };

  const handleResolveAlert = async (alertId) => {
    try {
      const response = await fetch(`/api/system/alerts/${alertId}/resolve`, {
        method: 'POST',
        credentials: 'include'
      });
      
      if (response.ok) {
        toast.success('告警已标记为已解决');
        fetchData();
      } else {
        throw new Error('Failed to resolve alert');
      }
    } catch (error) {
      console.error('Error resolving alert:', error);
      toast.error('解决告警失败');
    }
  };

  const getAlertLevelColor = (level) => {
    switch (level) {
      case 'critical': return 'bg-red-100 text-red-800 border-red-200';
      case 'warning': return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'info': return 'bg-blue-100 text-blue-800 border-blue-200';
      default: return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'RUNNING': return 'bg-green-100 text-green-800';
      case 'STOPPED': return 'bg-red-100 text-red-800';
      case 'STARTING': return 'bg-yellow-100 text-yellow-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
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
          <h1 className="text-2xl font-bold text-gray-900">系统监控</h1>
          <p className="text-gray-600">实时监控系统性能和告警管理</p>
        </div>
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2">
            <Switch
              checked={autoRefresh}
              onCheckedChange={setAutoRefresh}
            />
            <span className="text-sm text-gray-600">自动刷新</span>
          </div>
          <Select value={refreshInterval.toString()} onValueChange={(value) => setRefreshInterval(parseInt(value))}>
            <SelectTrigger className="w-24">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="5">5s</SelectItem>
              <SelectItem value="10">10s</SelectItem>
              <SelectItem value="30">30s</SelectItem>
              <SelectItem value="60">60s</SelectItem>
            </SelectContent>
          </Select>
          <Button onClick={fetchData} variant="outline">
            <RefreshCw className="w-4 h-4 mr-2" />
            刷新
          </Button>
          <Dialog open={showCreateAlertDialog} onOpenChange={setShowCreateAlertDialog}>
            <DialogTrigger asChild>
              <Button className="bg-blue-600 hover:bg-blue-700">
                <Plus className="w-4 h-4 mr-2" />
                创建告警
              </Button>
            </DialogTrigger>
          </Dialog>
        </div>
      </div>

      <Tabs defaultValue="overview" className="space-y-4">
        <TabsList>
          <TabsTrigger value="overview">系统概览</TabsTrigger>
          <TabsTrigger value="processes">进程监控</TabsTrigger>
          <TabsTrigger value="alerts">告警管理</TabsTrigger>
          <TabsTrigger value="history">历史数据</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-4">
          {/* 系统指标卡片 */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <Card>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">CPU使用率</p>
                    <p className="text-2xl font-bold">{systemMetrics.cpu?.usage}%</p>
                    <p className="text-xs text-gray-500">{systemMetrics.cpu?.cores} 核心</p>
                  </div>
                  <div className="p-3 bg-blue-100 rounded-full">
                    <Cpu className="w-6 h-6 text-blue-600" />
                  </div>
                </div>
                <Progress value={systemMetrics.cpu?.usage} className="mt-3" />
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">内存使用率</p>
                    <p className="text-2xl font-bold">{systemMetrics.memory?.usage}%</p>
                    <p className="text-xs text-gray-500">
                      {formatBytes(systemMetrics.memory?.used * 1024 * 1024)} / {formatBytes(systemMetrics.memory?.total * 1024 * 1024)}
                    </p>
                  </div>
                  <div className="p-3 bg-green-100 rounded-full">
                    <MemoryStick className="w-6 h-6 text-green-600" />
                  </div>
                </div>
                <Progress value={systemMetrics.memory?.usage} className="mt-3" />
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">磁盘使用率</p>
                    <p className="text-2xl font-bold">{systemMetrics.disk?.usage}%</p>
                    <p className="text-xs text-gray-500">
                      {systemMetrics.disk?.used}GB / {systemMetrics.disk?.total}GB
                    </p>
                  </div>
                  <div className="p-3 bg-purple-100 rounded-full">
                    <HardDrive className="w-6 h-6 text-purple-600" />
                  </div>
                </div>
                <Progress value={systemMetrics.disk?.usage} className="mt-3" />
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">网络流量</p>
                    <p className="text-2xl font-bold">
                      {systemMetrics.network?.interfaces?.[0]?.rxRate}MB/s
                    </p>
                    <p className="text-xs text-gray-500">下载速率</p>
                  </div>
                  <div className="p-3 bg-orange-100 rounded-full">
                    <Network className="w-6 h-6 text-orange-600" />
                  </div>
                </div>
                <div className="mt-3 space-y-1">
                  <div className="flex justify-between text-xs">
                    <span>上传: {systemMetrics.network?.interfaces?.[0]?.txRate}MB/s</span>
                    <span>下载: {systemMetrics.network?.interfaces?.[0]?.rxRate}MB/s</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* 实时图表 */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <Card>
              <CardHeader>
                <CardTitle>CPU & 内存趋势</CardTitle>
              </CardHeader>
              <CardContent>
                <ResponsiveContainer width="100%" height={300}>
                  <LineChart data={historicalData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="time" />
                    <YAxis />
                    <Tooltip />
                    <Line type="monotone" dataKey="cpu" stroke="#3b82f6" strokeWidth={2} name="CPU" />
                    <Line type="monotone" dataKey="memory" stroke="#10b981" strokeWidth={2} name="内存" />
                  </LineChart>
                </ResponsiveContainer>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>磁盘 & 网络趋势</CardTitle>
              </CardHeader>
              <CardContent>
                <ResponsiveContainer width="100%" height={300}>
                  <AreaChart data={historicalData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="time" />
                    <YAxis />
                    <Tooltip />
                    <Area type="monotone" dataKey="disk" stackId="1" stroke="#8b5cf6" fill="#8b5cf6" fillOpacity={0.6} name="磁盘" />
                    <Area type="monotone" dataKey="network" stackId="2" stroke="#f59e0b" fill="#f59e0b" fillOpacity={0.6} name="网络" />
                  </AreaChart>
                </ResponsiveContainer>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="processes" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>进程性能监控</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {processMetrics.map((process) => (
                  <Card key={process.id} className="border border-gray-200">
                    <CardContent className="p-4">
                      <div className="flex justify-between items-start mb-3">
                        <div>
                          <h3 className="font-semibold text-lg">{process.name}</h3>
                          <div className="flex items-center space-x-4 mt-1">
                            <Badge className={getStatusColor(process.status)}>
                              {process.status}
                            </Badge>
                            <span className="text-sm text-gray-600">运行时间: {process.uptime}</span>
                            <span className="text-sm text-gray-600">重启次数: {process.restarts}</span>
                          </div>
                        </div>
                        <Button size="sm" variant="outline">
                          <Eye className="w-3 h-3 mr-1" />
                          详情
                        </Button>
                      </div>
                      
                      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        <div>
                          <div className="flex justify-between items-center mb-1">
                            <span className="text-sm text-gray-600">CPU使用率</span>
                            <span className="text-sm font-medium">{process.cpu}%</span>
                          </div>
                          <Progress value={process.cpu} className="h-2" />
                        </div>
                        
                        <div>
                          <div className="flex justify-between items-center mb-1">
                            <span className="text-sm text-gray-600">内存使用率</span>
                            <span className="text-sm font-medium">{process.memoryPercent}%</span>
                          </div>
                          <Progress value={process.memoryPercent} className="h-2" />
                          <div className="text-xs text-gray-500 mt-1">
                            {formatBytes(process.memory * 1024 * 1024)}
                          </div>
                        </div>
                        
                        <div className="flex items-center justify-center">
                          <div className="text-center">
                            <div className="text-lg font-semibold text-green-600">
                              {process.status === 'RUNNING' ? '正常' : '异常'}
                            </div>
                            <div className="text-xs text-gray-500">运行状态</div>
                          </div>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="alerts" className="space-y-4">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {/* 活跃告警 */}
            <Card>
              <CardHeader>
                <div className="flex justify-between items-center">
                  <CardTitle>活跃告警</CardTitle>
                  <Badge variant="destructive">
                    {alerts.filter(alert => alert.status === 'active').length}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {alerts.filter(alert => alert.status === 'active').map((alert) => (
                    <div key={alert.id} className={`p-3 rounded-lg border ${getAlertLevelColor(alert.level)}`}>
                      <div className="flex justify-between items-start mb-2">
                        <div>
                          <h4 className="font-medium">{alert.name}</h4>
                          <p className="text-sm opacity-80">{alert.message}</p>
                        </div>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleResolveAlert(alert.id)}
                        >
                          解决
                        </Button>
                      </div>
                      <div className="flex justify-between items-center text-xs">
                        <span>来源: {alert.source}</span>
                        <span>{alert.timestamp}</span>
                      </div>
                    </div>
                  ))}
                  {alerts.filter(alert => alert.status === 'active').length === 0 && (
                    <div className="text-center text-gray-500 py-8">
                      <Bell className="w-8 h-8 mx-auto mb-2 opacity-50" />
                      <p>暂无活跃告警</p>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>

            {/* 告警规则 */}
            <Card>
              <CardHeader>
                <div className="flex justify-between items-center">
                  <CardTitle>告警规则</CardTitle>
                  <Button onClick={() => setShowCreateAlertDialog(true)} size="sm">
                    <Plus className="w-3 h-3 mr-1" />
                    新建
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {alertRules.map((rule) => (
                    <div key={rule.id} className="p-3 border rounded-lg">
                      <div className="flex justify-between items-start mb-2">
                        <div>
                          <h4 className="font-medium">{rule.name}</h4>
                          <p className="text-sm text-gray-600">
                            {rule.metric} {rule.condition === 'greater_than' ? '>' : '<'} {rule.threshold}%
                          </p>
                        </div>
                        <Switch
                          checked={rule.enabled}
                          onCheckedChange={(checked) => handleToggleAlert(rule.id, checked)}
                        />
                      </div>
                      <div className="flex justify-between items-center text-xs text-gray-500">
                        <span>持续时间: {rule.duration}分钟</span>
                        <span>
                          {rule.lastTriggered ? `最后触发: ${rule.lastTriggered}` : '从未触发'}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>

          {/* 告警历史 */}
          <Card>
            <CardHeader>
              <CardTitle>告警历史</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {alerts.map((alert) => (
                  <div key={alert.id} className="flex items-center justify-between p-3 border rounded-lg">
                    <div className="flex items-center space-x-3">
                      <div className={`w-3 h-3 rounded-full ${
                        alert.status === 'active' ? 'bg-red-500' : 'bg-green-500'
                      }`}></div>
                      <div>
                        <h4 className="font-medium">{alert.name}</h4>
                        <p className="text-sm text-gray-600">{alert.message}</p>
                      </div>
                    </div>
                    <div className="text-right">
                      <Badge className={getAlertLevelColor(alert.level)}>
                        {alert.level}
                      </Badge>
                      <p className="text-xs text-gray-500 mt-1">{alert.timestamp}</p>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="history" className="space-y-4">
          <Card>
            <CardHeader>
              <div className="flex justify-between items-center">
                <CardTitle>历史数据分析</CardTitle>
                <div className="flex items-center space-x-2">
                  <Select value={selectedTimeRange} onValueChange={setSelectedTimeRange}>
                    <SelectTrigger className="w-32">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="1h">最近1小时</SelectItem>
                      <SelectItem value="6h">最近6小时</SelectItem>
                      <SelectItem value="24h">最近24小时</SelectItem>
                      <SelectItem value="7d">最近7天</SelectItem>
                      <SelectItem value="30d">最近30天</SelectItem>
                    </SelectContent>
                  </Select>
                  <Button variant="outline" size="sm">
                    <Download className="w-3 h-3 mr-1" />
                    导出
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-6">
                {/* 系统资源使用趋势 */}
                <div>
                  <h3 className="text-lg font-semibold mb-3">系统资源使用趋势</h3>
                  <ResponsiveContainer width="100%" height={400}>
                    <LineChart data={historicalData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="time" />
                      <YAxis />
                      <Tooltip />
                      <Line type="monotone" dataKey="cpu" stroke="#3b82f6" strokeWidth={2} name="CPU (%)" />
                      <Line type="monotone" dataKey="memory" stroke="#10b981" strokeWidth={2} name="内存 (%)" />
                      <Line type="monotone" dataKey="disk" stroke="#8b5cf6" strokeWidth={2} name="磁盘 (%)" />
                    </LineChart>
                  </ResponsiveContainer>
                </div>

                {/* 性能统计 */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <Card>
                    <CardContent className="p-4">
                      <div className="text-center">
                        <div className="text-2xl font-bold text-blue-600">45.2%</div>
                        <div className="text-sm text-gray-600">平均CPU使用率</div>
                        <div className="flex items-center justify-center mt-2">
                          <TrendingUp className="w-4 h-4 text-green-500 mr-1" />
                          <span className="text-xs text-green-500">+2.1%</span>
                        </div>
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardContent className="p-4">
                      <div className="text-center">
                        <div className="text-2xl font-bold text-green-600">50.0%</div>
                        <div className="text-sm text-gray-600">平均内存使用率</div>
                        <div className="flex items-center justify-center mt-2">
                          <TrendingDown className="w-4 h-4 text-red-500 mr-1" />
                          <span className="text-xs text-red-500">-1.5%</span>
                        </div>
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardContent className="p-4">
                      <div className="text-center">
                        <div className="text-2xl font-bold text-purple-600">64.0%</div>
                        <div className="text-sm text-gray-600">平均磁盘使用率</div>
                        <div className="flex items-center justify-center mt-2">
                          <TrendingUp className="w-4 h-4 text-green-500 mr-1" />
                          <span className="text-xs text-green-500">+0.8%</span>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* 创建告警规则对话框 */}
      <Dialog open={showCreateAlertDialog} onOpenChange={setShowCreateAlertDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>创建告警规则</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">规则名称</label>
              <Input
                value={newAlert.name}
                onChange={(e) => setNewAlert({ ...newAlert, name: e.target.value })}
                placeholder="输入告警规则名称"
              />
            </div>
            
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">监控指标</label>
                <Select value={newAlert.metric} onValueChange={(value) => setNewAlert({ ...newAlert, metric: value })}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="cpu">CPU使用率</SelectItem>
                    <SelectItem value="memory">内存使用率</SelectItem>
                    <SelectItem value="disk">磁盘使用率</SelectItem>
                    <SelectItem value="network">网络流量</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              
              <div>
                <label className="block text-sm font-medium mb-1">条件</label>
                <Select value={newAlert.condition} onValueChange={(value) => setNewAlert({ ...newAlert, condition: value })}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="greater_than">大于</SelectItem>
                    <SelectItem value="less_than">小于</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">阈值 (%)</label>
                <Input
                  type="number"
                  value={newAlert.threshold}
                  onChange={(e) => setNewAlert({ ...newAlert, threshold: parseInt(e.target.value) })}
                  placeholder="80"
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium mb-1">持续时间 (分钟)</label>
                <Input
                  type="number"
                  value={newAlert.duration}
                  onChange={(e) => setNewAlert({ ...newAlert, duration: parseInt(e.target.value) })}
                  placeholder="5"
                />
              </div>
            </div>
            
            <div>
              <label className="block text-sm font-medium mb-2">通知方式</label>
              <div className="space-y-2">
                <div className="flex items-center space-x-2">
                  <Switch
                    checked={newAlert.notification.email}
                    onCheckedChange={(checked) => setNewAlert({
                      ...newAlert,
                      notification: { ...newAlert.notification, email: checked }
                    })}
                  />
                  <span className="text-sm">邮件通知</span>
                </div>
                <div className="flex items-center space-x-2">
                  <Switch
                    checked={newAlert.notification.webhook}
                    onCheckedChange={(checked) => setNewAlert({
                      ...newAlert,
                      notification: { ...newAlert.notification, webhook: checked }
                    })}
                  />
                  <span className="text-sm">Webhook通知</span>
                </div>
                <div className="flex items-center space-x-2">
                  <Switch
                    checked={newAlert.notification.sms}
                    onCheckedChange={(checked) => setNewAlert({
                      ...newAlert,
                      notification: { ...newAlert.notification, sms: checked }
                    })}
                  />
                  <span className="text-sm">短信通知</span>
                </div>
              </div>
            </div>
            
            <div className="flex justify-end space-x-2">
              <Button variant="outline" onClick={() => setShowCreateAlertDialog(false)}>
                取消
              </Button>
              <Button onClick={handleCreateAlert}>
                创建
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default SystemMonitoring;