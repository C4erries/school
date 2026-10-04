/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        canvas: "#F5F5F7",
        primary: {
          DEFAULT: "#4F46E5",
          hover: "#4338CA",
          light: "rgba(79, 70, 229, 0.12)",
        },
        mint: {
          DEFAULT: "#10B981",
          hover: "#059669",
          light: "rgba(16, 185, 129, 0.12)",
        },
        amber: {
          DEFAULT: "#F59E0B",
          hover: "#D97706",
          light: "rgba(245, 158, 11, 0.12)",
        },
        coral: {
          DEFAULT: "#F43F5E",
          hover: "#E11D48",
          light: "rgba(244, 63, 94, 0.12)",
        },
      },
      boxShadow: {
        "glass": "0 10px 30px -5px rgba(0, 0, 0, 0.05), 0 1px 3px 0 rgba(0, 0, 0, 0.02)",
        "glass-hover": "0 20px 40px -10px rgba(0, 0, 0, 0.08), 0 1px 4px 0 rgba(0, 0, 0, 0.03)",
        "glass-elevated": "0 25px 50px -12px rgba(0, 0, 0, 0.12)",
      },
      borderRadius: {
        "squircle-sm": "12px",
        "squircle": "16px",
        "squircle-lg": "24px",
      },
      backdropBlur: {
        "xs": "2px",
        "glass": "20px",
      },
    },
  },
  plugins: [],
};
