import { useRef } from "react";
import { Upload, Camera, FileImage, X, CheckCircle } from "lucide-react";
import ExistingDocumentPreview from "./ExistingDocumentPreview";
import { useDocumentCamera } from "@/hooks/useDocumentCamera";
import CameraCapture from "./CameraCapture";
import { DocumentSummary } from "@/types/kyc";

export type SideFile = {
  file: File;
  preview: string;
};

interface DocumentSideInputProps {
  side: "front" | "back";
  file: SideFile | null;
  existingDoc?: DocumentSummary;
  error?: string;
  onFileChange: (side: "front" | "back", file: File) => void;
  onClear: (side: "front" | "back") => void;
}

export default function DocumentSideInput({
  side,
  file,
  existingDoc,
  error,
  onFileChange,
  onClear,
}: DocumentSideInputProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const {
    videoRef,
    cameraState,
    cameraError,
    openCamera,
    capturePhoto,
    closeCamera,
  } = useDocumentCamera();

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selected = e.target.files?.[0];
    if (selected) onFileChange(side, selected);
  };

  const handleCapture = async () => {
    const captured = await capturePhoto();
    if (captured) onFileChange(side, captured);
  };

  const handleExistingClear = () => {
    onClear(side);
    // Open file picker immediately so the user can replace it
    inputRef.current?.click();
  };

  return (
    <div>
      <label className="text-sm font-medium text-foreground mb-2 block capitalize">
        {side} side
      </label>

      {/* Hidden file input — always present so the ref is stable */}
      <input
        ref={inputRef}
        type="file"
        accept="image/*"
        className="hidden"
        onChange={handleFileInputChange}
      />

      {cameraState === "active" ? (
        <CameraCapture
          videoRef={videoRef}
          onCapture={handleCapture}
          onCancel={closeCamera}
        />
      ) : file ? (
        <FilePreview file={file} onRemove={() => onClear(side)} />
      ) : existingDoc ? (
        <ExistingDocumentPreview
          docId={existingDoc.id}
          side={side}
          onClear={handleExistingClear}
        />
      ) : (
        <EmptyState
          onUpload={() => inputRef.current?.click()}
          onCamera={openCamera}
        />
      )}

      {/* Camera-level error sits under the card */}
      {cameraError && (
        <p className="text-xs text-destructive mt-2">{cameraError}</p>
      )}

      {/* Form validation error */}
      {error && <p className="text-xs text-destructive mt-2">{error}</p>}
    </div>
  );
}

function FilePreview({
  file,
  onRemove,
}: {
  file: SideFile;
  onRemove: () => void;
}) {
  return (
    <div className="relative rounded-lg overflow-hidden border border-border">
      <img
        src={file.preview}
        alt="Document preview"
        className="w-full aspect-video object-cover"
      />
      <div className="absolute inset-0 bg-black/30 flex items-center justify-center opacity-0 hover:opacity-100 transition-opacity">
        <button
          type="button"
          onClick={onRemove}
          className="flex items-center gap-1.5 px-3 py-1.5 bg-white text-sm font-medium rounded-md text-gray-800"
        >
          <X className="w-3.5 h-3.5" />
          Remove
        </button>
      </div>
      <div className="absolute top-2 right-2 bg-green-500 text-white rounded-full p-0.5">
        <CheckCircle className="w-4 h-4" />
      </div>
    </div>
  );
}

function EmptyState({
  onUpload,
  onCamera,
}: {
  onUpload: () => void;
  onCamera: () => void;
}) {
  return (
    <div className="border border-dashed border-border rounded-lg p-6 flex flex-col items-center gap-4">
      <div className="size-10 rounded-full bg-muted flex items-center justify-center">
        <FileImage className="w-5 h-5 text-muted-foreground" />
      </div>
      <p className="text-xs text-muted-foreground text-center">
        Upload a photo or use your camera
      </p>
      <div className="flex gap-2">
        <button
          type="button"
          onClick={onUpload}
          className="flex items-center gap-2 px-4 py-2 text-sm font-medium border border-border rounded-md bg-background hover:bg-muted transition-colors"
        >
          <Upload className="w-4 h-4" />
          Upload file
        </button>
        <button
          type="button"
          onClick={onCamera}
          className="flex items-center gap-2 px-4 py-2 text-sm font-medium border border-border rounded-md bg-background hover:bg-muted transition-colors"
        >
          <Camera className="w-4 h-4" />
          Use camera
        </button>
      </div>
    </div>
  );
}
