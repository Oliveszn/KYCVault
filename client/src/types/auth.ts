export interface RegisterPayload {
  email: string;
  password: string;
  confirmPassword: string;
  firstName?: string;
  lastName?: string;
}

export interface LoginPayload {
  email: string;
  password: string;
}

export interface AuthResponse {
  accessToken: string;
  expiresIn: number;
  tokenType: "Bearer";
}

export interface UserResponse {
  id: number;
  email?: string;
  firstName?: string;
  lastName?: string;
  role: string;
}

export interface AuthState {
  accessToken: string | null;
  expiresAt: number | null; // Unix ms timestamp
  user: UserResponse | null;
  isAuthenticated: boolean;
}
