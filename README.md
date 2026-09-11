# LANFile

> Simple, zero-setup file sharing over your local network.

LANFile is a lightweight file-sharing server designed for **LAN use**, not public hosting. It lets anyone on your network upload and download files using nothing more than a web browser or command-line tools like `curl` and `wget`.

No accounts, no advanced configuration, no external client-side tools.

---

## Features

### Available

* [x] Upload files via CLI (`curl`, `wget`)
* [x] Upload files via the Web UI
* [x] Download files via CLI
* [x] Download files via the Web UI
* [x] Password-based file encryption/decryption in the Web UI
* [x] GPG encryption/decryption support via CLI
* [x] Authorization for overwriting files


### Planned

* [ ] Pastebin-style snippet sharing via CLI and Web UI

---

## Screenshots

### Web UI

![LANFile Web UI](media/home.png)

### CLI Demo

![LANFile CLI Demo](media/demo.gif)

---

## API

### Download

Download a file by name or ID.

| Method | Endpoint             | Description                        |
| ------ | -------------------- | ---------------------------------- |
| `GET`  | `/api/d/name/{filename}` | Download by filename (when unique) |
| `GET`  | `/api/d/id/{fileID}`     | Download by file ID                |

Example:

```bash
curl -O server/api/d/name/example.txt
```

#### Authentication

Files can be protected with a password. If a file is password-protected, you must provide the correct password to download it using HTTP Basic Auth.

```bash
curl -O -u :{password} https://server/api/d/id/{FileID}
```

---

### Upload

Upload a file by name or ID.

| Method | Endpoint             | Description                           |
| ------ | -------------------- | ------------------------------------- | 
| `PUT`  | `/api/u/name/{filename}` | Upload a file with a specific filename|
| `PUT`  | `/api/u/id/{fileID}`     | Update an existing file by ID         |

Example:

```bash
curl -T example.txt "server/api/u/name/example.txt"
```

#### Authentication

Using HTTP Basic Auth, you can provide a password to stop unauthorized access (overwriting and downloading) of a file. The password is hashed and stored in the database, so it is not recoverable. If you forget the password, you will need to delete the file and re-upload it.

e.g.

```bash
curl -T example.txt -u :{password} "https://server/api/u/id/{FileID}"
```

#### Query Parameters

| Parameter    | Values                    | Default | Description                                                  |
| ------------ | ------------------------- | ------- | ------------------------------------------------------------ |
| `encrypted` | `key`, `password`, `none` | `none`  | Indicates the file's encryption method (informational only). |
| `overwrite`  | `true`, `false`           | `false` | Overwrite an existing file with the same name for `/api/u/name/{filename}`. |

---

### Search

Search for files by name or ID.

| Method | Endpoint               | Description                               |
| ------ | ---------------------- | ----------------------------------------- |
| `GET`  | `/api/s/name/{searchTerm}` | Search by filename (substring by default) |
| `GET`  | `/api/s/id/{searchTerm}`   | Search by exact file ID                   |
| `GET`  | `/api/s/getall`            | List all stored file metadata             |

#### Query Parameters

| Parameter | Values          | Default | Description                                                 |
| --------- | --------------- | ------- | ----------------------------------------------------------- |
| `exact`   | `true`, `false` | `false` | Require an exact filename match for `/api/s/name/{searchTerm}`. |

Example:

```bash
curl "server/api/s/name/report?exact=true"
```

---

## Security

LANFile is intended for **trusted local networks**. It is **not** designed as a public-facing file-sharing service. By default, anyone with access to the webserver can upload, overwrite, and download files without authentication, which allows for an attacker to supply-chain attack potentialy malicious files to your network or read sensitive data.

To better secure your files and yourself, you can use the following methods:

* **Password protection**: Use HTTP Basic Auth to set a password for files. This prevents unauthorized access to overwrite or download files. See the [Upload](#upload) section for more details.
* **Encryption**: Encrypt files before uploading them. LANFile allows encryption via key-based or password-based encryption, which can be done using GPG on the CLI or GPG through the web interface. 
