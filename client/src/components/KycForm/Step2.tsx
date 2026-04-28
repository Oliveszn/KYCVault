import { useState, useRef } from "react";
import { Upload, Camera, FileImage, X, CheckCircle } from "lucide-react";
import { Separator } from "../ui/separator";
import FormNavigation from "./FormNavigation";
import {
  uploadDocumentSchema,
  UploadDocumentValues,
} from "@/utils/validation/kycSchema";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { updateFormData } from "@/store/kyc-slice";
import { useNavigate, useParams } from "react-router-dom";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import { useUploadDocument } from "@/hooks/useDocument";
import { toast } from "sonner";
import { idTypes } from "@/config/idtypes";
import { useSession } from "@/hooks/useKyc";
import ExistingDocumentPreview from "./ExistingDocumentPreview";
import { KYCDocument } from "@/types/document";

type SideFile = {
  file: File;
  preview: string;
} | null;

export default function UploadDocument() {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const [frontFile, setFrontFile] = useState<SideFile>(null);
  const [backFile, setBackFile] = useState<SideFile>(null);
  const [cameraError, setCameraError] = useState<string | null>(null);

  const frontInputRef = useRef<HTMLInputElement>(null);
  const backInputRef = useRef<HTMLInputElement>(null);
  const frontVideoRef = useRef<HTMLVideoElement>(null);
  const backVideoRef = useRef<HTMLVideoElement>(null);

  const [activeCam, setActiveCam] = useState<"front" | "back" | null>(null);

  const { id: sessionId } = useParams<{ id: string }>();
  const { data: session } = useSession(sessionId!);

  // find existing docs from session
  const existingFront = session?.documents?.find((d) => d.side === "front");
  const existingBack = session?.documents?.find((d) => d.side === "back");

  const uploadDocument = useUploadDocument();
  const {
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<UploadDocumentValues>({
    resolver: zodResolver(uploadDocumentSchema),
    // pre-mark as valid if docs already exist so form doesn't block submission
    defaultValues: {
      front: existingFront ? (true as any) : undefined,
      back: existingBack ? (true as any) : undefined,
    },
  });

  const formData = useAppSelector((state) => state.createKyc.formData);

  const country = session?.country || formData.country;
  const idType = session?.id_type || formData.IDType;
  const selectedIdType = idTypes.find((t) => t.value === idType);

  // const onSubmit = (values: UploadDocumentValues) => {
  //   if (!sessionId) {
  //     toast.error("Session not found");
  //     return;
  //   }

  //   // upload front first, then back on success
  //   uploadDocument.mutate(
  //     { sessionId, file: values.front, side: "front" },
  //     {
  //       onSuccess: () => {
  //         uploadDocument.mutate(
  //           { sessionId, file: values.back, side: "back" },
  //           {
  //             onSuccess: () => {
  //               dispatch(
  //                 updateFormData({
  //                   documents: { front: values.front, back: values.back },
  //                 }),
  //               );
  //               toast.success("Documents uploaded successfully");
  //               navigate(`/kyc/sessions/${sessionId}/face`);
  //             },
  //             onError: (err: any) => {
  //               const message =
  //                 err?.response?.data?.message || "Failed to upload back side";
  //               toast.error(message);
  //             },
  //           },
  //         );
  //       },
  //       onError: (err: any) => {
  //         const message =
  //           err?.response?.data?.message || "Failed to upload front side";
  //         toast.error(message);
  //       },
  //     },
  //   );
  // };
  const onSubmit = async (values: UploadDocumentValues) => {
    if (!sessionId) return;

    const uploads: Promise<KYCDocument>[] = [];

    // only upload sides that were freshly selected
    if (values.front && !existingFront) {
      uploads.push(
        uploadDocument.mutateAsync({
          sessionId,
          file: values.front,
          side: "front",
        }),
      );
    }
    if (values.back && !existingBack) {
      uploads.push(
        uploadDocument.mutateAsync({
          sessionId,
          file: values.back,
          side: "back",
        }),
      );
    }

    // if both already exist and user didn't replace either, just advance
    if (uploads.length === 0) {
      navigate(`/kyc/sessions/${sessionId}/face`);
      return;
    }

    try {
      await Promise.all(uploads);
      dispatch(
        updateFormData({
          documents: { front: values.front, back: values.back },
        }),
      );
      toast.success("Documents uploaded successfully");
      navigate(`/kyc/sessions/${sessionId}/face`);
    } catch (err: any) {
      const message = err?.response?.data?.message || "Upload failed";
      toast.error(message);
    }
  };

  const handleFileChange = (
    e: React.ChangeEvent<HTMLInputElement>,
    side: "front" | "back",
  ) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const preview = URL.createObjectURL(file);
    setValue(side, file, { shouldValidate: true });
    if (side === "front") setFrontFile({ file, preview });
    else setBackFile({ file, preview });
  };

  const openCamera = async (side: "front" | "back") => {
    setCameraError(null);
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ video: true });
      setActiveCam(side);
      setTimeout(() => {
        const videoEl =
          side === "front" ? frontVideoRef.current : backVideoRef.current;
        if (videoEl) {
          videoEl.srcObject = stream;
          videoEl.play();
        }
      }, 100);
    } catch {
      setCameraError(
        "Camera access was denied. Please allow camera access or upload a file instead.",
      );
    }
  };

  const capturePhoto = (side: "front" | "back") => {
    const videoEl =
      side === "front" ? frontVideoRef.current : backVideoRef.current;
    if (!videoEl) return;
    const canvas = document.createElement("canvas");
    canvas.width = videoEl.videoWidth;
    canvas.height = videoEl.videoHeight;
    canvas.getContext("2d")?.drawImage(videoEl, 0, 0);
    canvas.toBlob((blob) => {
      if (!blob) return;
      const file = new File([blob], `${side}-capture.jpg`, {
        type: "image/jpeg",
      });
      setValue(side, file, { shouldValidate: true });
      const preview = URL.createObjectURL(blob);
      if (side === "front") setFrontFile({ file, preview });
      else setBackFile({ file, preview });
      const stream = videoEl.srcObject as MediaStream;
      stream?.getTracks().forEach((t) => t.stop());
      setActiveCam(null);
    }, "image/jpeg");
  };

  const clearFile = (side: "front" | "back") => {
    if (side === "front") setFrontFile(null);
    else setBackFile(null);
    setValue(side, undefined as any, { shouldValidate: true });
  };

  return (
    <div className="py-8 px-6 max-w-lg">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-xl font-semibold text-foreground tracking-tight">
          Prepare your document
        </h1>
        <p className="text-sm text-muted-foreground mt-1 leading-relaxed">
          You'll need to scan both sides of your ID. Make sure you capture a
          clear and complete image.
        </p>
      </div>

      {/* Document summary */}
      <div className="bg-muted rounded-lg px-4 py-3 mb-8 flex flex-col gap-1.5">
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">Country of document</span>
          <span className="font-medium text-foreground">{country || "—"}</span>
        </div>
        <Separator />
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">Document type</span>
          <span className="font-medium text-foreground">
            {selectedIdType?.label || idType?.replace(/_/g, " ") || "—"}
          </span>
        </div>
      </div>

      {cameraError && (
        <div className="mb-6 text-xs text-destructive bg-destructive/10 px-4 py-3 rounded-lg">
          {cameraError}
        </div>
      )}

      <form onSubmit={handleSubmit(onSubmit)}>
        <div className="flex flex-col gap-6 mb-8">
          {(["front", "back"] as const).map((side) => {
            const file = side === "front" ? frontFile : backFile;
            const inputRef = side === "front" ? frontInputRef : backInputRef;
            const videoRef = side === "front" ? frontVideoRef : backVideoRef;
            const isCamActive = activeCam === side;
            const existingDoc = side === "front" ? existingFront : existingBack;

            return (
              <div key={side}>
                <label className="text-sm font-medium text-foreground mb-2 block capitalize">
                  {side} side
                </label>
                <input
                  ref={inputRef}
                  type="file"
                  accept="image/*"
                  className="hidden"
                  onChange={(e) => handleFileChange(e, side)}
                />

                {isCamActive ? (
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
                        onClick={() => capturePhoto(side)}
                        className="flex-1 py-2 text-sm font-medium bg-primary text-primary-foreground rounded-md"
                      >
                        Capture
                      </button>
                      <button
                        type="button"
                        onClick={() => setActiveCam(null)}
                        className="px-4 py-2 text-sm font-medium bg-background border border-border rounded-md"
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                ) : file ? (
                  <div className="relative rounded-lg overflow-hidden border border-border">
                    <img
                      src={file.preview}
                      alt={`${side} side`}
                      className="w-full aspect-video object-cover"
                    />
                    <div className="absolute inset-0 bg-black/30 flex items-center justify-center opacity-0 hover:opacity-100 transition-opacity">
                      <button
                        type="button"
                        onClick={() => clearFile(side)}
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
                ) : existingDoc ? (
                  <ExistingDocumentPreview
                    docId={existingDoc.id}
                    side={side}
                    onClear={() => {
                      // let them replace it — opens file picker
                      if (side === "front") setFrontFile(null);
                      else setBackFile(null);
                      inputRef.current?.click();
                    }}
                  />
                ) : (
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
                        onClick={() => inputRef.current?.click()}
                        className="flex items-center gap-2 px-4 py-2 text-sm font-medium border border-border rounded-md bg-background hover:bg-muted transition-colors"
                      >
                        <Upload className="w-4 h-4" />
                        Upload file
                      </button>
                      <button
                        type="button"
                        onClick={() => openCamera(side)}
                        className="flex items-center gap-2 px-4 py-2 text-sm font-medium border border-border rounded-md bg-background hover:bg-muted transition-colors"
                      >
                        <Camera className="w-4 h-4" />
                        Use camera
                      </button>
                    </div>
                  </div>
                )}
                {errors[side] && (
                  <p className="text-xs text-destructive mt-2">
                    {errors[side]?.message}
                  </p>
                )}
              </div>
            );
          })}
        </div>

        <FormNavigation
          onNext={handleSubmit(onSubmit)}
          isSubmitting={uploadDocument.isPending}
        />
      </form>
    </div>
  );
}
