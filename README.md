# restmail -- sendmail for email rest APIs

restmail is a sendmail-compatible CLI using gmail & outlook rest APIs. It can be used
for email automation, notifications and with `git send-mail` for sharing patches. 

restmail requires only minimal authorization scope to send mail.  No access to reading
or deleting mail is needed.

## Install
```
# install to ~/go/bin ($GOPATH)
go install github.com/tonymet/restmail/cmd/restmail@latest
```

## Initial Provider Setup (gmail or outlook)
Initial setup is done once / provider. This configures oauth2 clientID & Secret

```
FROM=your.name@gmail.com
CLIENT_ID=xxxxxx
CLIENT_SECRET=yyyy
restmail -configClient -provider gmail -clientId "${CLIENT_ID}" \ 
   -clientSecret "${CLIENT_SECRET}" -f "${FROM}"
```

## Authorization Setup
Once provider config is set up above, you need to do the oauth flow through
a web browser.  Setup is only done once or if the tokens are invalidated. 
```
FROM=your.name@gmail.com
restmail -authorize -provider gmail -f "${FROM}"
```
When this is complete, your auth-token & refresh-token are saved and refreshed
automatically.

## Sending Messages
```
FROM=your.name@gmail.com
TO=friend@gmail.com
CC=billy@gmail.com
echo "subject: test subject\n\ntest message" | restmail -f "${FROM}" -provider gmail "${TO}" cc:"${CC}"

```

## Git send-mail configuration

```
[sendemail]
        smtpServer = /home/USERNAME/go/bin/restmail
        smtpServerOption = -f=your.name@gmail.com 
        smtpServerOption = -provider 
        smtpServerOption = gmail
```

## Running Via Container & Cloud Storage

restmail can be run via container with Google Cloud storage

### Build 
```
 docker build . -t containermail 
```

### Env  & Permissions
```
GCS_PREFIX=test/restmail
GCS_BUCKET=YOUR_BUCKET
```
GCS Perms will load using APP Default Credentials.  Be sure to authorize `gcloud auth login` before proceeding.
The user will need role/Storage Object User for saving & loading the tokens

### Setup OAuth Config & Token

re-run the OAuth Config with `-storage gcs`. This will save config to your bucket

```
FROM=your.name@gmail.com
CLIENT_ID=xxxxxx
CLIENT_SECRET=yyyy
restmail -configClient -storage gcs -provider gmail -clientId "${CLIENT_ID}" \ 
   -clientSecret "${CLIENT_SECRET}" -f "${FROM}"
```

re-run the auth step to save the token. this only needs doing once (refresh token will save)

```
FROM=your.name@gmail.com
restmail -authorize -storage gcs -provider gmail -f "${FROM}"
```

## Launch the Container

You can run the container on cloud run with args `-provider $PROVIDER -f $FROM -m "$MESSAGE" $TO_ADDR`






## Related Projects
* [sendgmail & gmail-oauth-tools](https://github.com/google/gmail-oauth2-tools) send mail via sendmail-compatible
CLI to google via SMTP
