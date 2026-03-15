.PHONY=build
build:
	go build -v -o shorter.exe ./cmd/shorter
.PHONY=start
	$env:CONFIG_PATH="local.yaml"; ./shorter.exe