build-app:
	cd frontend && ng build

mv-app:
	mv frontend/dist cmd/httpd

build-svc:
	go build -o feature-httpd cmd/httpd/main.go

cleanup:
	rm -rf cmd/httpd/dist

build: build-app mv-app build-svc cleanup