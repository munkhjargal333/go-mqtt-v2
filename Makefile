airapp: 
	@air -c ./.air.app.toml

runapp: buildapp
	@./bin/app/main

runadmin: buildadmin
	@./bin/admin/main

buildapp:
	@go build ./cmd/all/main.go

