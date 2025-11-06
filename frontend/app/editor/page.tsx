'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { useRouter } from 'next/navigation';
import dynamic from 'next/dynamic';
import ConnectionStatus from '@/components/ConnectionStatus';
import LanguageSelector from '@/components/LanguageSelector';
import ErrorPanel from '@/components/ErrorPanel';
import { useWebSocket } from '@/hooks/useWebSocket';
import { Language } from '@/lib/types';
import type { CodeEditorRef } from '@/components/CodeEditor';

const CodeEditor = dynamic(() => import('@/components/CodeEditor'), {
  ssr: false,
  loading: () => <div style={{ padding: '20px' }}>Loading editor...</div>,
});

const DEFAULT_CODE: { [key in Language]: string } = {
  typescript: `const x: number = 'hello';\nconst y = 42;\nconsole.log(x + y);`,
  python: `def greet(name)\n    print(f"Hello, {name}")`,
  go: `package main\n\nfunc main {\n    fmt.Println("Hello")\n}`,
  dart: `void main() {\n  var x: int = "hello";\n  print(x);\n}`,
  cpp: `#include <iostream>\n\nint main {\n    std::cout << "Hello";\n    return 0;\n}`,
};

export default function Editor() {
  const router = useRouter();
  const [token, setToken] = useState('');
  const [language, setLanguage] = useState<Language>('typescript');
  const [code, setCode] = useState(DEFAULT_CODE.typescript);
  const [wsUrl, setWsUrl] = useState('');
  const editorRef = useRef<CodeEditorRef>(null);

  useEffect(() => {
    const storedToken = localStorage.getItem('authToken');
    if (!storedToken) {
      console.log('No auth token found, redirecting to login');
      router.push('/login');
      return;
    }

    setToken(storedToken);
    const url = process.env.NEXT_PUBLIC_WS_URL || 'ws:
    setWsUrl(url);
    console.log('Using WebSocket URL:', url);
  }, [router]);

  const { isConnected, errors, sendMessage, executionTime } = useWebSocket(wsUrl, token);

  useEffect(() => {
    if (!isConnected || !code) return;

    const timer = setTimeout(() => {
      console.log('Sending analysis request for', language);
      sendMessage({
        action: 'analyze',
        language,
        code,
      });
    }, 300);

    return () => clearTimeout(timer);
  }, [code, language, isConnected, sendMessage]);

  const handleLanguageChange = useCallback((newLanguage: Language) => {
    setLanguage(newLanguage);
    setCode(DEFAULT_CODE[newLanguage]);
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('authToken');
    router.push('/');
  };

  const handleErrorClick = (line: number) => {
    editorRef.current?.goToLine(line);
  };

  if (!token) {
    return <div style={{ padding: '20px' }}>Loading...</div>;
  }

  return (
    <div style={{ padding: '20px', maxWidth: '1200px', margin: '0 auto' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '10px' }}>
        <h1 style={{ margin: 0 }}>Code Linter</h1>
        <button
          onClick={handleLogout}
          style={{
            padding: '8px 16px',
            backgroundColor: '#dc3545',
            color: 'white',
            border: 'none',
            borderRadius: '5px',
            cursor: 'pointer',
          }}
        >
          Logout
        </button>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', marginBottom: '10px', gap: '20px' }}>
        <div>
          <label style={{ marginRight: '10px' }}>Language:</label>
          <LanguageSelector value={language} onChange={handleLanguageChange} />
        </div>
        <ConnectionStatus isConnected={isConnected} />
      </div>

      <div style={{ border: '1px solid #ccc', marginBottom: '10px' }}>
        <CodeEditor
          ref={editorRef}
          language={language}
          value={code}
          onChange={setCode}
          errors={errors}
        />
      </div>

      <ErrorPanel
        errors={errors}
        executionTime={executionTime}
        onErrorClick={handleErrorClick}
      />

      <div style={{ marginTop: '20px', padding: '10px', backgroundColor: '#f5f5f5', borderRadius: '5px' }}>
        <h4 style={{ margin: '0 0 10px 0' }}>Instructions:</h4>
        <ul style={{ margin: 0, paddingLeft: '20px' }}>
          <li>Type code in the editor above</li>
          <li>Errors will appear as red squiggles in the editor</li>
          <li>Full error details are shown in the PROBLEMS panel below</li>
          <li>Click on any error in the PROBLEMS panel to jump to that line</li>
          <li>Switch languages using the dropdown</li>
          <li>Real TypeScript linting is active!</li>
        </ul>
      </div>
    </div>
  );
}
