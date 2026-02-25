'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { supabase } from '@/lib/supabase';

export default function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const router = useRouter();

  const handleLogin = async () => {
    setLoading(true);
    setError('');

    try {
      const { data, error } = await supabase.auth.signInWithPassword({
        email,
        password,
      });

      if (error) {
        setError(error.message);
        setLoading(false);
        return;
      }

      if (data.session) {
        localStorage.setItem('authToken', data.session.access_token);
        console.log('Login successful');
        router.push('/editor');
      }
    } catch (err: any) {
      setError(err.message || 'Login failed');
      setLoading(false);
    }
  };

  const handleSignUp = async () => {
    setLoading(true);
    setError('');

    try {
      const { data, error } = await supabase.auth.signUp({
        email,
        password,
      });

      if (error) {
        setError(error.message);
        setLoading(false);
        return;
      }

      if (data.session) {
        localStorage.setItem('authToken', data.session.access_token);
        console.log('Sign up successful');
        router.push('/editor');
      } else {
        setError('Please check your email to confirm your account');
        setLoading(false);
      }
    } catch (err: any) {
      setError(err.message || 'Sign up failed');
      setLoading(false);
    }
  };

  return (
    <div style={{ padding: '20px', maxWidth: '400px', margin: '0 auto' }}>
      <h2>Login</h2>
      {error && (
        <div style={{ padding: '10px', marginBottom: '10px', backgroundColor: '#fee', color: 'red', borderRadius: '5px' }}>
          {error}
        </div>
      )}
      <input
        type="email"
        placeholder="Email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        disabled={loading}
        style={{
          display: 'block',
          width: '100%',
          marginBottom: '10px',
          padding: '8px',
          boxSizing: 'border-box',
        }}
      />
      <input
        type="password"
        placeholder="Password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        disabled={loading}
        style={{
          display: 'block',
          width: '100%',
          marginBottom: '10px',
          padding: '8px',
          boxSizing: 'border-box',
        }}
      />
      <button
        onClick={handleLogin}
        disabled={loading || !email || !password}
        style={{
          marginRight: '10px',
          padding: '8px 16px',
          backgroundColor: loading ? '#ccc' : '#0070f3',
          color: 'white',
          border: 'none',
          borderRadius: '5px',
          cursor: loading ? 'not-allowed' : 'pointer',
        }}
      >
        {loading ? 'Loading...' : 'Login'}
      </button>
      <button
        onClick={handleSignUp}
        disabled={loading || !email || !password}
        style={{
          padding: '8px 16px',
          backgroundColor: loading ? '#ccc' : '#666',
          color: 'white',
          border: 'none',
          borderRadius: '5px',
          cursor: loading ? 'not-allowed' : 'pointer',
        }}
      >
        {loading ? 'Loading...' : 'Sign Up'}
      </button>
    </div>
  );
}
