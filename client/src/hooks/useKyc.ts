import { kycApi } from "@/lib/api/kyc";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

export const kycKeys = {
  all: ["kyc"] as const,

  session: (id: string) => [...kycKeys.all, "session", id] as const,
  active: () => [...kycKeys.all, "active"] as const,
  history: () => [...kycKeys.all, "history"] as const,
  // admin
  queue: (params?: { limit?: number; offset?: number }) =>
    [...kycKeys.all, "queue", params] as const,

  counts: () => [...kycKeys.all, "counts"] as const,

  adminSession: (id: string) => [...kycKeys.all, "admin-session", id] as const,
};

export const useInitiateSession = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: kycApi.initiateSession,

    onSuccess: (data) => {
      // cache the new session
      queryClient.setQueryData(kycKeys.session(data.id), data);

      // invalidate active session
      queryClient.invalidateQueries({ queryKey: kycKeys.active() });
    },
  });
};

export const useActiveSession = () => {
  return useQuery({
    queryKey: kycKeys.active(),
    queryFn: kycApi.getActiveSession,
    retry: false, // don't retry 404
  });
};

export const useSession = (sessionId: string) => {
  return useQuery({
    queryKey: kycKeys.session(sessionId),
    queryFn: () => kycApi.getSession(sessionId),
    enabled: !!sessionId,
  });
};

export const useSessionHistory = () => {
  return useQuery({
    queryKey: kycKeys.history(),
    queryFn: kycApi.getSessionHistory,
  });
};

//ADMIN
export const useSessionQueue = (limit = 20, offset = 0) => {
  return useQuery({
    queryKey: kycKeys.queue({ limit, offset }),
    queryFn: () => kycApi.getSessionQueue(limit, offset),
  });
};

export const useStatusCounts = () => {
  return useQuery({
    queryKey: kycKeys.counts(),
    queryFn: kycApi.getStatusCounts,
  });
};

export const useAdminSession = (sessionId: string) => {
  return useQuery({
    queryKey: kycKeys.adminSession(sessionId),
    queryFn: () => kycApi.getSessionAdmin(sessionId),
    enabled: !!sessionId,
  });
};

export const useApproveSession = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      sessionId,
      payload,
    }: {
      sessionId: string;
      payload: { note: string };
    }) => kycApi.approveSession(sessionId, payload),

    onSuccess: (_, { sessionId }) => {
      // update specific session
      queryClient.invalidateQueries({
        queryKey: kycKeys.adminSession(sessionId),
      });

      queryClient.invalidateQueries({
        queryKey: kycKeys.session(sessionId),
      });

      // refresh queue  dashboard
      queryClient.invalidateQueries({ queryKey: kycKeys.queue() });
      queryClient.invalidateQueries({ queryKey: kycKeys.counts() });
    },
  });
};

export const useRejectSession = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      sessionId,
      payload,
    }: {
      sessionId: string;
      payload: { note: string; reason: string };
    }) => kycApi.rejectSession(sessionId, payload),

    onSuccess: (_, { sessionId }) => {
      queryClient.invalidateQueries({
        queryKey: kycKeys.adminSession(sessionId),
      });

      queryClient.invalidateQueries({
        queryKey: kycKeys.session(sessionId),
      });

      queryClient.invalidateQueries({ queryKey: kycKeys.queue() });
      queryClient.invalidateQueries({ queryKey: kycKeys.counts() });
    },
  });
};
