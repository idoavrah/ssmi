<p align="center">
  <a href="https://github.com/idoavrah/ssmi/releases/latest">
    <img src="https://img.shields.io/github/v/release/idoavrah/ssmi" />
  </a>
  <a href="https://github.com/idoavrah/ssmi/actions/workflows/build-and-release.yaml">
    <img src="https://github.com/idoavrah/ssmi/actions/workflows/build-and-release.yaml/badge.svg" />
  </a>
  <a href="https://goreportcard.com/report/github.com/idoavrah/ssmi">
    <img src="https://goreportcard.com/badge/github.com/idoavrah/ssmi" />
  </a>
  <img src="https://img.shields.io/github/downloads/idoavrah/ssmi/total" />
  <a href="https://github.com/idoavrah/ssmi/blob/main/LICENSE.txt">
    <img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" />
  </a>
</p>

## SSMI (SSM Into) - Your SSM sidekick

**SSMI** (pronounced "sesame", as in "open sesame") is a CLI tool that helps you log into supported EC2 instances with SSM.

### ✨ Features

- SSMI supports multiple AWS profiles.
- SSMI saves your connection history for easy access.
- SSMI allows connecting using different user accounts.

> **_NOTE:_** If you are using [granted.dev](https://granted.dev), please read [this](https://docs.commonfate.io/granted/recipes/credential-process). If you don't, you really should, it's amazing.

### 🚀 Usage

| Option                 | Description                                                                 |
|------------------------|-----------------------------------------------------------------------------|
| `--profile [profile]`  | Specify the AWS profile to use, otherwise AWS_PROFILE env var will be used  |
| `--offline`            | Run in offline mode, no telemetry will be sent.                             |
| `--version`            | Show version information and exit.                                          |
| `--help`               | Show help message and exit.                                                 |

### 💾 Installation

- Install it using `brew install idoavrah/ssmi`.
- Or, download the latest release from the [releases page](https://github.com/idoavrah/ssmi/releases/latest).
- Or, if you prefer to build it yourself, clone the repo, execute `make build` locally and copy the binary to your PATH.

### 

### 📺 Screenshot

![ssmi-screenshot](screenshot.png)

### Usage Tracking
- SSMI utilizes [PostHog](https://posthog.com) to track usage of the application.
- This is done to help us understand if the tool is being used and possibly understand how it can be improved.
- No personal data is being sent to the tracking service.
- You can opt-out of usage tracking completely by setting the --offline flag when running the tool.
