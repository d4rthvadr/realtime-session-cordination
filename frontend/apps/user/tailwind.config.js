/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        // Stripe-inspired color palette
        surface: "#ffffff",
        "surface-secondary": "#f6f9fc",
        "surface-tertiary": "#e3e8ef",
        primary: "#635bff",
        "primary-dark": "#0a2540",
        "primary-light": "#7a73ff",
        accent: "#00d4ff",
        "accent-cyan": "#00d4ff",
        "accent-purple": "#7c66ff",
        "text-primary": "#0a2540",
        "text-secondary": "#425466",
        "text-tertiary": "#8898aa",
        border: "#e6ebf1",
        "border-light": "#f0f4f8",
        safe: "#00d924",
        warning: "#ffb800",
        critical: "#ff5263",
        "gradient-start": "#635bff",
        "gradient-end": "#7c66ff",
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
