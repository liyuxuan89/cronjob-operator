img=liyuxuan89/cronjob:v1.0.0

make docker-build IMG=$img
docker save -o cronjob-controller.tar $img
minikube cp ./cronjob-controller.tar minikube:/home/docker/cronjob-controller.tar
minikube ssh -n minikube -- "docker rmi $img"
minikube ssh -n minikube -- 'docker load -i cronjob-controller.tar'

minikube cp ./cronjob-controller.tar minikube-m02:/home/docker/cronjob-controller.tar
minikube ssh -n minikube -- "docker rmi $img"
minikube ssh -n minikube-m02 -- 'docker load -i cronjob-controller.tar'
docker rmi $(docker images| grep "none" | awk '{print $3}')

make deploy IMG=$img