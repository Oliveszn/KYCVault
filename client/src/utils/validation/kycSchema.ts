import { z } from "zod";

export const initiateSessionSchema = z.object({
  IDType: z.string().min(1, "Please select a document type"),
  country: z.string().min(1, "Please select a country"),
});

export const uploadDocumentSchema = z.object({
  front: z
    .instanceof(File, { message: "Front side is required" })
    .refine((f) => f.size > 0, "Front side is required"),
  back: z
    .instanceof(File, { message: "Back side is required" })
    .refine((f) => f.size > 0, "Back side is required"),
});

export const faceVerificationSchema = z.object({
  selfie: z
    .instanceof(File, { message: "Please take a selfie to continue" })
    .refine((f) => f.size > 0, "Please take a selfie to continue"),
});

export type InitiateSessionValues = z.infer<typeof initiateSessionSchema>;
export type UploadDocumentValues = z.infer<typeof uploadDocumentSchema>;
export type FaceVerificationValues = z.infer<typeof faceVerificationSchema>;
