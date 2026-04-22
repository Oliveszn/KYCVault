export const statusConfig: Record<
  string,
  { label: string; dot: string; text: string; bg: string }
> = {
  approved: {
    label: "Approved",
    dot: "bg-emerald-500",
    text: "text-emerald-600",
    bg: "bg-emerald-50",
  },
  rejected: {
    label: "Rejected",
    dot: "bg-red-500",
    text: "text-red-600",
    bg: "bg-red-50",
  },
  in_review: {
    label: "In Review",
    dot: "bg-amber-500",
    text: "text-amber-600",
    bg: "bg-amber-50",
  },
  initiated: {
    label: "Initiated",
    dot: "bg-blue-500",
    text: "text-blue-600",
    bg: "bg-blue-50",
  },
  doc_upload: {
    label: "Doc Upload",
    dot: "bg-blue-500",
    text: "text-blue-600",
    bg: "bg-blue-50",
  },
  face_verify: {
    label: "Face Verify",
    dot: "bg-violet-500",
    text: "text-violet-600",
    bg: "bg-violet-50",
  },
};
