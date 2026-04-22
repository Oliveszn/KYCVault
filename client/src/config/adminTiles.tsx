import { Clock, CheckCircle, XCircle, Users } from "lucide-react";
export const tiles = [
  {
    label: "In Review",
    key: "in_review",
    icon: Clock,
    color: "text-yellow-500 bg-yellow-500/10",
  },
  {
    label: "Approved",
    key: "approved",
    icon: CheckCircle,
    color: "text-green-500 bg-green-500/10",
  },
  {
    label: "Rejected",
    key: "rejected",
    icon: XCircle,
    color: "text-red-500 bg-red-500/10",
  },
  {
    label: "Total Sessions",
    key: null,
    icon: Users,
    color: "text-blue-500 bg-blue-500/10",
  },
];
