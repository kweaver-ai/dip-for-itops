  /* eslint-disable */
  import { createConfig, configAlias } from 'noya/test';

  export default async () => {
    return await configAlias({
      ...createConfig(),
      collectCoverageFrom: [
        'src/**/*.{ts,js,tsx,jsx}',
        '!src/.noya/**',
        '!src/.noya-test/**',
        '!src/.noya-production/**'
      ]
    });
  };
