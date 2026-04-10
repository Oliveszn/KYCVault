import { faceVerifyApi } from "@/lib/api/faceVerification";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

export const faceKeys = {
  all: ["face"] as const,

  verification: (sessionId: string) =>
    [...faceKeys.all, "verification", sessionId] as const,
};

export const useStartFaceVerification = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ sessionId, file }: { sessionId: string; file: File }) =>
      faceVerifyApi.startFaceVerification(sessionId, file),

    onSuccess: (_, { sessionId }) => {
      // immediately refetch face verification (starts polling)
      queryClient.invalidateQueries({
        queryKey: faceKeys.verification(sessionId),
      });

      // session might change later (passed → in_review / next stage)
      queryClient.invalidateQueries({
        queryKey: ["kyc", "session", sessionId],
      });
    },
  });
};

export const useFaceVerification = (sessionId: string) => {
  return useQuery({
    queryKey: faceKeys.verification(sessionId),
    queryFn: () => faceVerifyApi.getFaceVerification(sessionId),
    enabled: !!sessionId,
  });
};

export const useFaceVerificationPolling = (sessionId: string) => {
  return useQuery({
    queryKey: faceKeys.verification(sessionId),
    queryFn: () => faceVerifyApi.getFaceVerification(sessionId),
    enabled: !!sessionId,

    refetchInterval: (query) => {
      const data = query.state.data;
      if (!data) return 2000;

      return data.status === "pending" ? 2000 : false;
    },
  });
};

export const useFaceVerificationWithSessionSync = (sessionId: string) => {
  const queryClient = useQueryClient();

  return useQuery({
    queryKey: faceKeys.verification(sessionId),
    queryFn: () => faceVerifyApi.getFaceVerification(sessionId),
    enabled: !!sessionId,

    refetchInterval: (query) => {
      const data = query.state.data;
      if (!data) return 2000;

      const isPending = data.status === "pending";

      if (!isPending) {
        // 🔥 sync session when done
        queryClient.invalidateQueries({
          queryKey: ["kyc", "session", sessionId],
        });

        queryClient.invalidateQueries({
          queryKey: ["kyc", "active"],
        });
      }

      return isPending ? 2000 : false;
    },
  });
};
