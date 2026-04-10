import {
  DocumentListResponse,
  KYCDocument,
  PresignedURLResponse,
} from "@/types/document";
import { apiClient } from "../../api/client";
import { ApiResponse } from "@/types/kyc";

export const docsApi = {
  uploadDocument: async (
    sessionId: string,
    file: File,
    side: "front" | "back",
  ): Promise<KYCDocument> => {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("side", side);

    const { data } = await apiClient.post<ApiResponse<KYCDocument>>(
      `/kyc/sessions/${sessionId}/documents`,
      formData,
      {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      },
    );

    return data.payload;
  },

  getDocuments: async (sessionId: string): Promise<DocumentListResponse> => {
    const { data } = await apiClient.get<ApiResponse<DocumentListResponse>>(
      `/kyc/sessions/${sessionId}/documents`,
    );

    return data.payload;
  },

  //ADMIN
  getDocumentURL: async (docId: string): Promise<PresignedURLResponse> => {
    const { data } = await apiClient.get<ApiResponse<PresignedURLResponse>>(
      `/admin/documents/${docId}/url`,
    );

    return data.payload;
  },
};
