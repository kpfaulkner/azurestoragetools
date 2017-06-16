#go build -ldflags "-X main.Version=0.1.0" main.go

gox -ldflags "-X main.Version=0.1.0"
