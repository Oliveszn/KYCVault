import { configureStore } from "@reduxjs/toolkit";
import AuthReducer from "./auth-slice";
import KYCReducer from "./kyc-slice";
import { persistKYCeMiddleware } from "@/middleware/PersistKYCData";

const store = configureStore({
  reducer: {
    auth: AuthReducer,
    createKyc: KYCReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {},
    }).concat(persistKYCeMiddleware),
});

export default store;
export type AppDispatch = typeof store.dispatch;
export type RootState = ReturnType<typeof store.getState>;
