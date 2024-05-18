/** @types {import('tailwindcss').Config} */
module.exports = {
    content: ["./**/*.html", "./**/*.templ", "./**/*.go",],
    safelist: [],
    plugins: [require("daisyui"), require("@tailwindcss/typography")],
    daisyui: {
        themes: [
            "acid"
        ]
    },
    theme: {
        extend: {
            animation: {
                "fade-out-delay": 'fadeOut 3000ms ease-in forwards',
                'gradient-x': 'gradient 5s ease infinite',
            },
            keyframes: theme => ({
                gradient: {
                    '0%, 100%': {
                        backgroundPosition: '0% 50%'
                    },
                    '50%': {
                        backgroundPosition: '100% 50%'
                    }
                },
                fadeOut: {
                    '0%': { opacity: 1 },
                    '75%': { opacity: 1 },
                    '95%': { opacity: 0 },
                    '100%': { opacity: 0 },
                },
            }),
        }
    }
}