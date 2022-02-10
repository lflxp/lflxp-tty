install: 
	cp cmd/main.go .
	go install
	rm -f main.go
	lflxp-tty -w top

build: asset
	cd cmd && go build && mv cmd ../lflxp-tty

asset: bindata clean
	cd pkg/static && go-bindata -o=../asset.go -pkg=pkg ./

bindata:
	@echo 安装预制环境
	go get -u github.com/jteeuwen/go-bindata/...
	go get -u github.com/elazarl/go-bindata-assetfs/...
	go get -u github.com/swaggo/swag/cmd/swag

push: asset pull
	git add .
	git commit -m "${m}"
	git push origin $(shell git branch|grep '*'|awk '{print $$2}')

pull:
	git pull origin $(shell git branch|grep '*'|awk '{print $$2}')

crt: csr
	openssl x509 -req -sha256 -days 3650 -in tls/server.csr -signkey tls/server.key -out tls/server.crt

csr: key
	openssl req -nodes -new -key tls/server.key -subj "/CN=www.lflxp.cn" -out tls/server.csr

key: clean tls
	openssl genrsa -out tls/server.key 2048

tls:
	mkdir -p tls

clean:
	rm -f cmd/cmd
	rm -f lflxp-tty
	rm -rf tls