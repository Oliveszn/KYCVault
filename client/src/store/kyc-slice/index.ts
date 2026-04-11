import { loadPersistedData } from "@/middleware/PersistKYCData";
import { createSlice, PayloadAction } from "@reduxjs/toolkit";

export interface KYCFormData {
  // Step 1
  country: string;
  IDType: string;

  // Step 2
  documents: {
    front?: File;
    back?: File;
  };

  // Step 3
  faceVerify?: File;
}

interface CreateKYCState {
  currentStep: number;
  formData: Partial<KYCFormData>;
  isSubmitting: boolean;
}

const persisted = loadPersistedData();

const initialState: CreateKYCState = {
  currentStep: persisted?.currentStep ?? 1,

  formData: {
    ...persisted?.formData,
    documents: {},
  },
  isSubmitting: false,
};

const createKYCSlice = createSlice({
  name: "createKYC",
  initialState,
  reducers: {
    setCurrentStep: (state, action: PayloadAction<number>) => {
      state.currentStep = action.payload;
    },
    nextStep: (state) => {
      if (state.currentStep < 3) {
        state.currentStep += 1;
      }
    },
    previousStep: (state) => {
      if (state.currentStep > 1) {
        state.currentStep -= 1;
      }
    },
    updateFormData: (state, action: PayloadAction<Partial<KYCFormData>>) => {
      state.formData = { ...state.formData, ...action.payload };
    },
    setSubmitting: (state, action: PayloadAction<boolean>) => {
      state.isSubmitting = action.payload;
    },
    resetForm: (state) => {
      state.currentStep = 1;

      state.isSubmitting = false;
    },
  },
});

export const {
  setCurrentStep,
  nextStep,
  previousStep,
  updateFormData,
  setSubmitting,
  resetForm,
} = createKYCSlice.actions;

export default createKYCSlice.reducer;
