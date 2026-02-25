interface ConnectionStatusProps {
  isConnected: boolean;
}

export default function ConnectionStatus({ isConnected }: ConnectionStatusProps) {
  return (
    <div style={{ color: isConnected ? 'green' : 'red', padding: '10px' }}>
      {isConnected ? '● Connected' : '○ Disconnected'}
    </div>
  );
}
