# mdm-xv

📱 **Mobile Device Management: Extended View**

A Jamf-compatible CLI that gives IT admins a fast, browser-free way to look up Apple devices by serial number or email, view FileVault status, and retrieve hardware and enrollment info — all in one readable terminal table.

> Built for speed. Built for visibility. Built for people who hate clicking through the Jamf UI.

## About

**mdm-xv** is a lightweight command-line tool for Jamf Pro admins who want instant visibility into their Apple fleet. It authenticates via your Jamf Pro API, securely stores your credentials, and returns clean, color-coded tables with critical device information — all without opening a browser.

With `mdm-xv`, you can:

- 🔍 Search by serial number or email address
- 🧾 View FileVault status, RAM, OS version, last check-in, and more
- 🔐 Authenticate once — token caching and secure keychain support included
- 🧹 Reset all saved credentials with a single command

Whether you're cleaning up stale records or just need quick intel, `mdm-xv` lets you skip the UI and get straight to the data.

## Key Features

- 🔐 **Secure Token Handling** – Authenticates via Jamf Pro API with token caching and Keychain storage
- 📧 **Email & Serial Lookup** – Pull device data by user email or serial number
- 📊 **Readable Output** – Terminal-friendly tables with bold headers and 🔴 red FileVault indicators
- 📨 **Email-less Device Discovery** – Leave the email field blank to return all devices without an assigned email (great for data hygiene)
- 🧹 **Reset Command** – Instantly wipe stored credentials and token from disk + Keychain
- 💨 **No Jamf UI Required** – Everything runs straight from the terminal

## Installation

### 🛠 Build from source

Make sure you have Go installed (`go version`), then run:

```bash
git clone https://github.com/susbot/mdm-xv
cd mdm-xv
go build -o mdm-xv mdm-xv.go
```  

### 📦 Optional: Add to PATH (Recommended)

If you want to run `mdm-xv` from any location in your terminal, move the binary to a directory that's in your system's `$PATH`:

```bash
sudo mv mdm-xv /usr/local/bin/

```
## Usage

### 🔐 Token

Generates or reuses a bearer token for Jamf Pro API access.

- If a valid token is already cached, it will be reused automatically.
- If no token is found or it’s expired, you'll be prompted to enter your Jamf username, password, and server URL.
- Credentials are stored securely using the system keychain.
- The token is cached at:  
  `~/Library/Application Support/mdm-xv/mdm_token.json`

```bash
mdm-xv token
```

### 📧 Email

Looks up devices associated with a user's email address.

- You’ll be prompted to enter a user email (e.g., `jdoe@example.com`)
- Up to **10 matching devices** will be returned in a table
- Each device includes:
    - Serial number
    - Model
    - OS version
    - FileVault status (color-coded 🔴 red if off, 🟢 green if on)
    - Last check-in
    - Enrollment date
    - RAM, MAC address, and more

```bash
mdm-xv email
```
### 🧹 Reset

Clears all saved credentials and the cached API token from your system.

- Deletes the bearer token stored at:  
  `~/Library/Application Support/mdm-xv/mdm_token.json`
- Removes the Jamf Pro username, password, and URL from your system keychain
- Useful when switching Jamf tenants, rotating credentials, or troubleshooting login issues

```bash
mdm-xv reset
```

## 🔒 Security & Data Handling

Your credentials and token are handled with security in mind:

- 🔐 **Credentials** (username, password, URL) are stored using your system’s secure keychain via [`go-keyring`](https://github.com/zalando/go-keyring)
- 🧾 **Bearer token** is saved locally at:  
  `~/Library/Application Support/mdm-xv/mdm_token.json`
- 📡 No credentials or data are ever sent anywhere except your own Jamf Pro API server
- 🧠 All actions are performed **locally on your machine** — there is no telemetry, logging, or analytics of any kind
- 🧹 You can clear all stored data at any time using:

```bash
mdm-xv reset
```

## ⚠️ Disclaimer

**mdm-xv** is an independent open source project and is **not affiliated with, endorsed by, or sponsored by Jamf** or Jamf Software, LLC.

This tool is provided **“as is”**, without warranty of any kind — express or implied.  
Use of this software is at your own risk.

The author assumes **no liability or responsibility** for any damages or data loss that may occur through the use or misuse of this tool.

Please ensure you comply with your organization's MDM policies and security requirements before using this tool in a production environment.

⚠️ This tool is provided as-is. While I intend to maintain and improve it, I may discontinue support or development at any time without notice.

## 📚 Dependencies & Acknowledgments

This tool uses the following open source Go libraries:

- [`urfave/cli/v2`](https://github.com/urfave/cli) – For CLI command structure and flags
- [`olekukonko/tablewriter`](https://github.com/olekukonko/tablewriter) – For rendering pretty tables in the terminal
- [`zalando/go-keyring`](https://github.com/zalando/go-keyring) – For securely storing credentials in the system keychain
- [`golang.org/x/term`](https://pkg.go.dev/golang.org/x/term) – For secure password input (no echo in terminal)

All libraries are used under their respective open source licenses.

## License

This project is licensed under the Creative Commons Attribution-NonCommercial 4.0 International License (CC BY-NC 4.0).

You may view, use, and modify the source code for educational and personal purposes, but commercial use or redistribution is strictly prohibited without written permission from the author.

See [LICENSE](./LICENSE) for full details.  
See [NOTICE.md](./NOTICE.md) for third-party license acknowledgments.

## 🤝 Contributions

This project is open source for transparency and educational purposes.  
At this time, I’m not accepting external pull requests as all development is being handled internally.

You're welcome to fork the code, explore it, or suggest ideas via Issues.

Thanks for understanding and supporting the project!


## 💡 Support This Project

If you’ve found **mdm-xv** helpful, consider supporting its development:

- ☕ [Buy Me a Coffee](https://buymeacoffee.com/susbot)
- 💾 [Buy the CLI on Gumroad](https://susbot.gumroad.com/l/jwskbq)
- 🌐 [Visit the Official Website](https://mdm-xv.com/)

