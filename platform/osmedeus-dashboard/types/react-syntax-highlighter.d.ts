declare module "react-syntax-highlighter" {
  import * as React from "react";
  export interface SyntaxHighlighterProps {
    language?: string;
    style?: any;
    customStyle?: React.CSSProperties;
    showLineNumbers?: boolean;
    children?: string;
  }
  export const Light: React.ComponentType<SyntaxHighlighterProps> & {
    registerLanguage: (name: string, language: any) => void;
  };
}

declare module "react-syntax-highlighter/dist/esm/languages/hljs/yaml" {
  const yaml: any;
  export default yaml;
}

declare module "react-syntax-highlighter/dist/esm/languages/hljs/bash" {
  const bash: any;
  export default bash;
}

declare module "react-syntax-highlighter/dist/esm/languages/hljs/json" {
  const json: any;
  export default json;
}

declare module "react-syntax-highlighter/dist/esm/languages/hljs/javascript" {
  const javascript: any;
  export default javascript;
}

declare module "react-syntax-highlighter/dist/esm/languages/hljs/markdown" {
  const markdown: any;
  export default markdown;
}

declare module "react-syntax-highlighter/dist/esm/languages/hljs/xml" {
  const xml: any;
  export default xml;
}

declare module "react-syntax-highlighter/dist/esm/styles/hljs" {
  export const atomOneDark: any;
  export const github: any;
}

declare module "react-syntax-highlighter/dist/esm/styles/hljs/github" {
  const github: any;
  export default github;
}

declare module "react-syntax-highlighter/dist/esm/styles/hljs/atom-one-dark" {
  const atomOneDark: any;
  export default atomOneDark;
}
