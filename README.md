# go-nfon-sso

package to login to the nfon portals with the nfon sso

## getting started

```sh
go get github.com/Lukas-Nielsen/go-nfon-sso
```

```go
import github.com/Lukas-Nielsen/go-nfon-sso
```

## usage

### conf

you need the portal base url eg. https://admin.nfon.com or https://start.cloudya.com and the client id eg. admin-portal or cloudya

```go
client, err := NewClient(<portalBaseUrl> string, <clientId> string);
```

### auth

#### login

```go
uri, err := client.Login(<username> string, <password> string);
```

#### otp

```go
err := client.OTP(<otp (6 digit)> string);
```

### token operation

```go
// refresh the token internally
err := client.RefreshToken()

// read token from json file
err := client.TokenFromJsonFile(</path/to/token/file.json> string)

// write token to json file
err := client.TokenToJsonFile(</path/to/token/file.json> string)

// set token
client.SetToken(<token object> Token)

// get token
token := client.GetToken()
```

### functions

#### get, delete

```go
*resty.Response, err := client.<get|delete>(<uri> string, <query> map[string]string, <header> map[string]string)
```

#### post, put, patch

```go
*resty.Response, err := client.<post|put|patch>(<uri> string, <payload> any, <query> map[string]string, <header> map[string]string)
```

## known possible client id's

### admin portal

- dts-admin-portal
- dts-admin-portal-preview
- dfn-admin-portal
- dfn-admin-portal-preview
- chess-admin-portal
- chess-admin-portal-preview
- dialog-telekom-admin-portal
- dialog-telekom-admin-portal-preview
- telekom-admin-portal
- telekom-admin-portal-preview
- o2-business-admin-portal
- o2-business-admin-portal-preview
- versatel-admin-portal
- versatel-admin-portal-preview
- smarticloud-admin-portal
- smarticloud-admin-portal-preview
- phoneup-admin-portal
- phoneup-admin-portal-preview
- admin-portal
- admin-portal-preview

### user portal

- centrexx
- cloudya
- dfn
- dialog-telekom
- o2
- telekom
- one-and-one
- promelit
- phoneup
- nconnect-voice
- sip-trunk
- o2-business-teams-telefonie
