testapi:
	go test -v -cover ./api

testauth:
	go test -v -cover ./auth

testutil:
	go test -v -cover ./util


testall:
	make testutil testauth testapi

testdb:
	go test -v -cover ./db

init-test-db:
	sudo docker-compose -f ./db/test/docker-compose.yaml up -d

remove-test-db:
	sudo docker stop $$(sudo docker ps -aq)
	sudo docker rm $$(sudo docker ps -aq)

interact-db:
	 sudo docker exec -it $$(sudo docker ps -aq) bash

