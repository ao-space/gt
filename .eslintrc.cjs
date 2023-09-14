// @see: http://eslint.cn

module.exports = {
  root: true,
  env: {
    browser: true,
    node: true,
    es6: true
  },
  // Specifies how to parse syntax
  parser: "vue-eslint-parser",
  // Has a lower priority than the parse syntax configuration
  parserOptions: {
    parser: "@typescript-eslint/parser",
    ecmaVersion: 2020,
    sourceType: "module",
    jsxPragma: "React",
    ecmaFeatures: {
      jsx: true
    }
  },
  // Inheritance of some existing rules
  extends: ["plugin:vue/vue3-recommended", "plugin:@typescript-eslint/recommended", "plugin:prettier/recommended"],
  /**
   * "off" 或 0    ==>  Turn off the rule
   * "warn" 或 1   ==>  Turn on the rule ad a warning
   * "error" 或 2  ==>  Rule as an error
   */
  rules: {
    // eslint (http://eslint.cn/docs/rules)
    "no-var": "error", // Required use let or const instead of var
    "no-multiple-empty-lines": ["error", { max: 1 }], // Disallow multiple empty lines
    "prefer-const": "off", //  For variables declared with the let keyword but never reassigned after the initial assignment, require the use of const
    "no-use-before-define": "off", // Prohibit the use of variables before they are defined

    // typeScript (https://typescript-eslint.io/rules)
    "@typescript-eslint/no-unused-vars": "error", // Prohibit the use of variables before they are defined
    "@typescript-eslint/prefer-ts-expect-error": "error", // prohibit the use of @ts-ignore
    "@typescript-eslint/ban-ts-comment": "error", // Prohibit the use of @ts-<directive> comments or require a description after the directive
    "@typescript-eslint/no-inferrable-types": "off", // Explicit types that can be easily inferred may add unnecessary verbosity
    "@typescript-eslint/no-namespace": "off", // Prohibit the use of custom TypeScript modules and namespaces
    "@typescript-eslint/no-explicit-any": "off", // Prohibit the use of the any type
    "@typescript-eslint/ban-types": "off", // Prohibit the use of specific types
    "@typescript-eslint/no-var-requires": "off", // Allow the use of the require() function to import modules
    "@typescript-eslint/no-empty-function": "off", // Prohibit empty functions
    "@typescript-eslint/no-non-null-assertion": "off", // Do not allow the use of the non-null assertion postfix operator(!)

    // vue (https://eslint.vuejs.org/rules)
    "vue/script-setup-uses-vars": "error", // Prevent variables used in <script setup> from being marked as unused in <template>, this rule is only effective when the no-unused-vars rule is enabled
    "vue/v-slot-style": "error", // Enforce the v-slot directive style
    "vue/no-mutating-props": "error", // Do not allow mutating component props
    "vue/custom-event-name-casing": "error", // Enforce a specific case for custom event names
    "vue/html-closing-bracket-newline": "error", // Require or disallow a newline before the right bracket of a tag
    "vue/attribute-hyphenation": "error", // Enforce attribute naming style for custom components in the template: my-prop="prop"
    "vue/attributes-order": "off", // Vue API usage order, enforce attribute order
    "vue/no-v-html": "off", // Prohibit the use of v-html
    "vue/require-default-prop": "off", // This rule requires a default value to be provided for each prop when it is required
    "vue/multi-word-component-names": "off" // Require component names to always be multi-word linked by "-"
  }
};
