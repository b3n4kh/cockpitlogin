# Cockpitlogin Broker

This service provides a daemon that sets a random password for a user to allow
users without password to be able to login into cockpit.

The user has to be authenticated on a nginx Webserver.
Preferrable via either Client Certifactes, or via Basic Auth.

The Service needs to be run as a user that is permitted to execute:

```
/usr/bin/sudo /usr/bin/passwd
```

