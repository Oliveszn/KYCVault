import { Sun, Eye } from "lucide-react";

//tips used in step 3
export const tips = [
  {
    icon: <Sun className="w-4 h-4" />,
    text: "Find an area with good lighting",
  },
  {
    icon: <Eye className="w-4 h-4" />,
    text: "Remove anything that covers your face",
  },
  {
    icon: (
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        className="w-4 h-4"
      >
        <circle cx="12" cy="12" r="10" />
        <line x1="4.93" y1="4.93" x2="19.07" y2="19.07" />
      </svg>
    ),
    text: "No glasses, they cause glare and reflections",
  },
];
