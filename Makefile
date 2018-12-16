VERSION := $(shell cat ./version)

clean:
	rm -r dist/ || true
.PHONY: clean

define gobuild
	CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go build \
		-ldflags='-X main.version=$(VERSION) -extldflags "-static"' \
		-o dist/kubedeploy-$(VERSION)-$(1)-$(2) .
endef

build: clean
	mkdir dist
	helm package --version $(VERSION) ./charts/app -d dist
	helm package --version $(VERSION) ./charts/kubedeploy -d dist
	$(call gobuild,darwin,386)
	$(call gobuild,darwin,amd64)
	$(call gobuild,linux,386)
	$(call gobuild,linux,amd64)
	$(call gobuild,linux,arm)
	$(call gobuild,linux,arm64)
	$(call gobuild,windows,386)
	$(call gobuild,windows,amd64)
	(cd dist; shasum -a 256 ./* > $(VERSION)-SHA256SUM)
.PHONY: build

release: build
	make -C shim release
	ghr $(VERSION) dist
.PHONY: release

bump:
	echo -n $(v) > version
.PHONY: bump

install:
	go build -o bin/kubedeploy .
.PHONY: install
