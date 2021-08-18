#!/bin/bash
PROJECT_NAME=microservice
PROJECT_TAG=latest
go mod tidy
_buildtime="$(date '+%Y-%m-%d-%H-%M-%S')"
_commitid=${CI_COMMIT_SHORT_SHA}
go build -ldflags '-X "main.commitid='"${_commitid}"'" -X "main.buildtime='"${_buildtime}"'" -X "main.version='"${PROJECT_TAG}"'"' -o ${PROJECT_NAME}_${PROJECT_TAG}

cat > run.sh <<EOF
#!/bin/bash 
set -e 
exec ./${PROJECT_NAME}_${PROJECT_TAG}
EOF

cat > dockerfile <<EOF
FROM centos:7
MAINTAINER AriesWang 
ADD ${PROJECT_NAME}_${PROJECT_TAG} ${PROJECT_NAME}_${PROJECT_TAG} 
ADD run.sh run.sh 
RUN ls -all
RUN chmod +x /${PROJECT_NAME}_${PROJECT_TAG} 
RUN chmod +x /run.sh 
ENTRYPOINT ["/run.sh"]
EOF

docker rmi -f $(docker images | grep ${PROJECT_NAME})
docker build -t ${PROJECT_NAME}:${PROJECT_TAG} .
docker tag ${PROJECT_NAME}:${PROJECT_TAG} arieswangdocker/${PROJECT_NAME}:${PROJECT_TAG}
docker push arieswangdocker/${PROJECT_NAME}:${PROJECT_TAG}

rm ${PROJECT_NAME}_${PROJECT_TAG}
rm dockerfile
rm run.sh