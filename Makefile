testapi:
	go test -v -cover ./api

testauth:
	go test -v -cover ./auth

testutil:
	go test -v -cover ./util


testall:
	make testutil testauth testapi testdb

testdb:
	go test -v -cover ./db

init-test-db:
	sudo docker-compose -f ./db/test/docker-compose.yaml up -d

teardown-test-container:
	sudo docker stop $$(sudo docker ps -aq)
	sudo docker rm $$(sudo docker ps -aq)

interact-db:
	 sudo docker exec -it $$(sudo docker ps -aq) bash

