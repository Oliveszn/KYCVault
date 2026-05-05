import { Camera, X } from "lucide-react";

interface CameraStageProps {
  videoRef: React.RefObject<HTMLVideoElement | null>;
  onCapture: () => void;
  onCancel: () => void;
}

export default function CameraStage({
  videoRef,
  onCapture,
  onCancel,
}: CameraStageProps) {
  return (
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
          onClick={onCancel}
          className="flex items-center gap-2 px-5 py-2.5 rounded-lg border border-border text-sm font-medium bg-background hover:bg-muted transition-colors"
        >
          <X className="w-4 h-4" />
          Cancel
        </button>
        <button
          type="button"
          onClick={onCapture}
          className="flex-1 flex items-center justify-center gap-2 py-2.5 rounded-lg bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 transition-colors"
        >
          <Camera className="w-4 h-4" />
          Take photo
        </button>
      </div>
    </>
  );
}
