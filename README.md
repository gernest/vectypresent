# vectypresent

Go presentation tool with vecty frontend

[demo](https://vectypresent.bq.co.tz)

This is a fork of [go present tool](https://github.com/golang/tools/tree/master/cmd/present) that uses vecty as the frontend.

# Installation

```
go get github.com/gernest/vectypresent
```

# Usage

```
vectypresent /path/to/directory/with/*.slides
```

A quick way to test is to clone the go talks  repo

```
git clone https://github.com/golang/talks.git

vectypresent talks/
```

Open your browser on localhost:8080 to browse for the slides presentation


## TODO

- [x] render slides
- [x] render articles
- [x] render directories
- [x] render raw files
- [ ] render notes
