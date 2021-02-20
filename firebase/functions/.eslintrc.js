module.exports = {
  root: true,
  env: {
    es6: true,
    node: true,
  },
  extends: ["eslint:recommended"],
  rules: {
    quotes: ["error", "double"],
    indent: ["error", 2],
  },
  parserOptions: {
    ecmaVersion: 2018,
  },

};
