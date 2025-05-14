# LND-Fuzz

LND-Fuzz is a Go native fuzzing tool that automatically detects and runs fuzz targets available in your repository. It is designed to run multiple fuzzing processes concurrently and persist the generated input corpus, helping you continuously test and improve your codebase's resilience.

## Features

- **Automatic Fuzz Target Detection:** Scans your repository and identifies all available fuzz targets.
- **Concurrent Fuzzing:** Runs multiple fuzzing processes concurrently, with the default set to the number of available CPU cores.
- **Customizable Execution:** Configure the duration and target package for fuzzing with environment variables.
- **Corpus Persistence:** Saves the input corpus for each fuzz target into a designated Git storage repository, ensuring your test cases are retained for future runs.

## Deployment & Execution

LND-Fuzz is designed with built-in coordination logic, eliminating the need for external CI frameworks like Jenkins or Buildbot. It can be deployed as a long-running service on any cloud instance (e.g., AWS EC2, GCP Compute Engine, or DigitalOcean Droplet). Once initiated, the application autonomously manages its execution cycles, running continuously and restarting the fuzzing process at intervals defined by the `FUZZ_TIME` environment variable.

This self-sufficient design simplifies deployment and ensures consistent fuzz testing without manual intervention.
