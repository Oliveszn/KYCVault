import { configureStore } from "@reduxjs/toolkit";
import AuthReducer from "./auth-slice";

const store = configureStore({
  reducer: {
    auth: AuthReducer,
  },
});

export default store;
export type AppDispatch = typeof store.dispatch;
export type RootState = ReturnType<typeof store.getState>;
