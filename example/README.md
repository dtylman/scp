## SCP Example

### Compile 
```
go build
```

### Running
`example` will use `scp` to copy a remote file locally. Local files created on the working folder with the remote base name.

Usage: host:port username remotepath

#### Notes

Port must be set!

Password will be prompted for.


```
localhost$ ./example naboo.local:22 root /var/log/lastlog
Password: ********
Opening tcp to naboo.local:22
Establishing ssh session naboo.local:22...
Copied 292584 bytes in 327.874073ms
```
