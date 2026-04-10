import { docsApi } from "@/lib/api/document";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

export const documentKeys = {
  all: ["documents"] as const,

  session: (sessionId: string) =>
    [...documentKeys.all, "session", sessionId] as const,
};

export const useDocuments = (sessionId: string) => {
  return useQuery({
    queryKey: documentKeys.session(sessionId),
    queryFn: () => docsApi.getDocuments(sessionId),
    enabled: !!sessionId,
  });
};

export const useUploadDocument = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      sessionId,
      file,
      side,
    }: {
      sessionId: string;
      file: File;
      side: "front" | "back";
    }) => docsApi.uploadDocument(sessionId, file, side),

    onSuccess: (_, { sessionId }) => {
      // refresh documents list
      queryClient.invalidateQueries({
        queryKey: documentKeys.session(sessionId),
      });

      // refresh session
      queryClient.invalidateQueries({
        queryKey: ["kyc", "session", sessionId],
      });

      queryClient.invalidateQueries({
        queryKey: ["kyc", "active"],
      });
    },
  });
};

export const useDocumentURL = (docId: string) => {
  return useQuery({
    queryKey: ["documents", "url", docId],
    queryFn: () => docsApi.getDocumentURL(docId),
    enabled: !!docId,
  });
};
