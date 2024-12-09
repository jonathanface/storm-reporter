import eslintPluginJest from 'eslint-plugin-jest';
import eslintPluginPrettier from 'eslint-plugin-prettier';
import eslintPluginTypeScript from '@typescript-eslint/eslint-plugin';
import parser from '@typescript-eslint/parser';

export default [
    {
        files: ['**/*.ts'], // Apply these rules to TypeScript files
        languageOptions: {
            parser: parser, // Use TypeScript parser
            ecmaVersion: 'latest',
            sourceType: 'module',
        },
        plugins: {
            '@typescript-eslint': eslintPluginTypeScript,
            jest: eslintPluginJest,
            prettier: eslintPluginPrettier,
        },
        rules: {
            '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
            '@typescript-eslint/no-explicit-any': 'warn',
            'prettier/prettier': ['warn', { singleQuote: true, trailingComma: 'all' }],
            'jest/no-disabled-tests': 'warn',
            'jest/no-focused-tests': 'error',
            'jest/no-identical-title': 'error',
            'jest/prefer-to-have-length': 'warn',
            'jest/valid-expect': 'error',
            'no-console': 'warn',
        },
    },
    {
        files: ['**/*.test.ts', '**/*.spec.ts'], // Specific settings for test files
        rules: {
            '@typescript-eslint/no-non-null-assertion': 'off',
        },
    },
];
