import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import {
  faceVerificationSchema,
  FaceVerificationValues,
} from "@/utils/validation/kycSchema";
import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate, useParams } from "react-router-dom";
import { useStartFaceVerification } from "@/hooks/useFaceVerify";
import type { FaceVerification } from "@/types/faceVerification";
import { toast } from "sonner";
import InstructionsStage from "./Faceverificationstages/InstructionStage";
import CameraStage from "./Faceverificationstages/CameraStage";
import PreviewStage from "./Faceverificationstages/PreviewStage";
import ProcessingStage from "./Faceverificationstages/ProcessingStage";
import DoneStage from "./Faceverificationstages/DoneStage";
import { useDocumentCamera } from "@/hooks/useDocumentCamera";

type Stage = "instructions" | "camera" | "preview" | "processing" | "done";

export default function FaceVerification() {
  const { id: sessionId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const startFaceVerification = useStartFaceVerification();

  const [stage, setStage] = useState<Stage>("instructions");
  const [selfiePreview, setSelfiePreview] = useState<string | null>(null);

  const {
    videoRef,
    cameraState,
    cameraError,
    openCamera,
    capturePhoto,
    closeCamera,
  } = useDocumentCamera({ facingMode: "user", filename: "selfie.jpg" });

  const {
    setValue,
    handleSubmit,
    formState: { errors },
  } = useForm<FaceVerificationValues>({
    resolver: zodResolver(faceVerificationSchema),
  });

  // Stop the stream if the component unmounts mid-flow
  useEffect(() => () => closeCamera(), []);

  const handleStartCamera = async () => {
    await openCamera();
    // Only advance if the camera opened without error
    if (cameraState !== "error") setStage("camera");
  };

  const handleCapture = async () => {
    const file = await capturePhoto();
    if (!file) return;
    setValue("selfie", file, { shouldValidate: true });
    setSelfiePreview(URL.createObjectURL(file));
    setStage("preview");
  };

  const handleCancelCamera = () => {
    closeCamera();
    setStage("instructions");
  };

  const handleRetake = async () => {
    setSelfiePreview(null);
    await openCamera();
    setStage("camera");
  };

  const onSubmit = (values: FaceVerificationValues) => {
    if (!sessionId) return;

    setStage("processing");

    startFaceVerification.mutate(
      { sessionId, file: values.selfie },
      {
        onSuccess: () => setStage("done"),
        onError: (err: any) => {
          const message = err?.response?.data?.message || "Upload failed";
          toast.error(message);
          setStage("preview");
        },
      },
    );
  };

  return (
    <div className="py-8 px-6 max-w-lg">
      {/* Instructions */}
      {stage === "instructions" && (
        <InstructionsStage
          cameraError={cameraError}
          onContinue={handleStartCamera}
        />
      )}

      {/* Live camera */}
      {stage === "camera" && (
        <CameraStage
          videoRef={videoRef}
          onCapture={handleCapture}
          onCancel={handleCancelCamera}
        />
      )}

      {/* Preview */}
      {stage === "preview" && selfiePreview && (
        <PreviewStage
          previewUrl={selfiePreview}
          error={errors.selfie?.message}
          onRetake={handleRetake}
          onConfirm={handleSubmit(onSubmit)}
        />
      )}

      {/* Processing */}
      {stage === "processing" && <ProcessingStage />}

      {/* Done */}
      {stage === "done" && (
        <DoneStage onNavigate={() => navigate("/dashboard")} />
      )}
    </div>
  );
}
