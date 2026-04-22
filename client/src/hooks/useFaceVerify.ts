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
    retry: false,
  });
};

export const useFaceVerificationAdmin = (sessionId: string) => {
  return useQuery({
    queryKey: ["face", "admin", sessionId],
    queryFn: () => faceVerifyApi.getFaceVerificationAdmin(sessionId),
    enabled: !!sessionId,
  });
};

export const useSelfieURL = (verificationId: string) => {
  return useQuery({
    queryKey: ["face", "selfie-url", verificationId],
    queryFn: () => faceVerifyApi.getSelfieURL(verificationId),
    enabled: !!verificationId,
  });
};

export const useReviewFaceVerification = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      verificationId,
      payload,
    }: {
      verificationId: string;
      payload: { passed: boolean; note: string };
    }) => faceVerifyApi.reviewFaceVerification(verificationId, payload),

    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["kyc", "queue"] });
      queryClient.invalidateQueries({ queryKey: ["kyc", "counts"] });
    },
  });
};
