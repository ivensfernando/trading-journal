# KuCoin SDK credential checklist

The sample client in `cmd/main.go` uses the official KuCoin universal SDK. A common
error is `400004` (`Invalid KC-API-PASSPHRASE`), which indicates the headers were
signed with a passphrase that does not match the API key.

Confirm the following before running `make run`:

1. **Use the API passphrase** – it is the passphrase you set when creating the API key.
   It is *not* your trading password. If you regenerated a key, double-check that the
   environment variable was updated to the new passphrase.
2. **Match API key versions** – API keys created as API-V2 require the same passphrase
   value you entered in KuCoin; the SDK hashes this internally for signing. Supplying
   a pre-hashed value will cause a 400004 response.
3. **Environment variables must be set** – export the credentials before running the
   sample:
   ```bash
   export KUCOIN_API_KEY="<api-key>"
   export KUCOIN_API_SECRET="<api-secret>"
   export KUCOIN_API_PASSPHRASE="<api-passphrase>"  # plaintext value
   make run
   ```

If you still receive `Invalid KC-API-PASSPHRASE`, rotate the API key in KuCoin,
set the new values, and try again.
