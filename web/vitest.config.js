import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    include: ['js/__tests__/**/*.test.js'],
    exclude: ['e2e/**', 'dist/**', 'node_modules/**'],
  },
});
