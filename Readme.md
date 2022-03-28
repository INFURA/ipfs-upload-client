`ipfs-upload-client` is a minimal CLI tool to upload files or directories to Infura's IPFS or another API endpoint.

## Example

`ipfs-upload-client --id xxxxx --secret yyyyy /path/to/data`

## Installation

Pre-compiled binaries are available in the [latest release page](https://github.com/INFURA/ipfs-upload-client/releases/latest).

## Options
```
  --id string       your Infura ProjectID
  --pin             whether or not to pin the data (default true)
  --secret string   your Infura ProjectSecret
  --url string      the API URL (default "https://ipfs.infura.io:5001")
```

### Load options from configuration file

Create `.infura-ipfs-upload-client` file in your home directory.

```
ID=<YOUR INFURA PROJECT ID>
PIN=<WHETHER OR NOT TO PIN THE DATA OR NOT. (true / false)>
SECRET=<YOUR INFURA PROJECT SECRET>
URL=<THE API URL>
```