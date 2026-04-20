import { ApiResponse } from "@/types/kyc";
import { apiClient } from "../../api/client";
import { Notification } from "@/types/notification";

export const notificationsApi = {
  getAll: async (): Promise<Notification[]> => {
    const { data } =
      await apiClient.get<ApiResponse<Notification[]>>("/notifications");
    return data.payload;
  },

  markAsRead: async (id: string): Promise<void> => {
    await apiClient.patch(`/notifications/${id}/read`);
  },
};
