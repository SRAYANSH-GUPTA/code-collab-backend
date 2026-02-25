'use client';

import { useEffect, useRef, useImperativeHandle, forwardRef } from 'react';
import Editor, { Monaco } from '@monaco-editor/react';
import { LintError } from '@/lib/types';

interface CodeEditorProps {
  language: string;
  value: string;
  onChange: (value: string) => void;
  errors: LintError[];
}

export interface CodeEditorRef {
  goToLine: (line: number) => void;
}

const CodeEditor = forwardRef<CodeEditorRef, CodeEditorProps>(
  ({ language, value, onChange, errors }, ref) => {
    const editorRef = useRef<any>(null);
    const monacoRef = useRef<Monaco | null>(null);

    const handleEditorDidMount = (editor: any, monaco: Monaco) => {
      editorRef.current = editor;
      monacoRef.current = monaco;
    };

    useImperativeHandle(ref, () => ({
      goToLine: (line: number) => {
        if (!editorRef.current) return;

        editorRef.current.revealLineInCenter(line);
        editorRef.current.setPosition({ lineNumber: line, column: 1 });
        editorRef.current.focus();
      }
    }));

    useEffect(() => {
      if (!editorRef.current || !monacoRef.current) return;

      const monaco = monacoRef.current;
      const model = editorRef.current.getModel();
      if (!model) return;

      const markers = errors.map((error) => ({
        startLineNumber: error.line,
        startColumn: error.column,
        endLineNumber: error.line,
        endColumn: error.column + error.length,
        message: error.message,
        severity: error.severity === 'error'
          ? monaco.MarkerSeverity.Error
          : error.severity === 'warning'
          ? monaco.MarkerSeverity.Warning
          : monaco.MarkerSeverity.Info,
      }));

      monaco.editor.setModelMarkers(model, 'linter', markers);
    }, [errors]);

    const getMonacoLanguage = (lang: string) => {
      const languageMap: { [key: string]: string } = {
        typescript: 'typescript',
        python: 'python',
        dart: 'dart',
        go: 'go',
        cpp: 'cpp',
      };
      return languageMap[lang] || 'plaintext';
    };

    return (
      <Editor
        height="500px"
        language={getMonacoLanguage(language)}
        value={value}
        onChange={(value) => onChange(value || '')}
        onMount={handleEditorDidMount}
        theme="vs-dark"
        options={{
          minimap: { enabled: true },
          fontSize: 14,
          lineNumbers: 'on',
          automaticLayout: true,
          scrollBeyondLastLine: false,
          wordWrap: 'on',
        }}
      />
    );
  }
);

CodeEditor.displayName = 'CodeEditor';

export default CodeEditor;
