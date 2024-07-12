server: buildsrv
	@./bin/server

tui: buildtui
	@./bin/tui

buildsrv:
	@go build -o ./bin/server cmd/main.go

buildtui:
	@cargo build --release
	@cp ./target/release/tcp-chat-tui ./bin/tui
