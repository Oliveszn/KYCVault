import { FaceVerification } from "@/types/faceVerification";
import { apiClient } from "../../api/client";
import { ApiResponse } from "@/types/kyc";
import { PresignedURLResponse } from "@/types/document";

export const faceVerifyApi = {
  startFaceVerification: async (
    sessionId: string,
    file: File,
  ): Promise<FaceVerification> => {
    const formData = new FormData();
    formData.append("file", file);

    const { data } = await apiClient.post<ApiResponse<FaceVerification>>(
      `/kyc/sessions/${sessionId}/face`,
      formData,
      {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      },
    );

    return data.payload;
  },

  getFaceVerification: async (sessionId: string): Promise<FaceVerification> => {
    const { data } = await apiClient.get<ApiResponse<FaceVerification>>(
      `/kyc/sessions/${sessionId}/face`,
    );

    return data.payload;
  },

  getFaceVerificationAdmin: async (
    sessionId: string,
  ): Promise<FaceVerification> => {
    const { data } = await apiClient.get<ApiResponse<FaceVerification>>(
      `/admin/face/${sessionId}/face`,
    );
    return data.payload;
  },

  getSelfieURL: async (
    verificationId: string,
  ): Promise<PresignedURLResponse> => {
    const { data } = await apiClient.get<ApiResponse<PresignedURLResponse>>(
      `/admin/face/${verificationId}/selfie-url`,
    );
    return data.payload;
  },

  reviewFaceVerification: async (
    verificationId: string,
    payload: { passed: boolean; note: string },
  ): Promise<void> => {
    await apiClient.post(`/admin/face/${verificationId}/review`, payload);
  },
};
