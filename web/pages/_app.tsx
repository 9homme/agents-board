import "@/styles/globals.css";
// highlight.js CSS theme for syntax-highlighted code fences (US003)
import "highlight.js/styles/github.css";
import type { AppProps } from "next/app";

export default function App({ Component, pageProps }: AppProps) {
  return <Component {...pageProps} />;
}
