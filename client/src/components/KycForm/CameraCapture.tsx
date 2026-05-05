interface CameraCaptureProps {
  videoRef: React.RefObject<HTMLVideoElement | null>;
  onCapture: () => void;
  onCancel: () => void;
}

export default function CameraCapture({
  videoRef,
  onCapture,
  onCancel,
}: CameraCaptureProps) {
  return (
    <div className="rounded-lg overflow-hidden border border-border">
      <video
        ref={videoRef}
        className="w-full aspect-video object-cover"
        autoPlay
        muted
        playsInline
      />
      <div className="flex gap-2 p-3 bg-muted">
        <button
          type="button"
          onClick={onCapture}
          className="flex-1 py-2 text-sm font-medium bg-primary text-primary-foreground rounded-md"
        >
          Capture
        </button>
        <button
          type="button"
          onClick={onCancel}
          className="px-4 py-2 text-sm font-medium bg-background border border-border rounded-md"
        >
          Cancel
        </button>
      </div>
    </div>
  );
}
