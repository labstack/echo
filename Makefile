DEP_VERSION=0.4.1

dependency:
	curl -fsSL -o ${GOPATH}/bin/dep https://github.com/golang/dep/releases/download/v${DEP_VERSION}/dep-linux-amd64
	chmod +x ${GOPATH}/bin/dep
	dep ensure

test:
	echo "" > coverage.txt
	for d in $(shell go list ./... | grep -v vendor); do \
		go test -race -coverprofile=profile.out -covermode=atomic $$d || exit 1; \
		[ -f profile.out ] && cat profile.out >> coverage.txt && rm profile.out; \
	done

tag:
	@git tag `grep -P '^\tversion = ' echo.go|cut -f2 -d'"'`
	@git tag|grep -v ^v
