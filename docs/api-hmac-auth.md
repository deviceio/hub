# Summary

This document describes the methods used to provide strong Authentication 
against the HUB HTTP API. The Deviceio Hub provides extremely sensative access to 
connected devices and weak API security could present unacceptable risk to an
organization and potential system compromise. 

The Deviceio Hub attempts to provide strong authentication, tamper protection and
replay attack protection via asymetric request signing.

# Users

The only authentication point is a user. All requests MUST be associated with a 
user and as such each user contains one and only one set of credential factors to
authenticate requests. The Deviceio HUB does not seperate users from credentials to enforce
identity when conducting authentication.

As of this writing a "User" contains the following credential factors:

* `ID`: A unique v4 UUID of the user
* `Email`: The email address of the user. (ex: admin@localhost)
* `Login`: The login name of the user. (ex: admin)
* `PasswordHash`: SHA-512-hash(`<password-salt>+<user-password>`)
* `PasswordSalt`: value used as salt when constructing password hash
* `ED25519PublicKey`: public key used to verify request signatures
* `TOTPSecret`: shared secret used for TOTP passcode generation

All credential factors MUST be UNIQUE against any other user record. The Hub will prefer
to automatically generate these values (where possible) when conducting user administration.

# Authorization Header

All HTTP requests to the HUB API must include the `Authorization` header in form:

```
Authorization: DEVICEIO-HUB-AUTH <user-id>:<ed25519-signature-base64>
```

* `<user-id>` : Supply the user's ID, Email or Login to identify
the authorizing user. It is recommended to supply the users ID (v4 uuid).
* `<user-password` : Supply the user's known password.
* `<ed25519-signature-base64>` : The constructed ed25519 signature base64 encoded. 

# ed25519 Signature Construction

The ed25519 Signature is generated from a message which is a concatination of data values seperated by newlines (\r\n) 
which is subsequently hashed via SHA-512 and signed. The signature value is supplied along with the API reuqest. The Hub 
API will reconstruct this signature server-side from the incoming request and check that they match. If they do not match, 
the request will be rejected

Generate a message (string) as follows where `\r\n` are newlines:

```
<user-id>\r\n
<user-password>\r\n
<totp-passcode>\r\n
<http-scheme>\r\n
<http-method>\r\n
<http-host>\r\n
<http-path>\r\n
<http-query>\r\n
<http-content-type-header>\r\n
```

where

* `<user-id>` : Supply the user's ID, Email or Login to identify the authorizing user. It is recommended to supply the 
users ID (v4 uuid). If the user id supplied in the Authorization header does not exist in the generated signature the
API will reject the request
* `<totp-passcode>` : is a TOTP generated passcode from the user's `TOTPSecret`. 
this provides a time limited window (usually 30 seconds) in which a request with
this signature is valid. If an attacker attempts to re-use this signature after
the TOTP passcode expires the API will reject the request.
* `<http-scheme>` : The HTTP scheme (http -or- https) used for the request. If an
attacker is able to change the http scheme of the request and it differs from the signature the API will reject the request


