# LND-Fuzz

LND-Fuzz is a Go native fuzzing tool that automatically detects and runs fuzz targets available in your repository. It is designed to run multiple fuzzing processes concurrently and persist the generated input corpus, helping you continuously test and improve your codebase's resilience.

## Features

- **Automatic Fuzz Target Detection:** Scans your repository and identifies all available fuzz targets.
- **Concurrent Fuzzing:** Runs multiple fuzzing processes concurrently, with the default set to the number of available CPU cores.
- **Customizable Execution:** Configure the duration and target package for fuzzing with environment variables.
- **Corpus Persistence:** Saves the input corpus for each fuzz target into a designated Git storage repository, ensuring your test cases are retained for future runs.

## Environment Variables

Configure LND-Fuzz by setting the following environment variables in `.env` file:

- **FUZZ_NUM_PROCESSES**  
  Specifies the number of fuzzing processes to run concurrently.  
  _Default_: Maximum number of CPU cores available on the machine.

- **PROJECT_SRC_PATH** (_Required_)  
  The Git repository URL of the project to be fuzzed. Use one of the following formats:

  - For private repositories:  
    `https://oauth2:PAT@github.com/OWNER/REPO.git`
  - For public repositories:  
    `https://github.com/OWNER/REPO.git`

- **GIT_STORAGE_REPO** (_Required_)  
  The Git repository where the input corpus is stored. The URL should follow the format as `https://oauth2:PAT@github.com/OWNER/STORAGEREPO.git`.

- **FUZZ_TIME**  
  The duration (in seconds) for which the fuzzing engine should run.  
  _Default_: 120 Seconds.

- **FUZZ_PKG** (_Required_)
  The specific Go package within the repository that will be fuzzed.

## How It Works

1. **Configuration:**  
   Set the required environment variables in `.env` file to configure the fuzzing process.

2. **Fuzz Target Detection:**  
   The tool automatically detects all available fuzz targets in the provided project repository.

3. **Fuzzing Execution:**  
   Go's native fuzzing is executed on each detected fuzz target. The number of concurrent fuzzing processes is controlled by the `FUZZ_NUM_PROCESSES` variable.

4. **Corpus Persistence:**  
   For each fuzz target, the fuzzing engine generates an input corpus. If a `GIT_STORAGE_REPO` is provided, this corpus is persisted to the specified repository, ensuring that your test inputs are saved and can be reused in future runs.

## Usage

1. **Clone or Download LND-Fuzz:**

   ```bash
   git clone github.com/NishantBansal2003/LND-Fuzz.git
   cd LND-Fuzz
   ```

2. **Set Environment Variables:**  
   You can export the necessary environment variables in `.env` file:

   ```bash
   export FUZZ_NUM_PROCESSES=<number_of_processes>
   export PROJECT_SRC_PATH=<project_repo_url>
   export GIT_STORAGE_REPO=<storage_repo_url>
   export FUZZ_TIME=<time_in_seconds>
   export FUZZ_PKG=<target_package>
   ```

3. **Run the Fuzzing Engine:**  
   With your environment configured, start the fuzzing process. Run:
   ```bash
   go run main.go
   ```
