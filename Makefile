.PHONY: fetch
fetch:
	go run cmd/fetch/main.go -req ./cmd/fetch/samples/input.txt -resp ./cmd/fetch/samples/output.txt localhost:8080

.PHONY: gohttpd
gohttpd:
	go run cmd/gohttpd/main.go ./docroot_dirs/htdocs1 8080

.PHONY: tritonhttpd
tritonhttpd:
	go run cmd/tritonhttpd/main.go -port 8080 -vh_config ./virtual_hosts.yaml -docroot ./docroot_dirs

.PHONY: submission
submission:
	go mod tidy
	rm -f submission.zip
	zip -r submission.zip . -x /.git/*