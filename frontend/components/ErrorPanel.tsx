'use client';

import { LintError } from '@/lib/types';

interface ErrorPanelProps {
  errors: LintError[];
  executionTime: number | null;
  onErrorClick?: (line: number) => void;
}

export default function ErrorPanel({ errors, executionTime, onErrorClick }: ErrorPanelProps) {
  const errorCount = errors.filter(e => e.severity === 'error').length;
  const warningCount = errors.filter(e => e.severity === 'warning').length;
  const infoCount = errors.filter(e => e.severity === 'info').length;

  return (
    <div style={{
      border: '1px solid #ddd',
      borderRadius: '4px',
      marginTop: '10px',
      backgroundColor: '#1e1e1e',
      color: '#d4d4d4',
      fontFamily: 'Consolas, Monaco, monospace',
      fontSize: '13px'
    }}>
      <div style={{
        padding: '8px 12px',
        borderBottom: '1px solid #2d2d2d',
        backgroundColor: '#252526',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center'
      }}>
        <div style={{ display: 'flex', gap: '15px', alignItems: 'center' }}>
          <span style={{ fontWeight: 600 }}>PROBLEMS</span>
          {errorCount > 0 && (
            <span style={{ display: 'flex', alignItems: 'center', gap: '5px' }}>
              <span style={{ color: '#f48771' }}>●</span>
              <span>{errorCount}</span>
            </span>
          )}
          {warningCount > 0 && (
            <span style={{ display: 'flex', alignItems: 'center', gap: '5px' }}>
              <span style={{ color: '#cca700' }}>●</span>
              <span>{warningCount}</span>
            </span>
          )}
          {infoCount > 0 && (
            <span style={{ display: 'flex', alignItems: 'center', gap: '5px' }}>
              <span style={{ color: '#75beff' }}>●</span>
              <span>{infoCount}</span>
            </span>
          )}
        </div>
        {executionTime !== null && (
          <span style={{ fontSize: '11px', color: '#858585' }}>
            {executionTime}ms
          </span>
        )}
      </div>

      <div style={{
        maxHeight: '200px',
        overflowY: 'auto'
      }}>
        {errors.length === 0 ? (
          <div style={{
            padding: '20px',
            textAlign: 'center',
            color: '#858585'
          }}>
            No problems detected
          </div>
        ) : (
          <table style={{
            width: '100%',
            borderCollapse: 'collapse'
          }}>
            <tbody>
              {errors.map((err, i) => {
                const severityIcon = err.severity === 'error' ? '✖' :
                                    err.severity === 'warning' ? '⚠' : 'ℹ';
                const severityColor = err.severity === 'error' ? '#f48771' :
                                     err.severity === 'warning' ? '#cca700' : '#75beff';

                return (
                  <tr
                    key={i}
                    onClick={() => onErrorClick?.(err.line)}
                    style={{
                      cursor: onErrorClick ? 'pointer' : 'default',
                      borderBottom: '1px solid #2d2d2d',
                      transition: 'background-color 0.1s'
                    }}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.backgroundColor = '#2a2d2e';
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.backgroundColor = 'transparent';
                    }}
                  >
                    <td style={{
                      padding: '6px 12px',
                      width: '20px',
                      textAlign: 'center',
                      color: severityColor
                    }}>
                      {severityIcon}
                    </td>
                    <td style={{
                      padding: '6px 12px',
                      width: '100%'
                    }}>
                      <div style={{ marginBottom: '2px' }}>
                        {err.message}
                      </div>
                      <div style={{
                        fontSize: '11px',
                        color: '#858585'
                      }}>
                        Line {err.line}, Col {err.column}
                      </div>
                    </td>
                    <td style={{
                      padding: '6px 12px',
                      fontSize: '11px',
                      color: '#858585',
                      textAlign: 'right',
                      whiteSpace: 'nowrap'
                    }}>
                      {err.severity}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
