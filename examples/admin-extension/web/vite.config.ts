import { defineConfig } from "vite";
import path from "path";

// Vite configuration for Forge Admin extensions.
// Builds a library targeting ES module output with entry point index.js,
// externalizing React and mapping it to the global ForgeAdmin.react.
export default defineConfig({
  build: {
    outDir: "dist",
    emptyOutDir: true,
    lib: {
      entry: path.resolve(__dirname, "src/index.tsx"),
      name: "AnalyticsExtension",
      formats: ["es"],
      fileName: () => "index.js",
    },
    rollupOptions: {
      external: ["react", "react-dom"],
      output: {
        globals: {
          react: "ForgeAdmin.react",
        },
      },
    },
  },
});
