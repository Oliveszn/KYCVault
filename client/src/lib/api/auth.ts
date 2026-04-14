import { ApiResponse } from "@/types/kyc";
import { apiClient } from "../../api/client";
import type {
  RegisterPayload,
  LoginPayload,
  AuthResponse,
  UserResponse,
} from "@/types/auth";

export const authApi = {
  register: async (
    payload: RegisterPayload,
  ): Promise<ApiResponse<AuthResponse>> => {
    const { data } = await apiClient.post<ApiResponse<AuthResponse>>(
      "/auth/register",
      payload,
    );
    return data;
  },

  login: async (payload: LoginPayload): Promise<ApiResponse<AuthResponse>> => {
    const { data } = await apiClient.post<ApiResponse<AuthResponse>>(
      "/auth/login",
      payload,
    );
    // only handle accesstoken as refresh is in the httpOnly cookie.
    return data;
  },

  logout: async (): Promise<void> => {
    await apiClient.post("/auth/logout");
  },

  logoutAll: async (): Promise<void> => {
    await apiClient.post("/auth/logout-all");
  },

  me: async (): Promise<{ id: number; role: string }> => {
    const { data } = await apiClient.get<{
      payload: { id: number; role: string };
    }>("/auth/me");
    return data.payload;
  },
};
