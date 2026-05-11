/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        safe: "#16a34a",
        warning: "#ca8a04",
        critical: "#dc2626",
        slate: "#0f172a",
      },
    },
  },
  plugins: [],
};
