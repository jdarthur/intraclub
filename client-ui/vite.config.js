import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [react()],
    server: {
        proxy: {
            "/api": "http://localhost:8080",
            "/api/ws": {
                target: "http://localhost:8080",
                ws: true
            }
        },
        host: "0.0.0.0"
    }
})
