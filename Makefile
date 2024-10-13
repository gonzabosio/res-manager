build:
	GOARCH=wasm GOOS=js go build -o web/app.wasm
	go build -o view/home

run: build 
	./view/home