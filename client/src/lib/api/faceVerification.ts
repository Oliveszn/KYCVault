import { FaceVerification } from "@/types/faceVerification";
import { apiClient } from "../../api/client";
import { ApiResponse } from "@/types/kyc";

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
};
