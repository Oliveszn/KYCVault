import { useDocumentURL } from "@/hooks/useDocument";
import { FileImage, Loader2 } from "lucide-react";

type Props = {
  docId: string;
  label: string;
};
export default function DocumentPreview({ docId, label }: Props) {
  const { data: url, isLoading, error } = useDocumentURL(docId);
  return (
    <div className="space-y-2">
      <p className="text-xs font-medium text-muted-foreground capitalize">
        {label}
      </p>
      <div className="aspect-video rounded-lg border border-border overflow-hidden bg-muted flex items-center justify-center">
        {isLoading && (
          <Loader2 size={20} className="animate-spin text-muted-foreground" />
        )}
        {url?.url && (
          <img
            src={url.url}
            alt={label}
            className="w-full h-full object-cover"
          />
        )}
        {!isLoading && !url && (
          <div className="flex flex-col items-center gap-2 text-muted-foreground">
            <FileImage size={24} />
            <span className="text-xs">No document</span>
          </div>
        )}
      </div>
    </div>
  );
}
