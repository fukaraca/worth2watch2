testapi:
	go test -v -cover ./api

testauth:
	go test -v -cover ./auth

testutil:
	go test -v -cover ./util

#we need a running psql for test. so, use testall
testdb:
	go test -v -cover ./db

# go test -v ./... can result like "permission denied" due to ./db/volume folder which is persistent volume that's created by docker-compose
testall:
	make init-test-container testapi testauth testutil testdb teardown-test-container

init-test-container:
	sudo docker-compose -f ./db/test/docker-compose.yaml up -d

teardown-test-container:
	sudo docker stop $$(sudo docker ps -aqf "name= test-postgres-container")
	sudo docker rm $$(sudo docker ps -aqf "name= test-postgres-container")

interact-db:
	 sudo docker exec -it $$(sudo docker ps -aq) bash
