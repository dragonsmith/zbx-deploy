TMP_TARGET = /tmp/zbx-deploy-server
all: clean install build

install:
	gom install

clean:
	rm -rf bin/

build:
	gom build -o bin/server main.go zabbix.go config.go

deploy:
	GOOS=linux gom install
	GOOS=linux gom build -o $(TMP_TARGET) main.go zabbix.go config.go

	scp $(TMP_TARGET) $(ZBX_DEPLOY_TARGET)
