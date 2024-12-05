# restmail

restmail sends email from CLI using gmail & outlook rest APIs. It can be used
for email automation, notifications and with `git send-mail` for sharing patches. 

restmail requires only minimal authorization scope to send mail.  No access to reading
or deleting mail is requested.

## Install
```
# install to ~/go/bin ($GOPATH)
go install github.com/tonymet/restmail@latest
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
restmail -setup -provider gmail -sender "${FROM}"
```
When this is complete, your auth-token & refresh-token are saved and refreshed
automatically.

## Sending Messages
```
FROM=your.name@gmail.com
TO=friend@gmail.com
CC=billy@gmail.com
echo "subject: test subject\n\ntest messagee" | go run . -f "${FROM} -provider gmail "${TO}" cc:"${CC}"

```

## Git send-mail configuration

```
[sendemail]
        smtpServer = /home/USERNAME/go/bin/restmail
        smtpServerOption = -f=your.name@gmail.com 
        smtpServerOption = -provider 
        smtpServerOption = gmail
```

## Related Projects
* [sendgmail & gmail-oauth-tools](https://github.com/google/gmail-oauth2-tools) send mail via sendmail-compatible
CLI to google via SMTP