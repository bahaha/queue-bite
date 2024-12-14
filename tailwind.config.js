/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./internal/layouts/*.{html,templ}",
        "./internal/features/**/views/*.{html,templ}",
    ],
    theme: {
        extend: {},
    },
    plugins: [],
}

