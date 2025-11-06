import { Language } from '@/lib/types';

interface LanguageSelectorProps {
  value: Language;
  onChange: (language: Language) => void;
}

export default function LanguageSelector({ value, onChange }: LanguageSelectorProps) {
  return (
    <select
      value={value}
      onChange={(e) => onChange(e.target.value as Language)}
      style={{ padding: '8px', fontSize: '14px' }}
    >
      <option value="typescript">TypeScript</option>
      <option value="python">Python</option>
      <option value="dart">Dart</option>
      <option value="go">Go</option>
      <option value="cpp">C++</option>
    </select>
  );
}
