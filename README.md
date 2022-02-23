# Quickstart

```
$ caddy file-server --root ./public/ --listen :2000
```

```
$ fd .  | entr -c ./build
```

Don't forget that the preview API version is built using the `preview` branch!

You need to trigger a daily build so that the available dates are updated. Currently this is done via a Zapier daily Zap.
