.PHONY: clean

OpenSesame_DoubleLock.zip: OpenSesame_DoubleLock
	zip OpenSesame_DoubleLock.zip OpenSesame_DoubleLock

OpenSesame_DoubleLock: main.go
	GOOS=linux GOARCH=amd64 go build -o OpenSesame_DoubleLock main.go

clean:
	rm -rf ./OpenSesame_DoubleLock ./OpenSesame_DoubleLock.zip

