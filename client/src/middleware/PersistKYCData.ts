import { RootState } from "@/store/store";
import type { Middleware } from "@reduxjs/toolkit";

const STORAGE_KEY = "kycdata";

let timeout: ReturnType<typeof setTimeout>;

export const persistKYCeMiddleware: Middleware =
  (store) => (next) => (action: any) => {
    const result = next(action);

    // set it to only persist if the action belongs to the createCourse slice
    if (action.type?.startsWith("kyc-slice/")) {
      // dont persist the reset action just clear it
      if (action.type === "kyc-slice/resetForm") {
        clearPersistedData();
        return result;
      }

      clearTimeout(timeout);
      timeout = setTimeout(() => {
        const state: RootState = store.getState();
        const { formData, currentStep } = state.createKyc;

        ///Remove file object before serializing cos files cant store in localstorage
        const { ...serializableFormData } = formData;

        const payload = {
          currentStep,
          formData: serializableFormData,
        };

        try {
          localStorage.setItem(STORAGE_KEY, JSON.stringify(payload));
        } catch (e) {
          if (process.env.NODE_ENV === "development") {
            console.error("[createKYC] Failed to persist state:", e);
          }
        }
      }, 300);
    }

    return result;
  };

export const loadPersistedData = (): {
  currentStep: number;
  formData: Record<string, any>;
} | null => {
  if (typeof window === "undefined") {
    return null;
  }

  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    return JSON.parse(raw);
  } catch (e) {
    if (process.env.NODE_ENV === "development") {
      console.error("[createkyc] Failed to load persisted state:", e);
    }

    return null;
  }
};

export const clearPersistedData = () => {
  if (typeof window === "undefined") return;

  try {
    localStorage.removeItem(STORAGE_KEY);
  } catch (e) {
    if (process.env.NODE_ENV === "development") {
      console.error("create[kyc] Failed to clear persisted state:", e);
    }
  }
};
