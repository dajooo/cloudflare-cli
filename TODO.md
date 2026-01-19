### 1. Core Management
*Authentication, Account settings, and Team management.*

- [x] **`cf login`** `[Free]`
    - **Description:** Authenticates the user and securely stores credentials.
    - **Implementation:** Prompts for API token; uses `keyring` for secure storage.
    - **Example:** `cf login`
- [x] **`cf logout`** `[Free]`
    - **Description:** Removes the stored credentials from the system.
    - **Example:** `cf logout`
- [x] **`cf whoami`** `[Free]`
    - **Description:** Verifies credentials and displays current user/token status.
    - **Example:** `cf whoami`
- [x] **`cf account list`** `[Free]`
    - **Description:** Lists all accounts (organizations) the user has access to.
    - **Flags:** `--name` to filter by account name.
    - **Example:** `cf account list`
- [x] **`cf account switch <id>`** `[Free]`
    - **Description:** Switch the active context to a different account ID.
- [x] **`cf account members list`** `[Free]`
    - **Description:** View team members and their roles.
- [x] **`cf billing profile get`** `[Free]`
    - **Description:** View payment method and billing info.
- [x] **`cf audit-logs list`** `[Free/Ent]`
    - **Description:** View account changes (Basic for Free, Advanced for Ent).

---

### 2. Compute (Workers & Pages)
*Serverless functions and static sites.*

- [ ] **`cf workers deploy <script_path>`** `[Free]`
    - **Description:** Deploys a worker script. Reads `wrangler.toml` if present.
    - **Flags:** `--name <worker_name>`.
- [ ] **`cf workers tail <worker_name>`** `[Free]`
    - **Description:** Stream live logs from the worker to the console.
- [ ] **`cf workers secret put <key>`** `[Free]`
    - **Description:** Upload an encrypted environment variable/secret.
- [ ] **`cf pages deploy <directory>`** `[Free]`
    - **Description:** Deploys a folder of static assets to a Pages project.
    - **Flags:** `--project-name <name>`, `--branch <name>`.
- [ ] **`cf pages project list`** `[Free]`
    - **Description:** List all Pages projects.
- [ ] **`cf durable-objects create <name>`** `[Add-on]`
    - **Description:** Create a consistent storage object (Requires Workers Paid).

---

### 3. Storage & Databases
*Object storage, Key-Value, and SQL.*

- [x] **`cf r2 bucket create <name>`** `[Free]`
    - **Description:** Creates a new R2 storage bucket.
- [x] **`cf r2 bucket list`** `[Free]`
    - **Description:** Lists existing R2 buckets.
- [x] **`cf kv namespace create <name>`** `[Free]`
    - **Description:** Creates a KV namespace.
- [x] **`cf kv namespace list`** `[Free]`
    - **Description:** Lists KV namespaces.
- [x] **`cf kv key put <namespace> <key> <value>`** `[Free]`
    - **Description:** Writes a value to a KV key.
- [x] **`cf kv key get <namespace> <key>`** `[Free]`
    - **Description:** Reads a value from a KV key.
- [x] **`cf d1 create <name>`** `[Free]`
    - **Description:** Creates a D1 SQL database.
- [x] **`cf d1 exec <name> -- "<query>"`** `[Free]`
    - **Description:** Executes a SQL query against D1.
- [ ] **`cf d1 backup create <name>`** `[Free]`
    - **Description:** Manually triggers a database backup.
- [ ] **`cf queue create <name>`** `[Add-on]`
    - **Description:** Create a message queue (Requires Workers Paid).

---

### 4. Networking & Traffic
*DNS, Domains, and Traffic Routing.*

- [x] **`cf zone list`** `[Free]`
    - **Description:** Lists all zones (domains) in the account.
    - **Flags:** `--account-id`, `--status <active|pending>`.
- [x] **`cf zone create <domain>`** `[Free]`
    - **Description:** Adds a new domain to Cloudflare.
    - **Flags:** `--jumpstart` to scan DNS.
- [x] **`cf zone delete <zone>`** `[Free]`
    - **Description:** Deletes a zone.
- [x] **`cf zone settings <zone>`** `[Free]`
    - **Description:** View or toggle settings (e.g., Minify, Always Online).
- [x] **`cf dns list <zone>`** `[Free]`
    - **Description:** Lists DNS records.
    - **Flags:** `--type`, `--name`, `--content`.
- [x] **`cf dns create <zone>`** `[Free]`
    - **Description:** Creates a new DNS record.
    - **Flags:** `--ttl`, `--proxied`.
- [x] **`cf dns update <zone> <record>`** `[Free]`
    - **Description:** Updates an existing DNS record.
- [x] **`cf dns delete <zone> <record>`** `[Free]`
    - **Description:** Deletes a DNS record.
- [ ] **`cf cache purge`** `[Free]`
    - **Description:** Purges cache.
    - **Flags:** `--zone`, `--all`, `--files`, `--tags`.
- [ ] **`cf lb list`** `[Add-on]`
    - **Description:** List Load Balancers.
- [ ] **`cf lb monitor create`** `[Add-on]`
    - **Description:** Create a health check monitor for origins.
- [ ] **`cf argo enable <zone>`** `[Add-on]`
    - **Description:** Enable Argo Smart Routing.

---

### 5. Security (WAF & SSL)
*Firewall, Certificates, and Bot Protection.*

- [x] **`cf ssl get <zone>`** `[Free]`
    - **Description:** Retrieves the current SSL/TLS encryption mode.
- [x] **`cf ssl set <zone> <mode>`** `[Free]`
    - **Description:** Sets SSL mode (off, flexible, full, strict).
- [ ] **`cf ssl custom upload <zone>`** `[Biz]`
    - **Description:** Upload custom SSL certificates.
- [ ] **`cf firewall list <zone>`** `[Free]`
    - **Description:** Lists all firewall rules.
- [ ] **`cf firewall create <zone>`** `[Free]`
    - **Description:** Creates a new WAF rule.
    - **Flags:** `--expression`, `--action`, `--description`.
- [ ] **`cf rate-limit create <zone>`** `[Free/Add-on]`
    - **Description:** Create rate limiting rules (Advanced features require Biz/Ent).
- [ ] **`cf bot-management set <zone>`** `[Pro/Ent]`
    - **Description:** Configure Bot Fight Mode or Super Bot Fight Mode.
- [ ] **`cf api-shield create <zone>`** `[Ent]`
    - **Description:** Configure API Shield schema validation.

---

### 6. Zero Trust (Cloudflare One)
*Access, Gateway, and Tunnels.*

- [ ] **`cf access app list`** `[Free]`
    - **Description:** List Access applications.
- [ ] **`cf access policy create`** `[Free]`
    - **Description:** Create access policies (e.g., allow email@company.com).
- [ ] **`cf tunnel create <name>`** `[Free]`
    - **Description:** Create a secure tunnel to expose local services.
- [ ] **`cf tunnel run <name>`** `[Free]`
    - **Description:** Run the tunnel connector.
- [ ] **`cf gateway rule create`** `[Free]`
    - **Description:** Create DNS or Network filtering rules for teams.

---

### 7. specialized & Enterprise
*AI, Media, and Infrastructure.*

- [ ] **`cf ai run <model>`** `[Free]`
    - **Description:** Run inference on Workers AI models.
- [ ] **`cf image upload <file>`** `[Add-on]`
    - **Description:** Upload an image to Cloudflare Images.
- [ ] **`cf stream upload <file>`** `[Add-on]`
    - **Description:** Upload a video to Cloudflare Stream.
- [ ] **`cf logpush job create`** `[Ent]`
    - **Description:** Configure log pushing to S3/Splunk.
- [ ] **`cf magic-transit list`** `[Ent]`
    - **Description:** View Magic Transit configuration.