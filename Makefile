DOCKER_IMAGE_REPO=jagochiu/jable
# DOCKER_IMAGE_VERSION=0.0.0.1
DOCKER_IMAGE_VERSION=latest

run:
	touch resources
	go mod tidy 
	go mod vendor 
	go run main.go
debug:
	docker run --rm \
	  --name jable \
	  -v ${PWD}/main.go:/go/Jable/main.go \
	  -v ${PWD}/go.mod:/go/Jable/go.mod \
	  -v ${PWD}/bin:/go/Jable/bin \
	  -e TZ=Asia/Kuala_Lumpur \
	  --network main \
	  --shm-size 128m \
	  --workdir "/go/Jable" \
	  -it ${DOCKER_IMAGE_REPO}:${DOCKER_IMAGE_VERSION} \
	  /bin/bash

down:
	docker-compose down
build:
	docker rmi -f ${DOCKER_IMAGE_REPO}:${DOCKER_IMAGE_VERSION}
	docker build -f Dockerfile -t ${DOCKER_IMAGE_REPO}:${DOCKER_IMAGE_VERSION} . --no-cache
deploy:
	docker push ${DOCKER_IMAGE_REPO}:${DOCKER_IMAGE_VERSION}
init:
	sudo rm -rf config data logs 
	mkdir -p {config,data,logs}
	chmod -R 777 config data logs 
upgrade:
	make down && docker rmi -f ${DOCKER_IMAGE_REPO}:${DOCKER_IMAGE_VERSION} && make run 