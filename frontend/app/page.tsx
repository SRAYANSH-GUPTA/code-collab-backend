'use client';

import Link from 'next/link';

export default function Home() {
  return (
    <div style={{ padding: '20px', maxWidth: '600px', margin: '0 auto' }}>
      <h1>Code Linter Platform</h1>
      <p>Real-time code linting for multiple languages</p>
      <p>Supported languages: TypeScript, Python, Go, Dart, C++</p>
      <div style={{ marginTop: '20px' }}>
        <Link
          href="/login"
          style={{ marginRight: '10px', padding: '10px 20px', backgroundColor: '#0070f3', color: 'white', textDecoration: 'none', borderRadius: '5px' }}
        >
          Login
        </Link>
        <Link
          href="/editor"
          style={{ padding: '10px 20px', backgroundColor: '#666', color: 'white', textDecoration: 'none', borderRadius: '5px' }}
        >
          Go to Editor
        </Link>
      </div>
    </div>
  );
}
