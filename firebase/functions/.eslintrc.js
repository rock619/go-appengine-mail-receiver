module.exports = {
  root: true,
  env: {
    es6: true,
    node: true,
  },
  extends: ['eslint:recommended', 'google', 'prettier'],
  rules: {
    quotes: ['error', 'single'],
    indent: ['error', 2],
  },
  parserOptions: {
    ecmaVersion: 2018,
  },
};
