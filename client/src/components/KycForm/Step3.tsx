import { useState, useRef, useEffect } from "react";
import { Sun, Eye, Camera, X, CheckCircle, Loader2 } from "lucide-react";

type Stage = "instructions" | "camera" | "preview" | "processing" | "done";

export default function FaceVerification() {
  const [stage, setStage] = useState<Stage>("instructions");
  const [selfie, setSelfie] = useState<string | null>(null);
  const [cameraError, setCameraError] = useState<string | null>(null);
  const videoRef = useRef<HTMLVideoElement>(null);
  const streamRef = useRef<MediaStream | null>(null);

  const startCamera = async () => {
    setCameraError(null);
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: "user" },
      });
      streamRef.current = stream;
      setStage("camera");
      setTimeout(() => {
        if (videoRef.current) {
          videoRef.current.srcObject = stream;
          videoRef.current.play();
        }
      }, 100);
    } catch {
      setCameraError(
        "Camera access was denied. Please allow camera access in your browser settings and try again.",
      );
    }
  };

  const stopCamera = () => {
    streamRef.current?.getTracks().forEach((t) => t.stop());
    streamRef.current = null;
  };

  const capture = () => {
    if (!videoRef.current) return;
    const canvas = document.createElement("canvas");
    canvas.width = videoRef.current.videoWidth;
    canvas.height = videoRef.current.videoHeight;
    canvas.getContext("2d")?.drawImage(videoRef.current, 0, 0);
    const dataUrl = canvas.toDataURL("image/jpeg");
    setSelfie(dataUrl);
    stopCamera();
    setStage("preview");
  };

  const retake = () => {
    setSelfie(null);
    startCamera();
  };

  const processVerification = () => {
    setStage("processing");
    // TODO: submit selfie to backend
    setTimeout(() => setStage("done"), 3000); // simulated
  };

  useEffect(() => {
    return () => stopCamera();
  }, []);

  const tips = [
    {
      icon: <Sun className="w-4 h-4" />,
      text: "Find an area with good lighting",
    },
    {
      icon: <Eye className="w-4 h-4" />,
      text: "Remove anything that covers your face",
    },
    {
      icon: (
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          className="w-4 h-4"
        >
          <circle cx="12" cy="12" r="10" />
          <line x1="4.93" y1="4.93" x2="19.07" y2="19.07" />
        </svg>
      ),
      text: "No glasses, they cause glare and reflections",
    },
  ];

  return (
    <div className="py-8 px-6 max-w-lg">
      {/* Instructions */}
      {stage === "instructions" && (
        <>
          <div className="mb-8">
            <h1 className="text-xl font-semibold text-foreground tracking-tight">
              Prepare for the camera
            </h1>
            <p className="text-sm text-muted-foreground mt-1 leading-relaxed">
              In a moment we'll ask you to take a selfie by smiling. This will
              let us know it's really you.
            </p>
          </div>

          <ul className="flex flex-col gap-3 mb-10">
            {tips.map((tip, i) => (
              <li
                key={i}
                className="flex items-center gap-3 px-4 py-3.5 rounded-lg bg-muted"
              >
                <div className="size-8 rounded-md bg-background flex items-center justify-center shrink-0 text-muted-foreground">
                  {tip.icon}
                </div>
                <span className="text-sm text-foreground">{tip.text}</span>
              </li>
            ))}
          </ul>

          {cameraError && (
            <div className="mb-6 text-xs text-destructive bg-destructive/10 px-4 py-3 rounded-lg">
              {cameraError}
            </div>
          )}

          <button
            type="button"
            onClick={startCamera}
            className="w-full flex items-center justify-center gap-2 py-2.5 rounded-lg bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 transition-colors"
          >
            <Camera className="w-4 h-4" />
            Continue
          </button>
        </>
      )}

      {/* Live camera */}
      {stage === "camera" && (
        <>
          <div className="mb-6">
            <h1 className="text-xl font-semibold text-foreground tracking-tight">
              Take your selfie
            </h1>
            <p className="text-sm text-muted-foreground mt-1">
              Look straight at the camera and smile naturally.
            </p>
          </div>

          <div className="relative rounded-xl overflow-hidden mb-6 bg-black aspect-[3/4]">
            <video
              ref={videoRef}
              className="w-full h-full object-cover scale-x-[-1]"
              autoPlay
              muted
              playsInline
            />
            {/* Face guide oval */}
            <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
              <div className="w-48 h-64 rounded-full border-2 border-white/60 border-dashed" />
            </div>
          </div>

          <div className="flex gap-3">
            <button
              type="button"
              onClick={() => {
                stopCamera();
                setStage("instructions");
              }}
              className="flex items-center gap-2 px-5 py-2.5 rounded-lg border border-border text-sm font-medium bg-background hover:bg-muted transition-colors"
            >
              <X className="w-4 h-4" />
              Cancel
            </button>
            <button
              type="button"
              onClick={capture}
              className="flex-1 flex items-center justify-center gap-2 py-2.5 rounded-lg bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 transition-colors"
            >
              <Camera className="w-4 h-4" />
              Take photo
            </button>
          </div>
        </>
      )}

      {/* Preview */}
      {stage === "preview" && selfie && (
        <>
          <div className="mb-6">
            <h1 className="text-xl font-semibold text-foreground tracking-tight">
              Confirm your selfie
            </h1>
            <p className="text-sm text-muted-foreground mt-1">
              Make sure your face is clearly visible and well-lit.
            </p>
          </div>

          <div className="relative rounded-xl overflow-hidden mb-6 aspect-[3/4] bg-black">
            <img
              src={selfie}
              alt="selfie preview"
              className="w-full h-full object-cover scale-x-[-1]"
            />
          </div>

          <div className="flex gap-3">
            <button
              type="button"
              onClick={retake}
              className="flex items-center gap-2 px-5 py-2.5 rounded-lg border border-border text-sm font-medium bg-background hover:bg-muted transition-colors"
            >
              Retake
            </button>
            <button
              type="button"
              onClick={processVerification}
              className="flex-1 flex items-center justify-center gap-2 py-2.5 rounded-lg bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 transition-colors"
            >
              <CheckCircle className="w-4 h-4" />
              Looks good
            </button>
          </div>
        </>
      )}

      {/* Processing */}
      {stage === "processing" && (
        <div className="flex flex-col items-center justify-center py-20 gap-4">
          <Loader2 className="w-10 h-10 text-primary animate-spin" />
          <p className="text-sm font-medium text-foreground">
            Verifying your identity...
          </p>
          <p className="text-xs text-muted-foreground text-center">
            This usually takes a few seconds. Please don't close this page.
          </p>
        </div>
      )}

      {/* Done */}
      {stage === "done" && (
        <div className="flex flex-col items-center justify-center py-20 gap-4">
          <div className="size-16 rounded-full bg-green-100 flex items-center justify-center">
            <CheckCircle className="w-8 h-8 text-green-600" />
          </div>
          <p className="text-lg font-semibold text-foreground">
            Verification complete
          </p>
          <p className="text-sm text-muted-foreground text-center">
            Your identity has been verified successfully.
          </p>
        </div>
      )}
    </div>
  );
}
