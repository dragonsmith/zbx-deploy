TMP_TARGET = /tmp/zbx-deploy-server
all: clean install build

install:
	gom install

clean:
	rm -rf bin/

build:
	gom build -o bin/server main.go zabbix.go config.go
