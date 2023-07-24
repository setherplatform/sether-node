.PHONY: all
all: sether

GOPROXY ?= "https://proxy.golang.org,direct"
.PHONY: sether
sether:
	GIT_COMMIT=`git rev-list -1 HEAD 2>/dev/null || echo ""` && \
	GIT_DATE=`git log -1 --date=short --pretty=format:%ct 2>/dev/null || echo ""` && \
	GOPROXY=$(GOPROXY) \
	go build \
	    -ldflags "-s -w -X github.com/setherplatform/sether-node/cmd/sether/launcher.gitCommit=$${GIT_COMMIT} -X github.com/setherplatform/sether-node/cmd/sether/launcher.gitDate=$${GIT_DATE}" \
	    -o build/sether \
	    ./cmd/sether

	go build \
    	    -ldflags "-s -w -X github.com/setherplatform/sether-node/cmd/sether/launcher.gitCommit=$${GIT_COMMIT} -X github.com/setherplatform/sether-node/cmd/sether/launcher.gitDate=$${GIT_DATE}" \
    	    -o build/devp2p \
    	    ./cmd/devp2p

.PHONY: clean
clean:
	rm -fr ./build/*
