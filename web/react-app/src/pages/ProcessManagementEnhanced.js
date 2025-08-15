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
  Play,
  Square,
  RotateCcw,
  Plus,
  Edit,
  Trash2,
  Search,
  Clock,
  GitBranch,
  Layers,
  Settings,
  Calendar,
  Timer,
  ArrowRight,
  ChevronDown,
  ChevronRight
} from 'lucide-react';

const ProcessManagementEnhanced = () => {
  const [processes, setProcesses] = useState([]);
  const [groups, setGroups] = useState([]);
  const [dependencies, setDependencies] = useState([]);
  const [scheduledTasks, setScheduledTasks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedProcesses, setSelectedProcesses] = useState([]);
  const [showCreateGroupDialog, setShowCreateGroupDialog] = useState(false);
  const [showDependencyDialog, setShowDependencyDialog] = useState(false);
  const [showScheduleDialog, setShowScheduleDialog] = useState(false);
  const [expandedGroups, setExpandedGroups] = useState(new Set());
  const [newGroup, setNewGroup] = useState({ name: '', description: '', processes: [] });
  const [newDependency, setNewDependency] = useState({ source: '', target: '', type: 'start_after' });
  const [newSchedule, setNewSchedule] = useState({
    name: '',
    action: 'start',
    cron: '',
    processes: [],
    enabled: true
  });

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      
      // 获取进程数据
      try {
        const processesResponse = await fetch('/api/processes', {
          credentials: 'include'
        });
        if (processesResponse.ok) {
          const processes = await processesResponse.json();
          setProcesses(processes || []);
        } else {
          console.warn('Failed to fetch processes');
          setProcesses([]);
        }
      } catch (processError) {
        console.warn('Error fetching processes:', processError);
        setProcesses([]);
      }
      
      // 获取组数据
      try {
        const groupsResponse = await fetch('/api/groups', {
          credentials: 'include'
        });
        if (groupsResponse.ok) {
          const groupsData = await groupsResponse.json();
          setGroups(groupsData.groups || []);
        } else {
          console.warn('Failed to fetch groups');
          setGroups([]);
        }
      } catch (groupError) {
        console.warn('Error fetching groups:', groupError);
        setGroups([]);
      }
      
      // 获取依赖关系
      try {
        const dependenciesResponse = await fetch('/api/dependencies', {
          credentials: 'include'
        });
        if (dependenciesResponse.ok) {
          const dependencies = await dependenciesResponse.json();
          setDependencies(dependencies || []);
        } else {
          console.warn('Failed to fetch dependencies');
          setDependencies([]);
        }
      } catch (depError) {
        console.warn('Error fetching dependencies:', depError);
        setDependencies([]);
      }
      
      // 获取计划任务
      try {
        const tasksResponse = await fetch('/api/scheduled-tasks', {
          credentials: 'include'
        });
        if (tasksResponse.ok) {
          const tasks = await tasksResponse.json();
          setScheduledTasks(tasks || []);
        } else {
          console.warn('Failed to fetch scheduled tasks');
          setScheduledTasks([]);
        }
      } catch (taskError) {
        console.warn('Error fetching scheduled tasks:', taskError);
        setScheduledTasks([]);
      }
    } catch (error) {
      console.error('Error fetching data:', error);
      // 设置默认值确保页面不会崩溃
      setProcesses([]);
      setGroups([]);
      setDependencies([]);
      setScheduledTasks([]);
    } finally {
      setLoading(false);
    }
  };

  const handleBatchAction = async (action) => {
    if (selectedProcesses.length === 0) {
      toast.error('请选择要操作的进程');
      return;
    }

    try {
      // 模拟批量操作
      toast.success(`批量${action}操作已执行`);
      setSelectedProcesses([]);
      fetchData();
    } catch (error) {
      console.error('Error performing batch action:', error);
      toast.error('批量操作失败');
    }
  };

  const handleGroupAction = async (groupName, action) => {
    try {
      // 模拟分组操作
      toast.success(`分组 ${groupName} ${action} 操作已执行`);
      fetchData();
    } catch (error) {
      console.error('Error performing group action:', error);
      toast.error('分组操作失败');
    }
  };

  const handleCreateGroup = async () => {
    try {
      // 模拟创建分组
      toast.success('进程分组创建成功');
      setShowCreateGroupDialog(false);
      setNewGroup({ name: '', description: '', processes: [] });
      fetchData();
    } catch (error) {
      console.error('Error creating group:', error);
      toast.error('创建分组失败');
    }
  };

  const handleCreateDependency = async () => {
    try {
      const response = await fetch('/api/dependencies', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        credentials: 'include',
        body: JSON.stringify(newDependency)
      });
      
      if (response.ok) {
        toast.success('依赖关系创建成功');
        setShowDependencyDialog(false);
        setNewDependency({ source: '', target: '', type: 'start_after' });
        fetchData();
      } else {
        throw new Error('Failed to create dependency');
      }
    } catch (error) {
      console.error('Error creating dependency:', error);
      toast.error('创建依赖失败');
    }
  };

  const handleCreateSchedule = async () => {
    try {
      // 模拟创建定时任务
      toast.success('定时任务创建成功');
      setShowScheduleDialog(false);
      setNewSchedule({ name: '', action: 'start', cron: '', processes: [], enabled: true });
      fetchData();
    } catch (error) {
      console.error('Error creating schedule:', error);
      toast.error('创建定时任务失败');
    }
  };

  const toggleGroupExpansion = (groupId) => {
    const newExpanded = new Set(expandedGroups);
    if (newExpanded.has(groupId)) {
      newExpanded.delete(groupId);
    } else {
      newExpanded.add(groupId);
    }
    setExpandedGroups(newExpanded);
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'RUNNING': return 'bg-green-100 text-green-800';
      case 'STOPPED': return 'bg-red-100 text-red-800';
      case 'STARTING': return 'bg-yellow-100 text-yellow-800';
      case 'STOPPING': return 'bg-orange-100 text-orange-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const filteredProcesses = processes.filter(process =>
    process.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

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
          <h1 className="text-2xl font-bold text-gray-900">进程管理增强</h1>
          <p className="text-gray-600">进程分组、依赖管理和定时任务</p>
        </div>
        <div className="flex space-x-2">
          <Dialog open={showCreateGroupDialog} onOpenChange={setShowCreateGroupDialog}>
            <DialogTrigger asChild>
              <Button variant="outline">
                <Layers className="w-4 h-4 mr-2" />
                创建分组
              </Button>
            </DialogTrigger>
          </Dialog>
          <Dialog open={showDependencyDialog} onOpenChange={setShowDependencyDialog}>
            <DialogTrigger asChild>
              <Button variant="outline">
                <GitBranch className="w-4 h-4 mr-2" />
                设置依赖
              </Button>
            </DialogTrigger>
          </Dialog>
          <Dialog open={showScheduleDialog} onOpenChange={setShowScheduleDialog}>
            <DialogTrigger asChild>
              <Button className="bg-blue-600 hover:bg-blue-700">
                <Clock className="w-4 h-4 mr-2" />
                定时任务
              </Button>
            </DialogTrigger>
          </Dialog>
        </div>
      </div>

      <Tabs defaultValue="groups" className="space-y-4">
        <TabsList>
          <TabsTrigger value="groups">进程分组</TabsTrigger>
          <TabsTrigger value="dependencies">依赖管理</TabsTrigger>
          <TabsTrigger value="schedule">定时任务</TabsTrigger>
          <TabsTrigger value="batch">批量操作</TabsTrigger>
        </TabsList>

        <TabsContent value="groups" className="space-y-4">
          <Card>
            <CardHeader>
              <div className="flex justify-between items-center">
                <CardTitle>进程分组管理</CardTitle>
                <Button onClick={() => setShowCreateGroupDialog(true)}>
                  <Plus className="w-4 h-4 mr-2" />
                  创建分组
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {groups.map((group) => {
                  const groupProcesses = processes.filter(p => group.processes.includes(p.id));
                  const isExpanded = expandedGroups.has(group.id);
                  
                  return (
                    <Card key={group.id} className="border border-gray-200">
                      <CardContent className="p-4">
                        <div className="flex justify-between items-center mb-3">
                          <div className="flex items-center space-x-2">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => toggleGroupExpansion(group.id)}
                            >
                              {isExpanded ? (
                                <ChevronDown className="w-4 h-4" />
                              ) : (
                                <ChevronRight className="w-4 h-4" />
                              )}
                            </Button>
                            <div>
                              <h3 className="font-semibold text-lg">{group.name}</h3>
                              <p className="text-sm text-gray-600">{group.description}</p>
                            </div>
                          </div>
                          <div className="flex items-center space-x-2">
                            <Badge variant="outline">
                              {groupProcesses.length} 个进程
                            </Badge>
                            <div className="flex space-x-1">
                              <Button
                                size="sm"
                                variant="outline"
                                onClick={() => handleGroupAction(group.name, '启动')}
                                className="text-green-600"
                              >
                                <Play className="w-3 h-3" />
                              </Button>
                              <Button
                                size="sm"
                                variant="outline"
                                onClick={() => handleGroupAction(group.name, '停止')}
                                className="text-red-600"
                              >
                                <Square className="w-3 h-3" />
                              </Button>
                              <Button
                                size="sm"
                                variant="outline"
                                onClick={() => handleGroupAction(group.name, '重启')}
                                className="text-blue-600"
                              >
                                <RotateCcw className="w-3 h-3" />
                              </Button>
                            </div>
                          </div>
                        </div>
                        
                        {isExpanded && (
                          <div className="mt-4 space-y-2">
                            {groupProcesses.map((process) => (
                              <div key={process.id} className="flex items-center justify-between p-2 bg-gray-50 rounded">
                                <div className="flex items-center space-x-3">
                                  <span className="font-medium">{process.name}</span>
                                  <Badge className={getStatusColor(process.status)}>
                                    {process.status}
                                  </Badge>
                                  <span className="text-sm text-gray-500">{process.node}</span>
                                </div>
                                <div className="flex space-x-1">
                                  <Button size="sm" variant="ghost">
                                    <Play className="w-3 h-3" />
                                  </Button>
                                  <Button size="sm" variant="ghost">
                                    <Square className="w-3 h-3" />
                                  </Button>
                                  <Button size="sm" variant="ghost">
                                    <RotateCcw className="w-3 h-3" />
                                  </Button>
                                </div>
                              </div>
                            ))}
                          </div>
                        )}
                      </CardContent>
                    </Card>
                  );
                })}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="dependencies" className="space-y-4">
          <Card>
            <CardHeader>
              <div className="flex justify-between items-center">
                <CardTitle>依赖关系管理</CardTitle>
                <Button onClick={() => setShowDependencyDialog(true)}>
                  <Plus className="w-4 h-4 mr-2" />
                  添加依赖
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {dependencies.map((dependency) => (
                  <div key={dependency.id} className="flex items-center justify-between p-4 border rounded-lg">
                    <div className="flex items-center space-x-4">
                      <div className="text-center">
                        <div className="font-medium">{dependency.source}</div>
                        <div className="text-sm text-gray-500">源进程</div>
                      </div>
                      <ArrowRight className="w-5 h-5 text-gray-400" />
                      <div className="text-center">
                        <div className="font-medium">{dependency.target}</div>
                        <div className="text-sm text-gray-500">目标进程</div>
                      </div>
                      <Badge variant="outline">
                        {dependency.type === 'start_after' ? '启动后' : '停止前'}
                      </Badge>
                    </div>
                    <Button
                      size="sm"
                      variant="outline"
                      className="text-red-600"
                    >
                      <Trash2 className="w-3 h-3" />
                    </Button>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="schedule" className="space-y-4">
          <Card>
            <CardHeader>
              <div className="flex justify-between items-center">
                <CardTitle>定时任务管理</CardTitle>
                <Button onClick={() => setShowScheduleDialog(true)}>
                  <Plus className="w-4 h-4 mr-2" />
                  创建任务
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {scheduledTasks.map((task) => (
                  <Card key={task.id} className="border border-gray-200">
                    <CardContent className="p-4">
                      <div className="flex justify-between items-start mb-3">
                        <div>
                          <h3 className="font-semibold text-lg">{task.name}</h3>
                          <div className="flex items-center space-x-4 mt-1">
                            <Badge className={task.action === 'start' ? 'bg-green-100 text-green-800' : 
                                           task.action === 'stop' ? 'bg-red-100 text-red-800' : 
                                           'bg-blue-100 text-blue-800'}>
                              {task.action}
                            </Badge>
                            <span className="text-sm text-gray-600">{task.cron}</span>
                            <span className="text-sm text-gray-500">下次执行: {task.nextRun}</span>
                          </div>
                        </div>
                        <div className="flex items-center space-x-2">
                          <Badge variant={task.enabled ? 'default' : 'secondary'}>
                            {task.enabled ? '启用' : '禁用'}
                          </Badge>
                          <Button size="sm" variant="outline">
                            <Edit className="w-3 h-3" />
                          </Button>
                          <Button size="sm" variant="outline" className="text-red-600">
                            <Trash2 className="w-3 h-3" />
                          </Button>
                        </div>
                      </div>
                      
                      <div className="flex flex-wrap gap-1">
                        {task.processes.map((processName) => (
                          <Badge key={processName} variant="outline" className="text-xs">
                            {processName}
                          </Badge>
                        ))}
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="batch" className="space-y-4">
          <Card>
            <CardHeader>
              <div className="flex justify-between items-center">
                <CardTitle>批量操作</CardTitle>
                <div className="flex items-center space-x-2">
                  <div className="relative">
                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                    <Input
                      placeholder="搜索进程..."
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      className="pl-10 w-64"
                    />
                  </div>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="flex items-center space-x-2">
                  <span className="text-sm text-gray-600">
                    已选择 {selectedProcesses.length} 个进程
                  </span>
                  <div className="flex space-x-2">
                    <Button
                      size="sm"
                      onClick={() => handleBatchAction('启动')}
                      disabled={selectedProcesses.length === 0}
                      className="bg-green-600 hover:bg-green-700"
                    >
                      <Play className="w-3 h-3 mr-1" />
                      批量启动
                    </Button>
                    <Button
                      size="sm"
                      onClick={() => handleBatchAction('停止')}
                      disabled={selectedProcesses.length === 0}
                      className="bg-red-600 hover:bg-red-700"
                    >
                      <Square className="w-3 h-3 mr-1" />
                      批量停止
                    </Button>
                    <Button
                      size="sm"
                      onClick={() => handleBatchAction('重启')}
                      disabled={selectedProcesses.length === 0}
                      className="bg-blue-600 hover:bg-blue-700"
                    >
                      <RotateCcw className="w-3 h-3 mr-1" />
                      批量重启
                    </Button>
                  </div>
                </div>
                
                <div className="grid gap-2">
                  {filteredProcesses.map((process) => (
                    <div key={process.id} className="flex items-center justify-between p-3 border rounded-lg">
                      <div className="flex items-center space-x-3">
                        <Checkbox
                          checked={selectedProcesses.includes(process.id)}
                          onCheckedChange={(checked) => {
                            if (checked) {
                              setSelectedProcesses([...selectedProcesses, process.id]);
                            } else {
                              setSelectedProcesses(selectedProcesses.filter(id => id !== process.id));
                            }
                          }}
                        />
                        <span className="font-medium">{process.name}</span>
                        <Badge className={getStatusColor(process.status)}>
                          {process.status}
                        </Badge>
                        <span className="text-sm text-gray-500">{process.node}</span>
                        <Badge variant="outline">{process.group}</Badge>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* 创建分组对话框 */}
      <Dialog open={showCreateGroupDialog} onOpenChange={setShowCreateGroupDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>创建进程分组</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">分组名称</label>
              <Input
                value={newGroup.name}
                onChange={(e) => setNewGroup({ ...newGroup, name: e.target.value })}
                placeholder="输入分组名称"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">描述</label>
              <Input
                value={newGroup.description}
                onChange={(e) => setNewGroup({ ...newGroup, description: e.target.value })}
                placeholder="输入分组描述"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">选择进程</label>
              <div className="space-y-2 max-h-40 overflow-y-auto">
                {processes.map((process) => (
                  <div key={process.id} className="flex items-center space-x-2">
                    <Checkbox
                      checked={newGroup.processes.includes(process.id)}
                      onCheckedChange={(checked) => {
                        if (checked) {
                          setNewGroup({
                            ...newGroup,
                            processes: [...newGroup.processes, process.id]
                          });
                        } else {
                          setNewGroup({
                            ...newGroup,
                            processes: newGroup.processes.filter(id => id !== process.id)
                          });
                        }
                      }}
                    />
                    <span>{process.name}</span>
                    <Badge className={getStatusColor(process.status)}>
                      {process.status}
                    </Badge>
                  </div>
                ))}
              </div>
            </div>
            <div className="flex justify-end space-x-2">
              <Button variant="outline" onClick={() => setShowCreateGroupDialog(false)}>
                取消
              </Button>
              <Button onClick={handleCreateGroup}>
                创建
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* 创建依赖对话框 */}
      <Dialog open={showDependencyDialog} onOpenChange={setShowDependencyDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>设置进程依赖</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">源进程</label>
              <Select value={newDependency.source} onValueChange={(value) => setNewDependency({ ...newDependency, source: value })}>
                <SelectTrigger>
                  <SelectValue placeholder="选择源进程" />
                </SelectTrigger>
                <SelectContent>
                  {processes.map((process) => (
                    <SelectItem key={process.id} value={process.name}>
                      {process.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">目标进程</label>
              <Select value={newDependency.target} onValueChange={(value) => setNewDependency({ ...newDependency, target: value })}>
                <SelectTrigger>
                  <SelectValue placeholder="选择目标进程" />
                </SelectTrigger>
                <SelectContent>
                  {processes.map((process) => (
                    <SelectItem key={process.id} value={process.name}>
                      {process.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">依赖类型</label>
              <Select value={newDependency.type} onValueChange={(value) => setNewDependency({ ...newDependency, type: value })}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="start_after">启动后</SelectItem>
                  <SelectItem value="stop_before">停止前</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="flex justify-end space-x-2">
              <Button variant="outline" onClick={() => setShowDependencyDialog(false)}>
                取消
              </Button>
              <Button onClick={handleCreateDependency}>
                创建
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* 创建定时任务对话框 */}
      <Dialog open={showScheduleDialog} onOpenChange={setShowScheduleDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>创建定时任务</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">任务名称</label>
              <Input
                value={newSchedule.name}
                onChange={(e) => setNewSchedule({ ...newSchedule, name: e.target.value })}
                placeholder="输入任务名称"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">操作类型</label>
              <Select value={newSchedule.action} onValueChange={(value) => setNewSchedule({ ...newSchedule, action: value })}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="start">启动</SelectItem>
                  <SelectItem value="stop">停止</SelectItem>
                  <SelectItem value="restart">重启</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Cron表达式</label>
              <Input
                value={newSchedule.cron}
                onChange={(e) => setNewSchedule({ ...newSchedule, cron: e.target.value })}
                placeholder="例如: 0 2 * * * (每天凌晨2点)"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">选择进程</label>
              <div className="space-y-2 max-h-40 overflow-y-auto">
                {processes.map((process) => (
                  <div key={process.id} className="flex items-center space-x-2">
                    <Checkbox
                      checked={newSchedule.processes.includes(process.name)}
                      onCheckedChange={(checked) => {
                        if (checked) {
                          setNewSchedule({
                            ...newSchedule,
                            processes: [...newSchedule.processes, process.name]
                          });
                        } else {
                          setNewSchedule({
                            ...newSchedule,
                            processes: newSchedule.processes.filter(name => name !== process.name)
                          });
                        }
                      }}
                    />
                    <span>{process.name}</span>
                    <Badge className={getStatusColor(process.status)}>
                      {process.status}
                    </Badge>
                  </div>
                ))}
              </div>
            </div>
            <div className="flex justify-end space-x-2">
              <Button variant="outline" onClick={() => setShowScheduleDialog(false)}>
                取消
              </Button>
              <Button onClick={handleCreateSchedule}>
                创建
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default ProcessManagementEnhanced;