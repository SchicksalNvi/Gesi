import React, { useState, useEffect, useCallback } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Badge } from '../components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../components/ui/tabs';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '../components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../components/ui/select';
import { Checkbox } from '../components/ui/checkbox';
import { DatePicker } from '../components/ui/date-picker';
import { toast } from 'sonner';
import {
  Search,
  Filter,
  Download,
  RefreshCw,
  Clock,
  AlertCircle,
  Info,
  CheckCircle,
  XCircle,
  Eye,
  BarChart3,
  TrendingUp,
  FileText,
  Zap,
  Database,
  Server,
  Globe,
  Activity
} from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar, PieChart as RechartsPieChart, Pie, Cell } from 'recharts';

const LogAnalysis = () => {
  const [logs, setLogs] = useState([]);
  const [filteredLogs, setFilteredLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedLevel, setSelectedLevel] = useState('all');
  const [selectedSource, setSelectedSource] = useState('all');
  const [selectedTimeRange, setSelectedTimeRange] = useState('24h');
  const [startDate, setStartDate] = useState(null);
  const [endDate, setEndDate] = useState(null);
  const [showAdvancedFilters, setShowAdvancedFilters] = useState(false);
  const [selectedLog, setSelectedLog] = useState(null);
  const [showLogDetail, setShowLogDetail] = useState(false);
  const [logStats, setLogStats] = useState({});
  const [trendData, setTrendData] = useState([]);
  const [levelDistribution, setLevelDistribution] = useState([]);
  const [sourceDistribution, setSourceDistribution] = useState([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(50);
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [refreshInterval, setRefreshInterval] = useState(30);

  useEffect(() => {
    fetchLogs();
    if (autoRefresh) {
      const interval = setInterval(fetchLogs, refreshInterval * 1000);
      return () => clearInterval(interval);
    }
  }, [autoRefresh, refreshInterval, selectedTimeRange, startDate, endDate, selectedLevel, selectedSource, searchTerm]);



  const applyFilters = useCallback(() => {
    let filtered = [...logs];
    
    // 搜索过滤
    if (searchTerm) {
      filtered = filtered.filter(log => 
        log.message.toLowerCase().includes(searchTerm.toLowerCase()) ||
        log.source.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }
    
    // 级别过滤
    if (selectedLevel !== 'all') {
      filtered = filtered.filter(log => log.level === selectedLevel);
    }
    
    // 来源过滤
    if (selectedSource !== 'all') {
      filtered = filtered.filter(log => log.source === selectedSource);
    }
    
    setFilteredLogs(filtered);
  }, [logs, searchTerm, selectedLevel, selectedSource]);

  useEffect(() => {
    applyFilters();
  }, [applyFilters]);

  const fetchLogs = async () => {
    try {
      setLoading(true);
      
      // 构建查询参数
      const params = new URLSearchParams();
      if (selectedLevel !== 'all') {
        params.append('level', selectedLevel);
      }
      if (selectedSource !== 'all') {
        params.append('source', selectedSource);
      }
      if (searchTerm) {
        params.append('search', searchTerm);
      }
      if (selectedTimeRange !== '24h') {
        params.append('timeRange', selectedTimeRange);
      }
      if (startDate) {
        params.append('startDate', startDate.toISOString());
      }
      if (endDate) {
        params.append('endDate', endDate.toISOString());
      }
      
      // 获取日志数据
      const response = await fetch(`/api/logs?${params.toString()}`, {
        credentials: 'include'
      });
      
      if (!response.ok) {
        throw new Error('Failed to fetch logs');
      }
      
      const data = await response.json();
      setLogs(data.logs || []);
      
      // 设置统计数据
      if (data.stats) {
        setLogStats(data.stats);
      }
      
      // 设置趋势数据
      if (data.trends) {
        setTrendData(data.trends);
      }
      
      // 设置级别分布
      if (data.levelDistribution) {
        setLevelDistribution(data.levelDistribution);
      }
      
      // 设置来源分布
      if (data.sourceDistribution) {
        setSourceDistribution(data.sourceDistribution);
      }
      
    } catch (error) {
      console.error('Error fetching logs:', error);
      toast.error('获取日志数据失败');
    } finally {
      setLoading(false);
    }
  };



  const handleExportLogs = () => {
    try {
      const csvContent = [
        ['时间', '级别', '来源', '消息'],
        ...filteredLogs.map(log => [
          new Date(log.timestamp).toLocaleString(),
          log.level,
          log.source,
          log.message
        ])
      ].map(row => row.join(',')).join('\n');
      
      const blob = new Blob([csvContent], { type: 'text/csv' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `logs-${new Date().toISOString().split('T')[0]}.csv`;
      a.click();
      URL.revokeObjectURL(url);
      
      toast.success('日志导出成功');
    } catch (error) {
      console.error('Error exporting logs:', error);
      toast.error('日志导出失败');
    }
  };

  const getLevelIcon = (level) => {
    switch (level) {
      case 'ERROR': return <XCircle className="w-4 h-4 text-red-500" />;
      case 'WARN': return <AlertCircle className="w-4 h-4 text-yellow-500" />;
      case 'INFO': return <Info className="w-4 h-4 text-blue-500" />;
      case 'DEBUG': return <CheckCircle className="w-4 h-4 text-gray-500" />;
      default: return <Info className="w-4 h-4 text-gray-500" />;
    }
  };

  const getLevelColor = (level) => {
    switch (level) {
      case 'ERROR': return 'bg-red-100 text-red-800 border-red-200';
      case 'WARN': return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'INFO': return 'bg-blue-100 text-blue-800 border-blue-200';
      case 'DEBUG': return 'bg-gray-100 text-gray-800 border-gray-200';
      default: return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  const getSourceIcon = (source) => {
    switch (source) {
      case 'web-server': return <Globe className="w-4 h-4" />;
      case 'api-server': return <Server className="w-4 h-4" />;
      case 'database': return <Database className="w-4 h-4" />;
      case 'redis': return <Zap className="w-4 h-4" />;
      case 'worker': return <Activity className="w-4 h-4" />;
      default: return <FileText className="w-4 h-4" />;
    }
  };

  const paginatedLogs = filteredLogs.slice(
    (currentPage - 1) * pageSize,
    currentPage * pageSize
  );

  const totalPages = Math.ceil(filteredLogs.length / pageSize);

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
          <h1 className="text-2xl font-bold text-gray-900">日志分析</h1>
          <p className="text-gray-600">日志搜索、过滤和分析</p>
        </div>
        <div className="flex items-center space-x-2">
          <Button onClick={fetchLogs} variant="outline">
            <RefreshCw className="w-4 h-4 mr-2" />
            刷新
          </Button>
          <Button onClick={handleExportLogs} variant="outline">
            <Download className="w-4 h-4 mr-2" />
            导出
          </Button>
        </div>
      </div>

      <Tabs defaultValue="logs" className="space-y-4">
        <TabsList>
          <TabsTrigger value="logs">日志查看</TabsTrigger>
          <TabsTrigger value="analysis">统计分析</TabsTrigger>
          <TabsTrigger value="trends">趋势图表</TabsTrigger>
        </TabsList>

        <TabsContent value="logs" className="space-y-4">
          {/* 搜索和过滤 */}
          <Card>
            <CardContent className="p-4">
              <div className="flex flex-wrap items-center gap-4">
                <div className="flex-1 min-w-64">
                  <div className="relative">
                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                    <Input
                      placeholder="搜索日志内容或来源..."
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      className="pl-10"
                    />
                  </div>
                </div>
                
                <Select value={selectedLevel} onValueChange={setSelectedLevel}>
                  <SelectTrigger className="w-32">
                    <SelectValue placeholder="级别" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">所有级别</SelectItem>
                    <SelectItem value="ERROR">ERROR</SelectItem>
                    <SelectItem value="WARN">WARN</SelectItem>
                    <SelectItem value="INFO">INFO</SelectItem>
                    <SelectItem value="DEBUG">DEBUG</SelectItem>
                  </SelectContent>
                </Select>
                
                <Select value={selectedSource} onValueChange={setSelectedSource}>
                  <SelectTrigger className="w-32">
                    <SelectValue placeholder="来源" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">所有来源</SelectItem>
                    <SelectItem value="web-server">web-server</SelectItem>
                    <SelectItem value="api-server">api-server</SelectItem>
                    <SelectItem value="database">database</SelectItem>
                    <SelectItem value="redis">redis</SelectItem>
                    <SelectItem value="worker">worker</SelectItem>
                  </SelectContent>
                </Select>
                
                <Select value={selectedTimeRange} onValueChange={setSelectedTimeRange}>
                  <SelectTrigger className="w-32">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="1h">最近1小时</SelectItem>
                    <SelectItem value="6h">最近6小时</SelectItem>
                    <SelectItem value="24h">最近24小时</SelectItem>
                    <SelectItem value="7d">最近7天</SelectItem>
                    <SelectItem value="custom">自定义</SelectItem>
                  </SelectContent>
                </Select>
                
                <Button
                  variant="outline"
                  onClick={() => setShowAdvancedFilters(!showAdvancedFilters)}
                >
                  <Filter className="w-4 h-4 mr-2" />
                  高级过滤
                </Button>
              </div>
              
              {showAdvancedFilters && (
                <div className="mt-4 p-4 bg-gray-50 rounded-lg space-y-4">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium mb-1">开始时间</label>
                      <DatePicker
                        selected={startDate}
                        onChange={setStartDate}
                        showTimeSelect
                        dateFormat="yyyy-MM-dd HH:mm"
                        className="w-full"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium mb-1">结束时间</label>
                      <DatePicker
                        selected={endDate}
                        onChange={setEndDate}
                        showTimeSelect
                        dateFormat="yyyy-MM-dd HH:mm"
                        className="w-full"
                      />
                    </div>
                  </div>
                  
                  <div className="flex items-center space-x-4">
                    <div className="flex items-center space-x-2">
                      <Checkbox id="auto-refresh" checked={autoRefresh} onCheckedChange={setAutoRefresh} />
                      <label htmlFor="auto-refresh" className="text-sm">自动刷新</label>
                    </div>
                    {autoRefresh && (
                      <Select value={refreshInterval.toString()} onValueChange={(value) => setRefreshInterval(parseInt(value))}>
                        <SelectTrigger className="w-24">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="10">10s</SelectItem>
                          <SelectItem value="30">30s</SelectItem>
                          <SelectItem value="60">60s</SelectItem>
                        </SelectContent>
                      </Select>
                    )}
                  </div>
                </div>
              )}
            </CardContent>
          </Card>

          {/* 统计概览 */}
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
            <Card>
              <CardContent className="p-4 text-center">
                <div className="text-2xl font-bold">{logStats?.total || 0}</div>
                <div className="text-sm text-gray-600">总计</div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4 text-center">
                <div className="text-2xl font-bold text-red-600">{logStats?.error || 0}</div>
                <div className="text-sm text-gray-600">错误</div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4 text-center">
                <div className="text-2xl font-bold text-yellow-600">{logStats?.warn || 0}</div>
                <div className="text-sm text-gray-600">警告</div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4 text-center">
                <div className="text-2xl font-bold text-blue-600">{logStats?.info || 0}</div>
                <div className="text-sm text-gray-600">信息</div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4 text-center">
                <div className="text-2xl font-bold text-gray-600">{logStats?.debug || 0}</div>
                <div className="text-sm text-gray-600">调试</div>
              </CardContent>
            </Card>
          </div>

          {/* 日志列表 */}
          <Card>
            <CardHeader>
              <div className="flex justify-between items-center">
                <CardTitle>日志记录 ({filteredLogs.length})</CardTitle>
                <div className="flex items-center space-x-2">
                  <span className="text-sm text-gray-600">每页显示</span>
                  <Select value={pageSize.toString()} onValueChange={(value) => setPageSize(parseInt(value))}>
                    <SelectTrigger className="w-20">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="25">25</SelectItem>
                      <SelectItem value="50">50</SelectItem>
                      <SelectItem value="100">100</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {paginatedLogs.map((log) => (
                  <div
                    key={log.id}
                    className="flex items-start space-x-3 p-3 border rounded-lg hover:bg-gray-50 cursor-pointer"
                    onClick={() => {
                      setSelectedLog(log);
                      setShowLogDetail(true);
                    }}
                  >
                    <div className="flex items-center space-x-2 min-w-0">
                      {getLevelIcon(log.level)}
                      <Badge className={getLevelColor(log.level)}>
                        {log.level}
                      </Badge>
                    </div>
                    
                    <div className="flex items-center space-x-2 min-w-0">
                      {getSourceIcon(log.source)}
                      <span className="text-sm font-medium">{log.source}</span>
                    </div>
                    
                    <div className="flex-1 min-w-0">
                      <p className="text-sm truncate">{log.message}</p>
                    </div>
                    
                    <div className="flex items-center space-x-2 text-xs text-gray-500">
                      <Clock className="w-3 h-3" />
                      <span>{new Date(log.timestamp).toLocaleString()}</span>
                    </div>
                    
                    <Button size="sm" variant="ghost">
                      <Eye className="w-3 h-3" />
                    </Button>
                  </div>
                ))}
              </div>
              
              {/* 分页 */}
              {totalPages > 1 && (
                <div className="flex justify-between items-center mt-4">
                  <div className="text-sm text-gray-600">
                    显示 {(currentPage - 1) * pageSize + 1} - {Math.min(currentPage * pageSize, filteredLogs.length)} 条，共 {filteredLogs.length} 条
                  </div>
                  <div className="flex space-x-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                      disabled={currentPage === 1}
                    >
                      上一页
                    </Button>
                    <span className="flex items-center px-3 text-sm">
                      {currentPage} / {totalPages}
                    </span>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
                      disabled={currentPage === totalPages}
                    >
                      下一页
                    </Button>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="analysis" className="space-y-4">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {/* 级别分布 */}
            <Card>
              <CardHeader>
                <CardTitle>日志级别分布</CardTitle>
              </CardHeader>
              <CardContent>
                <ResponsiveContainer width="100%" height={300}>
                  <RechartsPieChart>
                    <Pie
                      data={levelDistribution}
                      cx="50%"
                      cy="50%"
                      outerRadius={80}
                      fill="#8884d8"
                      dataKey="value"
                      label={({ name, value }) => `${name}: ${value}`}
                    >
                      {levelDistribution.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip />
                  </RechartsPieChart>
                </ResponsiveContainer>
              </CardContent>
            </Card>

            {/* 来源分布 */}
            <Card>
              <CardHeader>
                <CardTitle>日志来源分布</CardTitle>
              </CardHeader>
              <CardContent>
                <ResponsiveContainer width="100%" height={300}>
                  <BarChart data={sourceDistribution}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="name" />
                    <YAxis />
                    <Tooltip />
                    <Bar dataKey="value" fill="#3b82f6" />
                  </BarChart>
                </ResponsiveContainer>
              </CardContent>
            </Card>
          </div>

          {/* 详细统计 */}
          <Card>
            <CardHeader>
              <CardTitle>详细统计信息</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div>
                  <h3 className="font-semibold mb-3">错误率分析</h3>
                  <div className="space-y-2">
                    <div className="flex justify-between">
                      <span>错误率:</span>
                      <span className="font-medium text-red-600">
                        {logStats && logStats.total > 0 ? (((logStats.error || 0) / logStats.total) * 100).toFixed(1) : '0.0'}%
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span>警告率:</span>
                      <span className="font-medium text-yellow-600">
                        {logStats && logStats.total > 0 ? (((logStats.warn || 0) / logStats.total) * 100).toFixed(1) : '0.0'}%
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span>正常率:</span>
                      <span className="font-medium text-green-600">
                        {logStats && logStats.total > 0 ? ((((logStats.info || 0) + (logStats.debug || 0)) / logStats.total) * 100).toFixed(1) : '0.0'}%
                      </span>
                    </div>
                  </div>
                </div>
                
                <div>
                  <h3 className="font-semibold mb-3">服务状态</h3>
                  <div className="space-y-2">
                    {sourceDistribution.map((source) => (
                      <div key={source.name} className="flex justify-between">
                        <span>{source.name}:</span>
                        <span className="font-medium">{source.value} 条</span>
                      </div>
                    ))}
                  </div>
                </div>
                
                <div>
                  <h3 className="font-semibold mb-3">时间分析</h3>
                  <div className="space-y-2">
                    <div className="flex justify-between">
                      <span>最新日志:</span>
                      <span className="font-medium">
                        {logs.length > 0 ? new Date(logs[0].timestamp).toLocaleTimeString() : '-'}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span>最旧日志:</span>
                      <span className="font-medium">
                        {logs.length > 0 ? new Date(logs[logs.length - 1].timestamp).toLocaleTimeString() : '-'}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span>时间跨度:</span>
                      <span className="font-medium">24小时</span>
                    </div>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="trends" className="space-y-4">
          {/* 时间趋势图 */}
          <Card>
            <CardHeader>
              <CardTitle>24小时日志趋势</CardTitle>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={400}>
                <LineChart data={trendData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="time" />
                  <YAxis />
                  <Tooltip />
                  <Line type="monotone" dataKey="total" stroke="#3b82f6" strokeWidth={2} name="总计" />
                  <Line type="monotone" dataKey="error" stroke="#ef4444" strokeWidth={2} name="错误" />
                  <Line type="monotone" dataKey="warn" stroke="#f59e0b" strokeWidth={2} name="警告" />
                  <Line type="monotone" dataKey="info" stroke="#10b981" strokeWidth={2} name="信息" />
                </LineChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>

          {/* 趋势分析 */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Card>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">平均每小时日志</p>
                    <p className="text-2xl font-bold">{Math.round((logStats?.total || 0) / 24)}</p>
                  </div>
                  <div className="p-3 bg-blue-100 rounded-full">
                    <BarChart3 className="w-6 h-6 text-blue-600" />
                  </div>
                </div>
                <div className="mt-2 flex items-center">
                  <TrendingUp className="w-4 h-4 text-green-500 mr-1" />
                  <span className="text-sm text-green-500">+12% 比昨天</span>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">峰值时段</p>
                    <p className="text-2xl font-bold">14:00-15:00</p>
                  </div>
                  <div className="p-3 bg-orange-100 rounded-full">
                    <Clock className="w-6 h-6 text-orange-600" />
                  </div>
                </div>
                <div className="mt-2">
                  <span className="text-sm text-gray-500">最高 {trendData.length > 0 ? Math.max(...trendData.map(d => d.total || 0)) : 0} 条/小时</span>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">错误增长率</p>
                    <p className="text-2xl font-bold text-red-600">+5.2%</p>
                  </div>
                  <div className="p-3 bg-red-100 rounded-full">
                    <AlertCircle className="w-6 h-6 text-red-600" />
                  </div>
                </div>
                <div className="mt-2">
                  <span className="text-sm text-red-500">需要关注</span>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>

      {/* 日志详情对话框 */}
      <Dialog open={showLogDetail} onOpenChange={setShowLogDetail}>
        <DialogContent className="max-w-4xl">
          <DialogHeader>
            <DialogTitle>日志详情</DialogTitle>
          </DialogHeader>
          {selectedLog && (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium mb-1">时间</label>
                  <p className="text-sm bg-gray-50 p-2 rounded">
                    {new Date(selectedLog.timestamp).toLocaleString()}
                  </p>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">级别</label>
                  <Badge className={getLevelColor(selectedLog.level)}>
                    {selectedLog.level}
                  </Badge>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">来源</label>
                  <div className="flex items-center space-x-2">
                    {getSourceIcon(selectedLog.source)}
                    <span className="text-sm">{selectedLog.source}</span>
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">请求ID</label>
                  <p className="text-sm bg-gray-50 p-2 rounded font-mono">
                    {selectedLog.details.requestId}
                  </p>
                </div>
              </div>
              
              <div>
                <label className="block text-sm font-medium mb-1">消息</label>
                <p className="text-sm bg-gray-50 p-3 rounded">
                  {selectedLog.message}
                </p>
              </div>
              
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium mb-1">IP地址</label>
                  <p className="text-sm bg-gray-50 p-2 rounded font-mono">
                    {selectedLog.details.ip}
                  </p>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">状态码</label>
                  <p className="text-sm bg-gray-50 p-2 rounded">
                    {selectedLog.details.statusCode}
                  </p>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">处理时间</label>
                  <p className="text-sm bg-gray-50 p-2 rounded">
                    {selectedLog.details.duration}ms
                  </p>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">用户ID</label>
                  <p className="text-sm bg-gray-50 p-2 rounded">
                    {selectedLog.details.userId || 'N/A'}
                  </p>
                </div>
              </div>
              
              <div>
                <label className="block text-sm font-medium mb-1">User Agent</label>
                <p className="text-sm bg-gray-50 p-2 rounded font-mono">
                  {selectedLog.details.userAgent}
                </p>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default LogAnalysis;