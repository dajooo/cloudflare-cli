# Cloudflare CLI

A modern, intuitive command-line interface for interacting with the Cloudflare API. Built with Go, Cobra, and Lipgloss for a polished user experience.

![Screenshot of cf whoami command](https://github.com/dajooo/cloudflare-cli/blob/main/assets/whoami.png)

## ✨ Features

-   **Secure Authentication**: `login`, `logout`, and `whoami` commands with credentials stored securely in your OS's native keyring.
-   **Account Management**: List all accessible accounts.
-   **Zone Management**: List, create, and delete DNS zones.
-   **DNS Record Management**: Full CRUD operations (Create, List, Update, Delete) for DNS records.
-   **Cache Management**: Purge the cache for entire zones, specific files, or tags.
-   **Interactive Prompts**: User-friendly prompts for login and confirmations.
-   **Environment Variable Support**: Configure via a YAML file or environment variables (`CF_API_TOKEN`, etc.).
-   **Modern UI**: Beautifully styled output with light/dark mode support.

## 🚀 Installation

### Using `go install`

If you have a Go environment set up, you can install the CLI with:

```sh
go install dario.lol/cf@latest
````

### From Binaries

Pre-compiled binaries for various operating systems will be available on the [Releases](https://github.com/dajooo/cloudflare-cli/releases) page.

## Usage

### 1\. Login

First, authenticate with your Cloudflare account. The recommended method is using an API Token.

```sh
cf login
```

The CLI will prompt you for your preferred authentication method and credentials. Your token will be securely stored in your system's keyring.

### Required Permissions

To use all features of this CLI (including future updates), we recommend the following permissions. Code marked with `*` is for planned features.

-   **Zone / Zone**: Read (Edit required for `create`/`delete`)
-   **Zone / DNS**: Edit
-   **Zone / Cache Purge**: Purge
-   **Zone / SSL and Certificates**: Edit
-   **Zone / Zone Settings**: Read (Edit required for SSL/TLS settings)
-   **Zone / Firewall Services**: Edit *
-   **Account / Account Settings**: Read
-   **Account / Workers Scripts**: Edit *
-   **Account / Cloudflare Pages**: Edit *
-   **User / User Details**: Read

> [!NOTE]
> **User** permissions are only accessible on User API Tokens, not on Account-bound API Tokens.

### 2\. General Commands

Once logged in, you can manage your Cloudflare resources.

```sh
# Verify your identity and token status
cf whoami

# List all your accounts
cf account list

# List all zones
cf zone list

# Create a new zone
cf zone create example.com

# Delete a zone (with a confirmation prompt)
cf zone delete example.com

# List DNS records for a zone
cf dns list example.com --type A

# Create a new DNS record
cf dns create example.com www A 1.2.3.4 --proxied

# Purge the entire cache for a zone
cf cache purge --zone example.com --all

# Get current SSL mode
cf ssl get example.com

# Set SSL mode
cf ssl set example.com full
```

For a full list of commands and options, use the `--help` flag:

```sh
cf --help
```

## ⚙️ Configuration

The CLI can be configured via a configuration file or environment variables.

### Config File

Upon successful login, a configuration file is created at `~/.cloudflare-cli.yaml`. The API credentials within this file are encrypted.

### Environment Variables

You can override the settings in the config file using the following environment variables:

- `CF_API_TOKEN`: Your Cloudflare API Token.
- `CF_API_KEY`: Your Cloudflare Global API Key (legacy).
- `CF_API_EMAIL`: Your Cloudflare account email (used with the Global API Key).

## 🤝 Contributing

Contributions are welcome\! Please feel free to open an issue or submit a pull request.

1.  Fork the repository.
2.  Create a new branch (`git checkout -b feature/your-feature`).
3.  Make your changes.
4.  Commit your changes (`git commit -am 'Add new feature'`).
5.  Push to the branch (`git push origin feature/your-feature`).
6.  Create a new Pull Request.

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

```
