import { Camera } from "lucide-react";
import { tips } from "@/config/tips";

interface InstructionsStageProps {
  cameraError: string | null;
  onContinue: () => void;
}

export default function InstructionsStage({
  cameraError,
  onContinue,
}: InstructionsStageProps) {
  return (
    <>
      <div className="mb-8">
        <h1 className="text-xl font-semibold text-foreground tracking-tight">
          Prepare for the camera
        </h1>
        <p className="text-sm text-muted-foreground mt-1 leading-relaxed">
          In a moment we'll ask you to take a selfie by smiling. This will let
          us know it's really you.
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
        onClick={onContinue}
        className="w-full flex items-center justify-center gap-2 py-2.5 rounded-lg bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 transition-colors"
      >
        <Camera className="w-4 h-4" />
        Continue
      </button>
    </>
  );
}
