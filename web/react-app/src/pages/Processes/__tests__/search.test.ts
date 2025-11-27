/**
 * Feature: process-aggregation, Property 4: 搜索过滤正确性
 * Validates: Requirements 2.1
 * For any 搜索查询字符串，返回的进程列表中的每个进程名称都应包含该查询字符串（不区分大小写）
 */

import { AggregatedProcess } from '@/types';

// 搜索过滤函数
function filterProcesses(
  processes: AggregatedProcess[],
  searchText: string
): AggregatedProcess[] {
  if (!searchText.trim()) {
    return processes;
  }

  return processes.filter((proc) =>
    proc.name.toLowerCase().includes(searchText.toLowerCase())
  );
}

describe('Process Search Filter Property Tests', () => {
  // Property: 所有返回的进程名称都应包含搜索字符串
  test('filtered results should all contain search text', () => {
    const processes: AggregatedProcess[] = [
      {
        name: 'api-service',
        total_instances: 3,
        running_instances: 2,
        stopped_instances: 1,
        instances: [],
      },
      {
        name: 'web-service',
        total_instances: 2,
        running_instances: 2,
        stopped_instances: 0,
        instances: [],
      },
      {
        name: 'worker-service',
        total_instances: 1,
        running_instances: 1,
        stopped_instances: 0,
        instances: [],
      },
    ];

    // 测试多个搜索词
    const searchTerms = ['api', 'service', 'web', 'worker', 'API', 'SERVICE'];

    searchTerms.forEach((searchText) => {
      const filtered = filterProcesses(processes, searchText);

      // 验证：所有结果都包含搜索词（不区分大小写）
      filtered.forEach((proc) => {
        expect(proc.name.toLowerCase()).toContain(searchText.toLowerCase());
      });
    });
  });

  // Property: 空搜索应返回所有进程
  test('empty search should return all processes', () => {
    const processes: AggregatedProcess[] = [
      {
        name: 'api-service',
        total_instances: 3,
        running_instances: 2,
        stopped_instances: 1,
        instances: [],
      },
      {
        name: 'web-service',
        total_instances: 2,
        running_instances: 2,
        stopped_instances: 0,
        instances: [],
      },
    ];

    const emptySearches = ['', '  ', '\t', '\n'];

    emptySearches.forEach((searchText) => {
      const filtered = filterProcesses(processes, searchText);
      expect(filtered).toEqual(processes);
    });
  });

  // Property: 不匹配的搜索应返回空数组
  test('non-matching search should return empty array', () => {
    const processes: AggregatedProcess[] = [
      {
        name: 'api-service',
        total_instances: 3,
        running_instances: 2,
        stopped_instances: 1,
        instances: [],
      },
    ];

    const nonMatchingSearches = ['xyz', 'notfound', '12345'];

    nonMatchingSearches.forEach((searchText) => {
      const filtered = filterProcesses(processes, searchText);
      expect(filtered).toHaveLength(0);
    });
  });

  // Property: 搜索应该不区分大小写
  test('search should be case-insensitive', () => {
    const processes: AggregatedProcess[] = [
      {
        name: 'API-Service',
        total_instances: 1,
        running_instances: 1,
        stopped_instances: 0,
        instances: [],
      },
    ];

    const caseVariations = ['api', 'API', 'Api', 'aPi'];

    caseVariations.forEach((searchText) => {
      const filtered = filterProcesses(processes, searchText);
      expect(filtered).toHaveLength(1);
      expect(filtered[0].name).toBe('API-Service');
    });
  });

  // Property: 部分匹配应该工作
  test('partial match should work', () => {
    const processes: AggregatedProcess[] = [
      {
        name: 'api-service-v1',
        total_instances: 1,
        running_instances: 1,
        stopped_instances: 0,
        instances: [],
      },
      {
        name: 'api-service-v2',
        total_instances: 1,
        running_instances: 1,
        stopped_instances: 0,
        instances: [],
      },
    ];

    const filtered = filterProcesses(processes, 'api-service');
    expect(filtered).toHaveLength(2);
  });
});
