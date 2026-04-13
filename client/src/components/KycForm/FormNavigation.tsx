import { ChevronLeft, ChevronRight } from "lucide-react";
import { useAppSelector, useAppDispatch } from "@/store/hooks";
import { useLocation, useNavigate } from "react-router-dom";

interface FormNavigationProps {
  onNext?: () => void;
  onPrevious?: () => void;
  isNextDisabled?: boolean;
  isLastStep?: boolean;
  isSubmitting?: boolean;
}

const steps = ["initiate-session", "upload-docs", "face-verify"];
export default function FormNavigation({
  onNext,
  onPrevious,
  isNextDisabled = false,
  isLastStep = false,
  isSubmitting = false,
}: FormNavigationProps) {
  const dispatch = useAppDispatch();
  const currentStep = useAppSelector((state) => state.createKyc.currentStep);
  const navigate = useNavigate();

  const { pathname } = useLocation();
  const segment = pathname.split("/").pop() ?? "";
  const currentIndex = steps.indexOf(segment);

  const handleNext = () => {
    if (onNext) return onNext();
    navigate(`/verify/${steps[currentIndex + 1]}`);
  };

  const handlePrevious = () => {
    if (onPrevious) return onPrevious();
    navigate(`/verify/${steps[currentIndex - 1]}`);
  };

  const isFirst = currentIndex === 0;
  const isLast = currentIndex === steps.length - 1;
  return (
    <div className="flex items-center justify-between pt-6 border-t border-gray-200">
      <button
        type="button"
        onClick={handlePrevious}
        disabled={isFirst || isSubmitting}
        className={`flex items-center gap-2 px-6 py-2.5 rounded-lg font-medium transition ${
          isFirst || isSubmitting
            ? "text-gray-400 cursor-not-allowed"
            : "text-gray-700 hover:bg-gray-100"
        }`}
      >
        <ChevronLeft className="w-5 h-5" />
        Previous
      </button>

      <button
        type="button"
        onClick={handleNext}
        disabled={isNextDisabled || isSubmitting}
        className={`flex items-center gap-2 px-6 py-2.5 rounded-lg font-medium transition ${
          isNextDisabled || isSubmitting
            ? "bg-gray-300 text-gray-500 cursor-not-allowed"
            : "bg-primary text-primary-foreground hover:bg-primary/90 cursor-pointer"
        }`}
      >
        {isSubmitting ? (
          <>
            <span>Submitting...</span>
            <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
          </>
        ) : (
          <>
            <span>{isLast ? "Submit" : "Continue"}</span>
            {!isLast && <ChevronRight className="w-5 h-5" />}
          </>
        )}
      </button>
    </div>
  );
}
