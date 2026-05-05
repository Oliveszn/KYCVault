import { CheckCircle } from "lucide-react";

interface PreviewStageProps {
  previewUrl: string;
  error?: string;
  onRetake: () => void;
  onConfirm: () => void;
}

export default function PreviewStage({
  previewUrl,
  error,
  onRetake,
  onConfirm,
}: PreviewStageProps) {
  return (
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
          src={previewUrl}
          alt="Selfie preview"
          className="w-full h-full object-cover scale-x-[-1]"
        />
      </div>

      {error && <p className="text-xs text-destructive mb-4">{error}</p>}

      <div className="flex gap-3">
        <button
          type="button"
          onClick={onRetake}
          className="flex items-center gap-2 px-5 py-2.5 rounded-lg border border-border text-sm font-medium bg-background hover:bg-muted transition-colors"
        >
          Retake
        </button>
        <button
          type="button"
          onClick={onConfirm}
          className="flex-1 flex items-center justify-center gap-2 py-2.5 rounded-lg bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 transition-colors"
        >
          <CheckCircle className="w-4 h-4" />
          Looks good
        </button>
      </div>
    </>
  );
}
