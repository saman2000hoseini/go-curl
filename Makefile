export APP=go-curl

export LDFLAGS="-w -s"

build:
	go build -ldflags $(LDFLAGS)

install:
	go install -ldflags $(LDFLAGS)