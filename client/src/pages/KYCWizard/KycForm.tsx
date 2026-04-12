import ProgressBar from "@/components/KycForm/ProgressBar";
import InitiateSession from "@/components/KycForm/Step1";
import UploadDocument from "@/components/KycForm/Step2";
import FaceVerfication from "@/components/KycForm/Step3";
import { useAppSelector } from "@/store/hooks";
import { useEffect } from "react";
import { useNavigate } from "react-router-dom";

const stepRoutes: Record<number, string> = {
  1: "/verify/initiate-session",
  2: "/verify/upload-docs",
  3: "/verify/face-verify",
};

export default function KycForm() {
  const currentStep = useAppSelector((state) => state.createKyc.currentStep);
  const navigate = useNavigate();

  useEffect(() => {
    navigate(stepRoutes[currentStep]);
  }, [currentStep]);
  const renderStep = () => {
    switch (currentStep) {
      case 1:
        return <InitiateSession />;
      case 2:
        return <UploadDocument />;
      case 3:
        return <FaceVerfication />;
      default:
        return <InitiateSession />;
    }
  };
  return (
    <div className="min-h-screen">
      <div className="max-w-4xl mx-auto">
        <ProgressBar />

        <div className="flex flex-col items-center justify-center">
          {renderStep()}
        </div>
      </div>
    </div>
  );
}
