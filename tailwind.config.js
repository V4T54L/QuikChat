/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './web/templates/**/*.html',
    './web/static/js/**/*.js',
  ],
  theme: {
    extend: {
        colors: {
            'primary': '#1e293b',    // slate-800
            'secondary': '#334155', // slate-700
            'accent': '#475569',    // slate-600
            'text-main': '#f8fafc', // slate-50
            'text-dim': '#94a3b8',  // slate-400
            'highlight': '#2dd4bf', // teal-400
        }
    },
  },
  plugins: [],
}

