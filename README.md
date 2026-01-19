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
-   **Account / Workers Scripts**: Edit
-   **Account / Cloudflare Pages**: Edit
-   **Account / Workers R2 Storage**: Edit
-   **Account / D1**: Edit
-   **Account / Workers KV Storage**: Edit
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

### 3. Developer Platform

Manage your Workers, Pages, R2, D1, and KV resources.

```sh
# Workers (Coming Soon: Deploy)
cf workers --help

# R2 Buckets
cf r2 bucket list
cf r2 bucket create my-bucket
cf r2 bind my-bucket --to my-pages-project --name BUCKET

# D1 Databases
cf d1 create my-db
cf d1 list
cf d1 exec my-db -- "SELECT * FROM users"
cf d1 bind my-db --to my-pages-project --name DB

# KV Namespaces & Keys
cf kv namespace create "My App KV"
cf kv namespace list
cf kv bind "My App KV" --to my-pages-project --name KV
cf kv key put my-key "some value" --namespace-id <id>
cf kv key put my-key "some value" --namespace-id <id>
cf kv key get my-key --namespace-id <id>
```

For a full list of commands and options, use the `--help` flag:

```sh
cf --help
```

## ⚙️ Configuration

The CLI can be configured via a configuration file or environment variables.

### Config Storage

Upon successful login, your configuration is stored in a local database at `~/.cloudflare-cli/cf.db`. API credentials are encrypted using `age` (X25519), with the private identity stored securely in your system's native keyring.

### Environment Variables

You can override the settings in the configuration using the following environment variables:

- `CF_API_TOKEN`: Your Cloudflare API Token.
- `CF_API_KEY`: Your Cloudflare Global API Key (legacy).
- `CF_API_EMAIL`: Your Cloudflare account email (used with the Global API Key).
- `CF_ACCOUNT_ID`: Your Cloudflare Account ID.

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
