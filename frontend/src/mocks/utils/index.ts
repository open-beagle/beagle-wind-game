import { mockGamePlatforms } from "../data/GamePlatform";
import { mockGameCards } from "../data/GameCard";
import { mockGameNodes } from "../data/GameNode";
import { mockGameInstances } from "../data/GameInstance";

// 模拟延迟
export const delay = (ms: number) =>
  new Promise((resolve) => setTimeout(resolve, ms));

// 模拟分页
export const paginate = <T>(data: T[], page: number, pageSize: number) => {
  const start = (page - 1) * pageSize;
  const end = start + pageSize;
  return {
    list: data.slice(start, end),
    total: data.length,
  };
};

// 模拟搜索
export const search = <T extends Record<string, any>>(
  data: T[],
  keyword: string,
  fields: (keyof T)[]
) => {
  if (!keyword) return data;
  const lowerKeyword = keyword.toLowerCase();
  return data.filter((item) =>
    fields.some((field) => {
      const value = item[field];
      return value && String(value).toLowerCase().includes(lowerKeyword);
    })
  );
};

// 模拟数据生成器
export const generateMockData = () => {
  return {
    platforms: mockGamePlatforms,
    gameCards: mockGameCards,
    nodes: mockGameNodes,
    gameInstances: mockGameInstances,
  };
};

// 模拟 API 响应
export const mockResponse = <T>(data: T, code = 200, message = "success") => {
  return {
    code,
    message,
    data,
  };
};

// 模拟错误响应
export const mockError = (message = "error", code = 500) => {
  return {
    code,
    message,
    data: null,
  };
};
