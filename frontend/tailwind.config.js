/** @type {import('tailwindcss').Config} */

// ImageForge design tokens — Apple/Linear/Raycast direction.
// The image is the primary visual element; chrome stays quiet.
const colors = {
  bg: '#F5F5F7',          // app background
  surface: '#FFFFFF',     // cards / panels
  'surface-2': '#FBFBFD', // subtle elevated surface
  text: '#1D1D1F',        // primary text
  'text-secondary': '#6E6E73', // secondary text
  border: '#E5E5E7',
  accent: '#0071E3',      // blue accent (no AI purple)
  'accent-hover': '#0058b0',
  danger: '#D70015',
  success: '#007d26',
  warning: '#b25000',
module.exports = {
  content: ['./index.html', './src/**/*.{vue,ts}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors,
      fontFamily: {
        sans: ['Geist', '-apple-system', 'BlinkMacSystemFont', 'SF Pro Text', 'Segoe UI', 'Arial', 'sans-serif'],
        mono: ['Geist Mono', 'SF Mono', 'JetBrains Mono', 'Menlo', 'monospace'],
      },
      fontSize: {
        xs: ['0.75rem', { lineHeight: '1.125rem', letterSpacing: '0' }],
        sm: ['0.875rem', { lineHeight: '1.25rem', letterSpacing: '-0.005em' }],
        base: ['0.9375rem', { lineHeight: '1.375rem', letterSpacing: '-0.01em' }],
        lg: ['1.125rem', { lineHeight: '1.5rem', letterSpacing: '-0.0125em' }],
        xl: ['1.5rem', { lineHeight: '1.75rem', letterSpacing: '-0.015625em' }],
        '2xl': ['2rem', { lineHeight: '2.25rem', letterSpacing: '-0.02em' }],
      },
      borderRadius: {
        sm: '6px',
        md: '10px',
        lg: '14px',
        xl: '20px',
      },
      boxShadow: {
        sm: '0 1px 2px rgba(0,0,0,0.04)',
        md: '0 2px 8px rgba(0,0,0,0.05)',
        lg: '0 8px 24px rgba(0,0,0,0.08)',
      },
      keyframes: {
        'fade-in': {
          '0%': { opacity: '0', transform: 'translateY(4px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        'scale-in': {
          '0%': { opacity: '0', transform: 'scale(0.97)' },
          '100%': { opacity: '1', transform: 'scale(1)' },
        },
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' },
        },
        'pulse-soft': {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.7' },
        },
        float: {
          '0%, 100%': { transform: 'translateY(0)' },
          '50%': { transform: 'translateY(-4px)' },
        },
      },
      animation: {
        'fade-in': 'fade-in 0.18s cubic-bezier(0.16, 1, 0.3, 1)',
        'scale-in': 'scale-in 0.15s cubic-bezier(0.16, 1, 0.3, 1)',
        shimmer: 'shimmer 2s linear infinite',
        'pulse-soft': 'pulse-soft 2s ease-in-out infinite',
        float: 'float 3s ease-in-out infinite',
      },
      zIndex: {
        dropdown: 50,
        popover: 60,
        tooltip: 70,
        'modal-backdrop': 80,
        modal: 90,
        overlay: 100,
      },
    },
  },
  plugins: [],
}
