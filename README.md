# Supabase Admin API Server

To run on KPS and administer configs for services:

- Kong
- Gotrue
- Realtime
- Postgrest
- Pg-listen

## ENV

requires a `.env` in the project root with:

```bash
JWT_SECRET=<project-jwt-secret>
```

## API Interface

### Auth

You must set the `apikey` header to be a valid JWT token, signed by JWT_SECRET and with a claim of: `role: supabase_admin`

### Configs

GET `/config/postgrest` - returns current config `{ raw_contents: <string-of-file-contents> }`

POST `/config/postgrest` - sets new config - params: `{ raw_contents: <string-of-file-contents>, restart_services : <bool> }`

GET `/config/kong` - returns current config `{ raw_contents: <string-of-file-contents> }`

POST `/config/kong` - sets new config - params: `{ raw_contents: <string-of-file-contents>, restart_services : <bool> }`

GET `/config/pglisten` - returns current config `{ raw_contents: <string-of-file-contents> }`

POST `/config/pglisten` - sets new config - params: `{ raw_contents: <string-of-file-contents>, restart_services : <bool> }`

GET `/config/realtime` - returns current config `{ raw_contents: <string-of-file-contents> }`

POST `/config/realtime` - sets new config - params: `{ raw_contents: <string-of-file-contents>, restart_services : <bool> }`

GET `/config/gotrue` - returns current config as `{ raw_contents: <string-of-file-contents> }`

POST `/config/gotrue` - sets new config - params: `{ raw_contents: <string-of-file-contents>, restart_services : <bool> }`

### Restarting

GET `/services/restart` - re-reads all configs and restarts all services

GET `/services/reboot` - reboot the server

### Logs

GET `/logs/<application>/<head|tail>/<max_lines>` - get logs for a given application (postgrest,kong,admin,gotrue,syslog,pglisten)

## Sponsors

We are building the features of Firebase using enterprise-grade, open source products. We support existing communities wherever possible, and if the products don’t exist we build them and open source them ourselves. Thanks to these sponsors who are making the OSS ecosystem better for everyone.

[![New Sponsor](https://user-images.githubusercontent.com/10214025/90518111-e74bbb00-e198-11ea-8f88-c9e3c1aa4b5b.png)](https://github.com/sponsors/supabase)
