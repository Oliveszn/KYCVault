import { useDocumentURL } from "@/hooks/useDocument";
import { CheckCircle, Loader2 } from "lucide-react";

export default function ExistingDocumentPreview({
  docId,
  side,
  onClear,
}: {
  docId: string;
  side: "front" | "back";
  onClear: () => void;
}) {
  const { data: urlData, isLoading } = useDocumentURL(docId);

  if (isLoading) {
    return (
      <div className="rounded-lg border border-border aspect-video bg-muted flex items-center justify-center">
        <Loader2 className="w-5 h-5 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="relative rounded-lg overflow-hidden border border-border">
      <img
        src={urlData?.url}
        alt={`${side} side`}
        className="w-full aspect-video object-cover"
      />
      <div className="absolute top-2 right-2 bg-green-500 text-white rounded-full p-0.5">
        <CheckCircle className="w-4 h-4" />
      </div>
      <div className="absolute bottom-0 inset-x-0 bg-black/40 px-3 py-1.5 flex items-center justify-between">
        <span className="text-xs text-white/80">Already uploaded</span>
        <button
          type="button"
          onClick={onClear}
          className="text-xs text-white/70 hover:text-white underline"
        >
          Replace
        </button>
      </div>
    </div>
  );
}
