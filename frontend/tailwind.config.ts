import type { Config } from 'tailwindcss';
import animate from 'tailwindcss-animate';

// Enterprise dark-first design system. Colors are exposed as CSS variables
// (see src/index.css) so we can theme without recompiling Tailwind.
export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    container: {
      center: true,
      padding: '2rem',
      screens: { '2xl': '1400px' },
    },
    extend: {
      colors: {
        border: 'hsl(var(--border))',
        input: 'hsl(var(--input))',
        ring: 'hsl(var(--ring))',
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        primary: {
          DEFAULT: 'hsl(var(--primary))',
          foreground: 'hsl(var(--primary-foreground))',
        },
        secondary: {
          DEFAULT: 'hsl(var(--secondary))',
          foreground: 'hsl(var(--secondary-foreground))',
        },
        muted: {
          DEFAULT: 'hsl(var(--muted))',
          foreground: 'hsl(var(--muted-foreground))',
        },
        accent: {
          DEFAULT: 'hsl(var(--accent))',
          foreground: 'hsl(var(--accent-foreground))',
        },
        destructive: {
          DEFAULT: 'hsl(var(--destructive))',
          foreground: 'hsl(var(--destructive-foreground))',
        },
        success: {
          DEFAULT: 'hsl(var(--success))',
          foreground: 'hsl(var(--success-foreground))',
        },
        warning: {
          DEFAULT: 'hsl(var(--warning))',
          foreground: 'hsl(var(--warning-foreground))',
        },
        card: {
          DEFAULT: 'hsl(var(--card))',
          foreground: 'hsl(var(--card-foreground))',
        },
        elevated: {
          DEFAULT: 'hsl(var(--elevated))',
          foreground: 'hsl(var(--elevated-foreground))',
        },
        violet: 'hsl(var(--violet))',
      },
      borderRadius: {
        xl: 'calc(var(--radius) + 4px)',
        lg: 'var(--radius)',
        md: 'calc(var(--radius) - 2px)',
        sm: 'calc(var(--radius) - 4px)',
      },
      boxShadow: {
        // Beveled console panel: hairline top highlight + soft drop.
        panel:
          'inset 0 1px 0 0 hsl(210 40% 100% / 0.04), 0 1px 2px 0 hsl(222 40% 2% / 0.4)',
        // Raised surface (popovers, dialogs, command palette).
        raised:
          'inset 0 1px 0 0 hsl(210 40% 100% / 0.05), 0 12px 32px -12px hsl(222 60% 2% / 0.7)',
        // Signal glow for live/primary emphasis.
        glow: '0 0 0 1px hsl(var(--primary) / 0.25), 0 0 24px -6px hsl(var(--primary) / 0.45)',
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'ui-monospace', 'monospace'],
      },
      keyframes: {
        'pulse-ring': {
          '0%': { boxShadow: '0 0 0 0 hsl(var(--success) / 0.5)' },
          '70%': { boxShadow: '0 0 0 5px hsl(var(--success) / 0)' },
          '100%': { boxShadow: '0 0 0 0 hsl(var(--success) / 0)' },
        },
        // Traveling sheen for skeleton loaders.
        shimmer: {
          '100%': { transform: 'translateX(100%)' },
        },
      },
      animation: {
        'pulse-ring': 'pulse-ring 2s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        shimmer: 'shimmer 1.6s infinite',
      },
    },
  },
  plugins: [animate],
} satisfies Config;
