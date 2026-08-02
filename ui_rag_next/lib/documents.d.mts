import type { IngestDocumentResponse } from "./contracts";

export type IngestResult = {
  tone: "success" | "warning" | "error";
  title: string;
  message: string;
  citationCount: number;
};

export function validateDocumentPath(path: string): string | null;
export function createIngestResult(response: IngestDocumentResponse): IngestResult;
export function describeDocumentRequestError(error: unknown): string;
