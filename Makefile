GOPROXY ?= "https://proxy.golang.org,direct"
GIT_COMMIT?=$(shell git rev-list -1 HEAD | xargs git rev-parse --short)
GIT_DATE?=$(shell git log -1 --date=short --pretty=format:%ct)
VERSION=1.0.0-rc.1-$(GIT_COMMIT)-$(GIT_DATE)
DOCKER_IMAGE=sether/node:$(VERSION)

.PHONY: all
all: sether

sether:
	@echo "Building version: $(VERSION)"
	go build \
	    -ldflags "-s -w -X github.com/setherplatform/sether-node/cmd/sether/launcher.gitCommit=$(GIT_COMMIT) -X github.com/setherplatform/sether-node/cmd/sether/launcher.gitDate=$(GIT_DATE)" \
	    -o build/sether \
	    ./cmd/sether


devp2p:
	go build \
		-ldflags "-s -w -X github.com/setherplatform/sether-node/cmd/sether/launcher.gitCommit=$(GIT_COMMIT) -X github.com/setherplatform/sether-node/cmd/sether/launcher.gitDate=$(GIT_DATE)" \
		-o build/devp2p \
		./cmd/devp2p

clean:
	rm -f ./build/sether
	rm -f ./build/devp2p

docker: docker_build docker_tag

docker_build:
	docker build --build-arg "GIT_COMMIT=$(GIT_COMMIT)" --build-arg "GIT_DATE=$(GIT_DATE)" . -t $(DOCKER_IMAGE)

docker_tag:
	docker tag $(DOCKER_IMAGE) sether/node:latest

check_changes:
	@if ! git diff-index --quiet HEAD --; then \
		echo "You have uncommitted changes. Please commit or stash them before making a release."; \
		exit 1; \
	fi

tag_release:
	git tag -a $(VERSION) -m "Release $(VERSION)"

push_changes:
	git push --tags

release: check_changes tag_release push_changes docker
	docker login && \
	docker image push $(DOCKER_IMAGE) && \
	docker image push sether/node:latest && \
	docker logout
