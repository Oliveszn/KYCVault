import { useRef, useState } from "react";

type CameraState = "idle" | "active" | "error";

interface UseDocumentCameraOptions {
  //    Which camera to prefer. Defaults to the environment-facing camera
  facingMode?: "user" | "environment";
  //  Filename used when creating the captured File. Defaults to "camera-capture.jpg".
  filename?: string;
}

interface UseDocumentCameraReturn {
  videoRef: React.RefObject<HTMLVideoElement | null>;
  cameraState: CameraState;
  cameraError: string | null;
  openCamera: () => Promise<void>;
  capturePhoto: () => Promise<File | null>;
  closeCamera: () => void;
}

export function useDocumentCamera(
  options: UseDocumentCameraOptions = {},
): UseDocumentCameraReturn {
  const { facingMode, filename = "camera-capture.jpg" } = options;

  const videoRef = useRef<HTMLVideoElement>(null);
  const [cameraState, setCameraState] = useState<CameraState>("idle");
  const [cameraError, setCameraError] = useState<string | null>(null);

  const openCamera = async () => {
    setCameraError(null);
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: facingMode ? { facingMode } : true,
      });
      setCameraState("active");

      // Wait for the video element to be in the DOM after state update
      setTimeout(() => {
        if (videoRef.current) {
          videoRef.current.srcObject = stream;
          videoRef.current.play();
        }
      }, 100);
    } catch {
      setCameraState("error");
      setCameraError(
        "Camera access was denied. Please allow camera access or upload a file instead.",
      );
    }
  };

  const capturePhoto = (): Promise<File | null> => {
    return new Promise((resolve) => {
      const videoEl = videoRef.current;
      if (!videoEl) return resolve(null);

      const canvas = document.createElement("canvas");
      canvas.width = videoEl.videoWidth;
      canvas.height = videoEl.videoHeight;
      canvas.getContext("2d")?.drawImage(videoEl, 0, 0);

      canvas.toBlob((blob) => {
        if (!blob) return resolve(null);

        const file = new File([blob], filename, {
          type: "image/jpeg",
        });

        closeCamera();
        resolve(file);
      }, "image/jpeg");
    });
  };

  const closeCamera = () => {
    const stream = videoRef.current?.srcObject as MediaStream | null;
    stream?.getTracks().forEach((t) => t.stop());
    if (videoRef.current) videoRef.current.srcObject = null;
    setCameraState("idle");
  };

  return {
    videoRef,
    cameraState,
    cameraError,
    openCamera,
    capturePhoto,
    closeCamera,
  };
}
