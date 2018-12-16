build:
	docker build -t coldog/kubedeploy-shim:latest .
.PHONY: build

release: build
	docker push coldog/kubedeploy-shim:latest
.PHONY: release
