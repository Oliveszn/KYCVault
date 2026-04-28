import { useLocation } from "react-router-dom";

const stepMap: Record<string, number> = {
  new: 1,
  documents: 2,
  face: 3,
};

export default function ProgressBar() {
  const { pathname } = useLocation();
  const segment = pathname.split("/").pop() ?? "";
  const currentStep = stepMap[segment] ?? 1;
  const progress = (currentStep / 3) * 100;
  return (
    <div className="px-6 py-4">
      <div className="h-1.5 bg-muted rounded-full overflow-hidden">
        <div
          className="h-full bg-primary transition-all duration-300 ease-out"
          style={{ width: `${progress}%` }}
        />
      </div>
    </div>
  );
}
