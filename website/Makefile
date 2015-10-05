.PHONY: build push deploy

NAME := echo
TAG := $(shell date +'%Y.%m.%dT%H.%M')

build:
	rm -rf public
	hugo
	gox -osarch=linux/amd64 -output server
	docker build -t gcr.io/$(PROJECT_ID)/$(NAME):$(TAG) .

push: build
	gcloud docker push gcr.io/$(PROJECT_ID)/$(NAME):$(TAG)

deploy: push
	# https://github.com/kubernetes/kubernetes/blob/release-1.0/docs/user-guide/kubectl/kubectl_rolling-update.md
	kubectl rolling-update $(NAME) --image=gcr.io/$(PROJECT_ID)/$(NAME):$(TAG)
