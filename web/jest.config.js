const nextJest = require('next/jest')

const createJestConfig = nextJest({
  dir: './',
})

const customJestConfig = {
  setupFiles: ['<rootDir>/jest.polyfills.js'],
  setupFilesAfterEnv: ['<rootDir>/jest.setup.ts'],
  testEnvironment: 'jest-environment-jsdom',
  testEnvironmentOptions: {
    customExportConditions: [''],
  },
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/$1',
  },
}

module.exports = async () => {
  const config = await createJestConfig(customJestConfig)();
  
  config.transformIgnorePatterns = [
    '/node_modules/(?!(msw|@mswjs|rettime|until-async|is-node-process|outvariant|strict-event-emitter|headers-polyfill|@open-draft|@bundled-es-modules)/)'
  ];
  return config;
}
