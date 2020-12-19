#!/bin/bash
go generate github.com/gen2brain/url2img/url2img
qtminimal linux $GOPATH/src/github.com/gen2brain/url2img/url2img
go install -tags minimal github.com/gen2brain/url2img/cmd/url2img
