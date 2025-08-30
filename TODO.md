### Tier 1: Core Functionality (MVP)

This tier represents the absolute essentials for a functional Cloudflare CLI.

- [x] **Authentication & Configuration**
    - [x] **`cf login`**
        - **Description:** Authenticates the user and securely stores credentials for subsequent commands. This is the entry point for using the CLI.
        - **Implementation:** Prompts for an API token. It will then use a library like `keyring` to store the credentials encrypted in the native OS secure storage (e.g., macOS Keychain, Windows Credential Manager).
        - **Example:** `cf login`
    - [x] **`cf logout`**
        - **Description:** Removes the stored credentials from the system.
        - **Example:** `cf logout`
    - [x] **`cf whoami`**
        - **Description:** Verifies the stored credentials by making a test API call and displays the current user and token status.
        - **Example:** `cf whoami`

- [x] **Account Management**
    - [x] **`cf account list`**
        - **Description:** Lists all accounts (organizations) the user has access to, which is crucial for users who manage multiple entities.
        - **Flags:** `--name` to filter by account name.
        - **Example:** `cf account list`

- [x] **Zone Management**
    - [x] **`cf zone list`**
        - **Description:** Lists all zones (domains) in a specific account. This is a primary command for getting context.
        - **Flags:** `--account-id`, `--status <active|pending>`.
        - **Example:** `cf zone list`
    - [x] **`cf zone create <domain>`**
        - **Description:** Adds a new domain to Cloudflare.
        - **Flags:** `--account-id`, `--jumpstart` to automatically scan for common DNS records.
        - **Example:** `cf zone create example.com --jumpstart`
    - [x] **`cf zone delete <zone_name|zone_id>`**
        - **Description:** Deletes a zone. This is a destructive operation and will require interactive confirmation.
        - **Flags:** `--yes` to bypass the confirmation prompt.
        - **Example:** `cf zone delete example.com`

- [x] **DNS Record Management**
    - [x] **`cf dns list <zone>`**
        - **Description:** Lists, searches, and filters DNS records for a given zone.
        - **Flags:** `--type <A|CNAME...>`, `--name <subdomain>`, `--content <ip/value>`.
        - **Example:** `cf dns list example.com --type A`
    - [x] **`cf dns create <zone> <name> <type> <content>`**
        - **Description:** Creates a new DNS record.
        - **Flags:** `--ttl <seconds>`, `--proxied`.
        - **Example:** `cf dns create example.com www A 1.2.3.4 --proxied`
    - [x] **`cf dns update <zone> <record_id>`**
        - **Description:** Updates an existing DNS record, identified by its ID.
        - **Flags:** `--name`, `--type`, `--content`, `--proxied <true|false>`.
        - **Example:** `cf dns update example.com <id> --content 5.6.7.8`
    - [x] **`cf dns delete <zone> <record_id>`**
        - **Description:** Deletes a DNS record.
        - **Example:** `cf dns delete example.com <id>`

- [x] **Cache Management**
    - [x] **`cf cache purge`**
        - **Description:** Purges the Cloudflare cache, which is essential for forcing updates to your site's content.
        - **Flags:** `--zone <zone>`, `--all`, `--files <url1,...>`, `--tags <tag1,...>`.
        - **Example:** `cf cache purge --zone example.com --all`

---

### Tier 2: Comprehensive Functionality

- [ ] **SSL/TLS Management**
    - [ ] **`cf ssl get <zone>`**: Retrieves the current SSL/TLS encryption mode.
    - [ ] **`cf ssl set <zone> <flexible|full|strict|off>`**: Sets the SSL/TLS encryption mode for a zone.

- [ ] **Firewall Rules**
    - [ ] **`cf firewall list <zone>`**: Lists all firewall rules for a zone.
    - [ ] **`cf firewall create <zone>`**: Creates a new firewall rule.
        - **Flags:** `--expression <exp>`, `--action <block|challenge|allow>`, `--description <desc>`.
        - **Example:** `cf firewall create example.com --expression "(ip.src eq 1.2.3.4)" --action block`

- [ ] **Developer Platform: Workers & Pages**
    - [ ] **`cf workers deploy <script_path>`**: Deploys a worker script. Will read configuration from a `wrangler.toml` file if present.
        - **Flags:** `--name <worker_name>`.
    - [ ] **`cf pages deploy <directory>`**: Deploys a folder of static assets to a Cloudflare Pages project.
        - **Flags:** `--project-name <name>`, `--branch <name>`.

---

### Tier 3: Complete API Coverage

- [ ] **Zero Trust / Access**
    - [ ] **`cf access app list/create/delete`**: Manages Access applications that protect internal sites. **(Enterprise Feature)**
    - [ ] **`cf tunnel list/create/run`**: Manages Cloudflare Tunnels for secure outbound connections, replacing `cloudflared`.

- [ ] **Advanced Security**
    - [ ] **`cf rate-limit list/create <zone>`**: Manages rate limiting rules to protect against denial-of-service attacks. **(Enterprise Feature)**
    - [ ] **`cf api-shield list/create <zone>`**: Manages API Shield for endpoint protection. **(Enterprise Feature)**
    - [ ] **`cf bot-management get/set <zone>`**: Manages Bot Fight Mode and Super Bot Fight Mode. **(Enterprise Feature for full control)**

- [ ] **Performance**
    - [ ] **`cf lb list/create`**: Manages Load Balancers. **(Enterprise Feature)**
    - [ ] **`cf argo get/set <zone>`**: Manages Argo Smart Routing. **(Add-on Feature)**

- [ ] **Full Developer Platform**
    - [ ] **R2**: `cf r2 bucket create/list`, `cf r2 object put/get <bucket> <key>`.
    - [ ] **D1**: `cf d1 create <name>`, `cf d1 exec <name> --query "..."`.
    - [ ] **KV**: `cf kv namespace create/list`, `cf kv key put/get <namespace> <key>`.

- [ ] **Monitoring & Logs**
    - [ ] **`cf logs tail <zone>`**: Live-streams HTTP request logs from the edge. **(Enterprise Feature)**
    - [ ] **`cf health-check list/create`**: Manages health checks for load balancing. **(Enterprise Feature)**