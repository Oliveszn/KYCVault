import { Loader2 } from "lucide-react";

export default function ProcessingStage() {
  return (
    <div className="flex flex-col items-center justify-center py-20 gap-4">
      <Loader2 className="w-10 h-10 text-primary animate-spin" />
      <p className="text-sm font-medium text-foreground">
        Verifying your identity...
      </p>
      <p className="text-xs text-muted-foreground text-center">
        This usually takes a few seconds. Please don't close this page.
      </p>
    </div>
  );
}
