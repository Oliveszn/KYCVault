import { ChevronLeft, ChevronRight } from "lucide-react";
import { useAppSelector, useAppDispatch } from "@/store/hooks";
import { nextStep, previousStep } from "@/store/kyc-slice";

interface FormNavigationProps {
  onNext?: () => void;
  onPrevious?: () => void;
  isNextDisabled?: boolean;
  isLastStep?: boolean;
  isSubmitting?: boolean;
}
export default function FormNavigation({
  onNext,
  onPrevious,
  isNextDisabled = false,
  isLastStep = false,
  isSubmitting = false,
}: FormNavigationProps) {
  const dispatch = useAppDispatch();
  const currentStep = useAppSelector((state) => state.createKyc.currentStep);

  const handlePrevious = () => {
    if (onPrevious) {
      onPrevious();
    } else {
      dispatch(previousStep());
    }
  };

  const handleNext = () => {
    if (onNext) {
      onNext();
    } else {
      dispatch(nextStep());
    }
  };

  return (
    <div className="flex items-center justify-between pt-6 border-t border-gray-200">
      <button
        type="button"
        onClick={handlePrevious}
        disabled={currentStep === 1 || isSubmitting}
        className={`flex items-center gap-2 px-6 py-2.5 rounded-lg font-medium transition ${
          currentStep === 1 || isSubmitting
            ? "text-gray-400 cursor-not-allowed"
            : "text-gray-700 hover:bg-gray-100"
        }`}
      >
        <ChevronLeft className="w-5 h-5" />
        Previous
      </button>

      <button
        // type="submit"
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
            <span>{isLastStep ? "Submit" : "Continue"}</span>
            {!isLastStep && <ChevronRight className="w-5 h-5" />}
          </>
        )}
      </button>
    </div>
  );
}
