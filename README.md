# Library install
```
export GOPATH={GlobalGopath}
dep ensure
```

# Dev
AppGopath is ./gopath

```
export GOPATH={AppGopath}
ln -s `pwd`/vendor `pwd`/gopath/src
dev_appserver.py app.yaml
```

# Deploy
```
export GOPATH={AppGopath}
ln -s `pwd`/vendor `pwd`/gopath/src
gcloud app deploy
gcloud app deploy cron.yaml
gcloud app browse
```