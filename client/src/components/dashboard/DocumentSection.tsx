import DocumentPreview from "./AdminDocumentPreview";
type Doc = {
  id: string;
  side: "front" | "back";
};

type Props = {
  frontDoc?: Doc;
  backDoc?: Doc;
};
export default function DocumentsSection({ frontDoc, backDoc }: Props) {
  return (
    <div>
      <h2 className="text-sm font-semibold mb-4">Identity Documents</h2>

      <div className="grid grid-cols-2 gap-4">
        {frontDoc && <DocumentPreview docId={frontDoc.id} label="Front side" />}
        {backDoc && <DocumentPreview docId={backDoc.id} label="Back side" />}

        {!frontDoc && !backDoc && (
          <p className="text-sm text-muted-foreground col-span-2">
            No documents uploaded.
          </p>
        )}
      </div>
    </div>
  );
}
