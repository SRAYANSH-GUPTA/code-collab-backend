import { useState, useEffect, useCallback, useRef } from 'react';
import { LintError, AnalyzeResponse } from '@/lib/types';

interface UseWebSocketReturn {
  isConnected: boolean;
  errors: LintError[];
  sendMessage: (message: object) => void;
  executionTime: number | null;
}

export function useWebSocket(url: string, token: string): UseWebSocketReturn {
  const [isConnected, setIsConnected] = useState(false);
  const [errors, setErrors] = useState<LintError[]>([]);
  const [executionTime, setExecutionTime] = useState<number | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!url || !token) {
      console.log('Missing URL or token, skipping WebSocket connection');
      return;
    }

    
    const wsUrl = `${url}/ws?token=${token}`;
    console.log('Connecting to WebSocket:', wsUrl);

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log('WebSocket connected');
      setIsConnected(true);
    };

    ws.onclose = () => {
      console.log('WebSocket disconnected');
      setIsConnected(false);
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      setIsConnected(false);
    };

    ws.onmessage = (event) => {
      try {
        const response: AnalyzeResponse = JSON.parse(event.data);
        console.log('Received message:', response);

        if (response.type === 'analysis_result') {
          setErrors(response.errors || []);
          setExecutionTime(response.executionTime || null);
        } else if (response.type === 'error') {
          console.error('Server error:', response.message);
          setErrors([]);
        }
      } catch (error) {
        console.error('Failed to parse message:', error);
      }
    };

    
    return () => {
      console.log('Closing WebSocket connection');
      ws.close();
    };
  }, [url, token]);

  const sendMessage = useCallback((message: object) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      console.log('Sending message:', message);
      wsRef.current.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected');
    }
  }, []);

  return {
    isConnected,
    errors,
    sendMessage,
    executionTime,
  };
}
