# Ivy iOS App source

This directory contains the source code to the Ivy iOS app.

To build, first create the Mobile.xcframework out of the Go
implementation of Ivy. Run:

```
go install golang.org/x/mobile/cmd/gomobile@latest
go install golang.org/x/mobile/cmd/gobind@latest
```

to install `gomobile` and `gobind`. Then:

```
mkdir work; cd work
go mod init work
go get -d golang.org/x/mobile/bind@latest
go get -d robpike.io/ivy/mobile
gomobile bind -target=ios,iossimulator,maccatalyst,macos robpike.io/ivy/mobile robpike.io/ivy/demo
```

Place the Mobile.xcframework directory in this directory, and
then open ivy.xcodeproj in Xcode.

You have to specify Development Team for code signing certificate in:
Project Settings -> Targets -> Signing & Capabilities -> Signing -> Team.
