export interface LintError {
  line: number;
  column: number;
  message: string;
  severity: 'error' | 'warning' | 'info';
  length: number;
}

export interface AnalyzeRequest {
  action: 'analyze';
  language: string;
  code: string;
}

export interface AnalyzeResponse {
  type: 'analysis_result' | 'error';
  errors?: LintError[];
  message?: string;
  executionTime?: number;
}

export type Language = 'typescript' | 'python' | 'dart' | 'go' | 'cpp';
