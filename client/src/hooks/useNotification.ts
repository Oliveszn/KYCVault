import { notificationsApi } from "@/lib/api/notification";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

export const notifKeys = {
  all: ["notifications"] as const,
};

export const useNotifications = () => {
  return useQuery({
    queryKey: notifKeys.all,
    queryFn: notificationsApi.getAll,
  });
};

export const useMarkAsRead = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: notificationsApi.markAsRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notifKeys.all });
    },
  });
};
