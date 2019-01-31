# go-sub

Find and download subtitles

`go-sub` is a little script to find and download subtitles for TV series

## Install

    go get -v github.com/bewie/go-sub
    cd $GOPATH/src/github.com/bewie/go-sub
    go install

## Synopsis

    $GOPATH/bin/go-sub -help
    $GOPATH/bin/go-sub -lang en -p /path/to/my/movies/directory

## Features

* Parse filename to detect show, season and episode
* Search and download appropriate subtitle from opensubtitles.org, extracting it from the archive and renaming it
* Accept a lang option (defaults to fr)
