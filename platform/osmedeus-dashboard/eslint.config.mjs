import coreWebVitals from "eslint-config-next/core-web-vitals"

export default [
  ...coreWebVitals,
  {
    // eslint-config-next 16.4 turned on these React Compiler rules as errors.
    // They flag the fetch-in-mount-effect pattern every page uses on purpose
    // (see CLAUDE.md, "Org Scope"); moving the pages off it is a refactor of its
    // own, so they warn until that lands rather than block the build.
    rules: {
      "react-hooks/set-state-in-effect": "warn",
      "react-hooks/immutability": "warn",
    },
  },
]
