/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{html,ts}",
  ],
  theme: {
    extend: {
      colors: {
        primary: "#0F172B",
        secondary: "#f0efeb",
        "gray-border": "#6c737f",
        selectedobj: "#009689",
        "default-black": "#191919",
        gray: "#666666",
        "gray-light": "#9d9d9d",
        "gray-light2": "#e5e7eb",
        "gray-light3": "#f3f3f3",
        "white-light": "#e0e6ed",
        "white-snow": "#fefefe",

        "yellow-light": "#fde68a",
        "yellow-dark": "#f59e0b",

        teal: "#14b8a6",
        red: "#ef4444",
        "red-crimson": "#dc2626",
        orange: "#f97316",
        emerald: "#10b981",
        purple: "#8b5cf6",

        gray: "#9ca3af",
        "white-light": "#f9fafb",
        "white-snow": "#ffffff",

        "red-bright": "#ff3b3b",
        "red-difference": "#b91c1c",
      },
    },
  },
  plugins: [],
}
