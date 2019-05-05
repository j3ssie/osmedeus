#### WARNING: You're need to create your own cert by this command.

Delete cert.pem and key.pem in this folder and create your own cert by this command below.

```
openssl req -x509 -newkey rsa:4096 -nodes -out cert.pem -keyout key.pem -days 365
```
