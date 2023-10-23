// @see: https://stylelint.io

module.exports = {
  root: true,
  // Inherit some predefined rules
  extends: [
    "stylelint-config-standard", // Configure stylelint extension plugin
    "stylelint-config-html/vue", // Configure style formatting in Vue's template
    "stylelint-config-standard-scss", // Configure stylelint SCSS plugin
    "stylelint-config-recommended-vue/scss", // Configure SCSS style formatting in Vue
    "stylelint-config-recess-order" // Configure stylelint CSS property writing order plugin
  ],
  overrides: [
    // Scan styles inside <style> tags in .vue/html files
    {
      files: ["**/*.{vue,html}"],
      customSyntax: "postcss-html"
    }
  ],
  rules: {
    "function-url-quotes": "always", // URL quotes "always (must be quoted)" | "never (no quotes)"
    "color-hex-length": "long", // Specify hexadecimal color shorthand or expansion "short (hexadecimal shorthand)" | "long (hexadecimal expansion)"
    "rule-empty-line-before": "never", // Require or disallow an empty line before rules "always (always must have a line before a rule)" | "never (never have a line before a rule)" | "always-multi-line (always must have a line before multi-line rules)" | "never-multi-line (never have a line before multi-line rules)"
    "font-family-no-missing-generic-family-keyword": null, // Disallow missing generic family keyword in font family name list
    "scss/at-import-partial-extension": null, // Resolve the issue of not being able to use @import to import SCSS files
    "property-no-unknown": null, // Disallow unknown properties
    "no-empty-source": null, // Disallow empty source code
    "selector-class-pattern": null, // Enforce selector class name format
    "value-no-vendor-prefix": null, // Turn off vendor-prefix (to resolve multi-line ellipsis -webkit-box)
    "no-descending-specificity": null, // Disallow selectors of lower specificity from appearing after overriding selectors of higher specificity
    "value-keyword-case": null, // Resolve the error of using v-bind uppercase words in SCSS
    "selector-pseudo-class-no-unknown": [
      true,
      {
        ignorePseudoClasses: ["global", "v-deep", "deep"]
      }
    ]
  },
  ignoreFiles: ["**/*.js", "**/*.jsx", "**/*.tsx", "**/*.ts"]
};
