docker save -o cronjob-controller.tar liyuxuan89/cronjob:v1.0.0
minikube cp ./cronjob-controller.tar minikube-m02:/home/docker/cronjob-controller.tar
minikube ssh -n minikube-m02 -- 'docker load -i cronjob-controller.tar'
docker rmi $(docker images | grep "none" | awk '{print $3}')
