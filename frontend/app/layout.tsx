import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Code Linting Platform',
  description: 'Real-time code linting for multiple languages',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body style={{ margin: 0, fontFamily: 'Arial, sans-serif' }}>
        {children}
      </body>
    </html>
  )
}
