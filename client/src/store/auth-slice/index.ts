import { AuthState, UserResponse } from "@/types/auth";
import { createSlice, PayloadAction } from "@reduxjs/toolkit";

const initialState: AuthState = {
  accessToken: null,
  expiresAt: null,
  user: null,
  isAuthenticated: false,
};

const authSlice = createSlice({
  name: "auth",
  initialState,
  reducers: {
    setCredentials: (
      state,
      action: PayloadAction<{ accessToken: string; expiresIn: number }>,
    ) => {
      const { accessToken, expiresIn } = action.payload;
      state.accessToken = accessToken;
      // Store expiry as a UTC timestamp so we can proactively refresh before expiry.
      state.expiresAt = Date.now() + expiresIn * 1000;
      state.isAuthenticated = true;
    },

    setUser: (state, action: PayloadAction<UserResponse>) => {
      state.user = action.payload;
    },

    clearCredentials: (state) => {
      state.accessToken = null;
      state.expiresAt = null;
      state.user = null;
      state.isAuthenticated = false;
    },
  },
});

export const { setCredentials, setUser, clearCredentials } = authSlice.actions;
export default authSlice.reducer;

//SELECTORS
export const selectAccessToken = (state: { auth: AuthState }) =>
  state.auth.accessToken;

export const selectIsAuthenticated = (state: { auth: AuthState }) =>
  state.auth.isAuthenticated;

export const selectUser = (state: { auth: AuthState }) => state.auth.user;

export const selectExpiresAt = (state: { auth: AuthState }) =>
  state.auth.expiresAt;

/** True if the access token will expire within the next `thresholdMs` ms. */
export const selectTokenExpiresSoon = (
  state: { auth: AuthState },
  thresholdMs = 60_000, // 1 minute before expiry
) => {
  const { expiresAt } = state.auth;
  if (!expiresAt) return false;
  return Date.now() >= expiresAt - thresholdMs;
};
