# Continuous Fuzzing for LND

## Installation Instructions

### Step 1: Install Go 1.24.0

1. Visit the official Go download page: [Go Downloads](https://go.dev/dl).
2. Download and install the appropriate version for your OS and hardware architecture.

### Step 2: Add GOBIN Path to Your $PATH

1. Open your terminal.
2. Add the following lines to your shell profile file (e.g., `~/.bashrc`, `~/.zshrc`, `~/.profile`):

```sh
export GOROOT="/usr/local/go" # your go installation path.
export GOPATH="$HOME/go"
export GOBIN="$GOPATH/bin"
export PATH="$PATH:$GOBIN:$GOROOT/bin"
```

3. Reload your shell profile:

```sh
source ~/.bashrc
```

### Step 3: Install Make Command

1. Ensure `make` is installed on your system. On most Unix-based systems, `make` is pre-installed. If not, install it using your package manager.
   - **Ubuntu/Debian**: `sudo apt-get install build-essential`
   - **MacOS**: `xcode-select --install`

### Step 4: Build the Continuous-Fuzz app

1. Run the following command to build the Continuous-Fuzz app:

```sh
make build
```

### Step 5: Run the Continuous-Fuzz App

1. Make sure the required environment variables are set.
For more details, see: [docs/USAGE.md](USAGE.md)
2. Run the following command to run the Continuous-Fuzz app:

```sh
make run
```

### Step 6: Run the Test Cases

1. Run the following command to execute the test cases:

```sh
make test
```
