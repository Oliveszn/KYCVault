import { CheckCircle } from "lucide-react";

interface DoneStageProps {
  onNavigate: () => void;
}

export default function DoneStage({ onNavigate }: DoneStageProps) {
  return (
    <div className="flex flex-col items-center justify-center py-20 gap-4">
      <div className="size-16 rounded-full bg-green-100 flex items-center justify-center">
        <CheckCircle className="w-8 h-8 text-green-600" />
      </div>
      <p className="text-lg font-semibold text-foreground">Selfie submitted</p>
      <p className="text-sm text-muted-foreground text-center">
        Your documents are under review. We'll notify you once verification is
        complete.
      </p>
      <button
        type="button"
        onClick={onNavigate}
        className="mt-2 flex items-center justify-center gap-2 px-6 py-2.5 rounded-lg bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 transition-colors"
      >
        Go to dashboard
      </button>
    </div>
  );
}
