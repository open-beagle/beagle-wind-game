/**
 * Mock数据配置
 * 通过vite.config.ts中定义的全局变量__USE_MOCK__来控制是否使用Mock数据
 */

declare const __USE_MOCK__: boolean;

/**
 * 判断当前是否应该使用Mock数据
 * @returns {boolean} 是否使用Mock数据
 */
export const shouldUseMock = (): boolean => {
  return __USE_MOCK__;
};

// 初始化时打印配置信息
console.log(
  `[配置] 当前使用数据源: ${shouldUseMock() ? "Mock数据" : "API数据"}`
);
