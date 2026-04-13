import ProgressBar from "@/components/KycForm/ProgressBar";
import { Outlet } from "react-router-dom";

export default function KycForm() {
  return (
    <div className="min-h-screen">
      <div className="max-w-4xl mx-auto">
        <ProgressBar />

        <div className="flex flex-col items-center justify-center">
          <Outlet />
        </div>
      </div>
    </div>
  );
}
