import { defineConfig } from "cypress";
import createBundler from "@bahmutov/cypress-esbuild-preprocessor";
import { addCucumberPreprocessorPlugin } from "@badeball/cypress-cucumber-preprocessor";
import { createEsbuildPlugin } from "@badeball/cypress-cucumber-preprocessor/esbuild";
import fs from "fs";

async function setupNodeEvents(
    on: Cypress.PluginEvents,
    config: Cypress.PluginConfigOptions
): Promise<Cypress.PluginConfigOptions> {
  // This is required for the preprocessor to be able to generate JSON reports after each run, and more,
  await addCucumberPreprocessorPlugin(on, config);

  on(
      "file:preprocessor",
      createBundler({
        plugins: [createEsbuildPlugin(config)],
      })
  );

  on('task', {
    deleteFolder (folderName) {
      return new Promise((resolve, reject) => {
        if (!fs.existsSync(folderName)) {
          return resolve(null);
        }
        fs.rm(folderName, { maxRetries: 10, recursive: true }, (err) => {
          if (err) {
            console.error(err);
            return reject(err)
          }

          resolve(null)
        })
      })
    },
  })

  // Make sure to return the config object as it might have been modified by the plugin.
  return config;
}

export default defineConfig({
  e2e: {
    specPattern: "features/**/*.feature",
    baseUrl: "http://localhost:8020",
    supportFile: false,
    setupNodeEvents,
    env: {
      omitFiltered: true,
      filterSpecs: true,
    },
  },
});
