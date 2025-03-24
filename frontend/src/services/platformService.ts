import api from "@/api";
import { DataService } from "./dataService";
import { mockGamePlatforms } from "@/mocks/data/GamePlatform";
import type { GamePlatform } from "@/types/GamePlatform";

class PlatformService extends DataService {
  /**
   * 获取平台列表
   */
  async getList(params?: any): Promise<{
    list: GamePlatform[];
    total: number;
  }> {
    if (this.useMock()) {
      await this.mockDelay(300);

      const { page = 1, pageSize = 10 } = params || {};
      const start = (page - 1) * pageSize;
      const end = start + pageSize;

      return {
        list: mockGamePlatforms.slice(start, end),
        total: mockGamePlatforms.length,
      };
    } else {
      try {
        const response = await api.platform.getList(params);
        return this.safelyExtractListData<GamePlatform>(response);
      } catch (error) {
        console.error("获取平台列表失败", error);
        return { list: [], total: 0 };
      }
    }
  }

  /**
   * 获取平台详情
   */
  async getDetail(id: string): Promise<GamePlatform | null> {
    if (this.useMock()) {
      await this.mockDelay(300);
      const platform = mockGamePlatforms.find((item) => item.id === id);
      return platform || null;
    } else {
      try {
        const response = await api.platform.getDetail(id);
        return this.safelyExtractData<GamePlatform | null>(response, null);
      } catch (error) {
        console.error("获取平台详情失败", error);
        return null;
      }
    }
  }

  /**
   * 创建平台
   */
  async create(data: any): Promise<string> {
    if (this.useMock()) {
      await this.mockDelay(500);
      return "mock-platform-id-" + Date.now();
    } else {
      try {
        const response = await api.platform.create(data);
        const result = this.safelyExtractData(response, { id: "" });
        return result.id || "";
      } catch (error) {
        console.error("创建平台失败", error);
        return "";
      }
    }
  }

  /**
   * 更新平台
   */
  async update(id: string, data: any): Promise<boolean> {
    if (this.useMock()) {
      await this.mockDelay(500);
      return true;
    } else {
      try {
        await api.platform.update(id, data);
        return true;
      } catch (error) {
        console.error("更新平台失败", error);
        return false;
      }
    }
  }

  /**
   * 删除平台
   */
  async delete(id: string): Promise<boolean> {
    if (this.useMock()) {
      await this.mockDelay(500);
      return true;
    } else {
      try {
        await api.platform.delete(id);
        return true;
      } catch (error) {
        console.error("删除平台失败", error);
        return false;
      }
    }
  }

  /**
   * 获取平台远程访问链接
   */
  async getAccess(id: string): Promise<string> {
    if (this.useMock()) {
      await this.mockDelay(300);
      console.log(`[PlatformService] 使用Mock数据生成平台访问链接: ${id}`);
      return `https://remote-access.example.com/${id}?token=mock-token-${Date.now()}`;
    } else {
      try {
        console.log(`[PlatformService] 正在请求平台访问链接: platformId=${id}`);
        const response = await api.platform.getAccess(id);
        console.log(`[PlatformService] 获取平台访问链接成功:`, response);

        if (!response || typeof response !== "object") {
          console.error(`[PlatformService] 响应格式错误:`, response);
          return "";
        }

        const url = this.safelyExtractData(response, { url: "" }).url || "";

        if (!url) {
          console.warn(`[PlatformService] 获取到的链接为空: platformId=${id}`);
        }

        return url;
      } catch (error) {
        console.error(`[PlatformService] 获取平台远程访问链接失败:`, {
          platformId: id,
          error,
        });
        throw error; // 将错误向上传播，让调用者处理
      }
    }
  }

  /**
   * 刷新平台远程访问链接
   */
  async refreshAccess(id: string): Promise<string> {
    if (this.useMock()) {
      await this.mockDelay(500);
      return `https://remote-access.example.com/${id}?token=mock-token-${Date.now()}`;
    } else {
      try {
        const response = await api.platform.refreshAccess(id);
        return this.safelyExtractData(response, { url: "" }).url || "";
      } catch (error) {
        console.error("刷新平台远程访问链接失败", error);
        return "";
      }
    }
  }
}

// 导出平台服务实例
export const platformService = new PlatformService();

// 导出便捷方法
export const getPlatformList = (params?: any) =>
  platformService.getList(params);
export const getPlatformDetail = (id: string) => platformService.getDetail(id);
export const createPlatform = (data: any) => platformService.create(data);
export const updatePlatform = (id: string, data: any) =>
  platformService.update(id, data);
export const deletePlatform = (id: string) => platformService.delete(id);
export const getPlatformAccess = (id: string) => platformService.getAccess(id);
export const refreshPlatformAccess = (id: string) =>
  platformService.refreshAccess(id);
