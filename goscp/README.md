# GO-SCP

A simple binary to help copying files with SCP from any machine (Windows for instance) to any other place using password auth

This tool can only UPLOAD via SCP.

## Compile 
```
go build
```

## Running

Copy all images from `/home/Pictures` that starts with `a` and ends with `png` to `server-host-name:/tmp`:

```
goscp -host [ssh-server-host-name] -username [user-name] -password [password] -from /home/Pictures -match a*png$ -to /tmp
```

It is also possible to avoid setting the password from the CLI if `GOSCP_PASSWORD` env variable is set.