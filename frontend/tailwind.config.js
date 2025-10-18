/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{js,ts,jsx,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                background: {
                    primary: '#BFC0C0',
                    secondary: '#4F5D75',
                    surface: '#FFFFFF',
                    'surface-alt': '#F3F4F6',
                    nav: '#2D3142',
                    'table-header': '#4F5D75',
                },
                text: {
                    primary: '#0F172A',
                    secondary: '#64748B',
                    'on-nav': '#FFFFFF',
                    'on-accent': '#FFFFFF',
                },
                accent: {
                    primary: '#4F46E5',
                    'primary-hover': '#6366F1',
                },
            },
        },
    },
    plugins: [],
}
