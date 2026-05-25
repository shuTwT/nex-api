"use client";

import { useEffect, useRef } from "react";
import type { editor } from "monaco-editor";
import Editor from "@monaco-editor/react";

interface MonacoEditorProps {
  value: string;
  onChange: (value: string) => void;
  language?: string;
  height?: string;
  placeholder?: string;
  disabled?: boolean;
}

export function MonacoEditor({
  value,
  onChange,
  language = "javascript",
  height = "300px",
  placeholder = "",
  disabled = false,
}: MonacoEditorProps) {
  const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null);

  useEffect(() => {
    if (editorRef.current && disabled) {
      editorRef.current.updateOptions({ readOnly: true });
    } else if (editorRef.current && !disabled) {
      editorRef.current.updateOptions({ readOnly: false });
    }
  }, [disabled]);

  return (
    <div className="border border-slate-200 rounded-md overflow-hidden">
      <Editor
        height={height}
        defaultLanguage={language}
        value={value || placeholder}
        onChange={(value) => onChange(value || "")}
        theme="vs-light"
        options={{
          minimap: { enabled: false },
          fontSize: 14,
          lineNumbers: "on",
          scrollBeyondLastLine: false,
          automaticLayout: true,
          tabSize: 2,
          wordWrap: "on",
          readOnly: disabled,
        }}
        onMount={(editor) => {
          editorRef.current = editor;
        }}
      />
    </div>
  );
}
