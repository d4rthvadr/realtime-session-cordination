/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        surface: "#f7f9fb",
        "surface-container-low": "#f2f4f6",
        "surface-container": "#eceef0",
        "surface-container-high": "#e6e8ea",
        "on-surface": "#191c1e",
        "on-surface-variant": "#45464d",
        outline: "#76777d",
        "outline-variant": "#c6c6cd",
        primary: "#000000",
        "on-primary": "#ffffff",
        secondary: "#4b41e1",
        "on-secondary": "#ffffff",
        "secondary-container": "#645efb",
        safe: "#16a34a",
        warning: "#ca8a04",
        critical: "#dc2626",
      },
      fontFamily: {
        display: ["Geist", "sans-serif"],
        headline: ["Geist", "sans-serif"],
        body: ["Inter", "sans-serif"],
      },
      fontSize: {
        "display-lg": [
          "48px",
          { lineHeight: "56px", letterSpacing: "-0.02em", fontWeight: "700" },
        ],
        "headline-lg": [
          "32px",
          { lineHeight: "40px", letterSpacing: "-0.01em", fontWeight: "600" },
        ],
        "headline-md": ["24px", { lineHeight: "32px", fontWeight: "600" }],
        "body-lg": ["18px", { lineHeight: "28px", fontWeight: "400" }],
        "body-md": ["16px", { lineHeight: "24px", fontWeight: "400" }],
        "body-sm": ["14px", { lineHeight: "20px", fontWeight: "400" }],
        "label-md": ["14px", { lineHeight: "20px", fontWeight: "600" }],
        "timer-display": [
          "64px",
          { lineHeight: "64px", letterSpacing: "-0.04em", fontWeight: "700" },
        ],
      },
      maxWidth: {
        "container-max": "1280px",
      },
    },
  },
  plugins: [],
};
