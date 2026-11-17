import defaultTheme from 'tailwindcss/defaultTheme';

/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{svelte,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      // Colour flags first (unicode-range in style.css), then the defaults.
      fontFamily: {
        sans: ['"Twemoji Flags"', ...defaultTheme.fontFamily.sans],
        mono: ['"Twemoji Flags"', ...defaultTheme.fontFamily.mono],
      },
    },
  },
  plugins: [
    require('@tailwindcss/typography'),
  ],
}