import { useSelfieURL } from "@/hooks/useFaceVerify";
import { FileImage, Loader2 } from "lucide-react";

type Props = {
  verificationId: string;
};

export default function SelfiePreview({ verificationId }: Props) {
  const { data: url, isLoading } = useSelfieURL(verificationId);

  return (
    <div>
      <h2 className="text-sm font-semibold mb-4">Selfie</h2>
      <div className="max-w-xs space-y-2">
        <div className="aspect-[3/4] rounded-lg border border-border overflow-hidden bg-muted flex items-center justify-center">
          {isLoading && (
            <Loader2 size={20} className="animate-spin text-muted-foreground" />
          )}
          {url?.url && (
            <img
              src={url.url}
              alt="selfie"
              className="w-full h-full object-cover"
            />
          )}
          {!isLoading && !url && (
            <div className="flex flex-col items-center gap-2 text-muted-foreground">
              <FileImage size={24} />
              <span className="text-xs">No selfie</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
