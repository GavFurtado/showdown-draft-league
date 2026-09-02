// @ts-check
const eslint = require('@eslint/js');
const { defineConfig } = require('eslint/config');
const tseslint = require('typescript-eslint');
const angular = require('angular-eslint');

module.exports = defineConfig([
  {
    files: ['**/*.ts'],
    extends: [
      eslint.configs.recommended,
      tseslint.configs.recommended,
      tseslint.configs.stylistic,
      angular.configs.tsRecommended,
    ],
    processor: angular.processInlineTemplates,
    rules: {
      '@angular-eslint/directive-selector': [
        'error',
        {
          type: 'attribute',
          prefix: 'app',
          style: 'camelCase',
        },
      ],
      '@angular-eslint/component-selector': [
        'error',
        {
          type: 'element',
          prefix: 'app',
          style: 'kebab-case',
        },
      ],
      '@typescript-eslint/naming-convention': [
        'error',
        { selector: 'default', format: ['camelCase'] },
        { selector: 'variable', format: ['camelCase', 'UPPER_CASE'] },
        { selector: 'typeLike', format: ['PascalCase'] },
        { selector: 'enumMember', format: ['camelCase', 'UPPER_CASE'] },
        {
          // server-sent fields and enum wire values are PascalCase/UPPER_CASE by design
          selector: ['classProperty', 'property'],
          format: ['camelCase', 'PascalCase', 'UPPER_CASE'],
        },
      ],
    },
  },
  {
    // object keys here mirror server field paths (e.g. 'Format.GroupCount') that
    // legitimately don't match any single casing convention
    files: ['src/app/features/create-league/create-league-validators.ts'],
    rules: {
      '@typescript-eslint/naming-convention': 'off',
    },
  },
  {
    files: ['**/*.html'],
    extends: [angular.configs.templateRecommended, angular.configs.templateAccessibility],
    rules: {
      // Taiga's tui-textfield associates <label tuiLabel> with the sibling input at runtime,
      // so this rule is a perpetual false positive for every labelled field. (a11y/SEO: off)
      '@angular-eslint/template/label-has-associated-control': 'off',
    },
  },
]);
