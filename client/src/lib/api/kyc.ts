import {
  ApiResponse,
  ApproveSessionPayload,
  InitiateSessionPayload,
  KYCSessionResponse,
  RejectSessionPayload,
  SessionHistoryResponse,
  SessionQueueResponse,
  StatusCounts,
} from "@/types/kyc";
import { apiClient } from "../../api/client";

export const kycApi = {
  initiateSession: async (
    payload: InitiateSessionPayload,
  ): Promise<KYCSessionResponse> => {
    const { data } = await apiClient.post<ApiResponse<KYCSessionResponse>>(
      "/kyc/sessions",
      payload,
    );

    return data.payload;
  },

  getActiveSession: async (): Promise<KYCSessionResponse> => {
    const { data } = await apiClient.get<ApiResponse<KYCSessionResponse>>(
      "/kyc/sessions/active",
    );

    return data.payload;
  },

  getSession: async (sessionId: string): Promise<KYCSessionResponse> => {
    const { data } = await apiClient.get<ApiResponse<KYCSessionResponse>>(
      `/kyc/sessions/${sessionId}`,
    );

    return data.payload;
  },

  getSessionHistory: async (): Promise<SessionHistoryResponse> => {
    const { data } = await apiClient.get<ApiResponse<SessionHistoryResponse>>(
      "/kyc/sessions/history",
    );

    return data.payload;
  },

  //ADMIN

  getSessionQueue: async (
    limit = 20,
    offset = 0,
  ): Promise<SessionQueueResponse> => {
    const { data } = await apiClient.get<ApiResponse<SessionQueueResponse>>(
      `/admin/kyc/sessions?limit=${limit}&offset=${offset}`,
    );

    return data.payload;
  },

  getStatusCounts: async (): Promise<StatusCounts> => {
    const { data } = await apiClient.get<ApiResponse<StatusCounts>>(
      "/admin/kyc/sessions/counts",
    );

    return data.payload;
  },

  getSessionAdmin: async (sessionId: string): Promise<KYCSessionResponse> => {
    const { data } = await apiClient.get<ApiResponse<KYCSessionResponse>>(
      `/admin/kyc/sessions/${sessionId}`,
    );

    return data.payload;
  },

  approveSession: async (
    sessionId: string,
    payload: ApproveSessionPayload,
  ): Promise<void> => {
    await apiClient.post<ApiResponse<null>>(
      `/admin/kyc/sessions/${sessionId}/approve`,
      payload,
    );
  },

  rejectSession: async (
    sessionId: string,
    payload: RejectSessionPayload,
  ): Promise<void> => {
    await apiClient.post<ApiResponse<null>>(
      `/admin/kyc/sessions/${sessionId}/reject`,
      payload,
    );
  },
};
