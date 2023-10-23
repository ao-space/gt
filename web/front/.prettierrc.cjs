// @see: https://www.prettier.cn

module.exports = {
  // Specify the maximum line length
  printWidth: 130,
  // Indentation tab width | number of spaces
  tabWidth: 2,
  // Use tabs instead of spaces for indentation (true: tab, false: space)
  useTabs: false,
  // Use a semicolon at the end (true: yes, false: no)
  semi: true,
  // Use single quotes (true: single quotes, false: double quotes)
  singleQuote: false,
  // Decide whether to wrap property names with quotes in object literals. Options "<as-needed|consistent|preserve>"
  quoteProps: "as-needed",
  // Use single quotes in JSX instead of double quotes (true: single quotes, false: double quotes)
  jsxSingleQuote: false,
  // Print trailing commas when possible in multiline. Options "<none|es5|all>"
  trailingComma: "none",
  // Add space between object, array brackets and text "{ foo: bar }" (true: yes, false: no)
  bracketSpacing: true,
  // Place the > of a multi-line element at the end of the last line instead of on its own line (true: at the end, false: on its own line)
  bracketSameLine: false,
  // (x) => {} Whether to have parentheses in arrow functions with a single parameter (avoid: omit parentheses, always: do not omit)
  arrowParens: "avoid",
  // Specify the parser to use, no need to write @prettier at the beginning of the file
  requirePragma: false,
  // Insert a special marker at the top of the file to indicate that the file has been formatted with Prettier
  insertPragma: false,
  // Control whether the text should be wrapped and how to wrap
  proseWrap: "preserve",
  // Whether whitespace in html is sensitive "css" - respects the default value of the CSS display property, "strict" - whitespace is considered sensitive, "ignore" - whitespace is considered insensitive
  htmlWhitespaceSensitivity: "css",
  // Control the indentation of code inside <script> and <style> tags in Vue single-file components
  vueIndentScriptAndStyle: false,
  // Use lf as the line ending. Options "<auto|lf|crlf|cr>"
  endOfLine: "auto",
  // These two options can be used to format code that starts and ends at given character offsets (respectively inclusive and exclusive) (rangeStart: start, rangeEnd: end)
  rangeStart: 0,
  rangeEnd: Infinity
};
