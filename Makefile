help:
	@echo "Usage: make [help|build|test|loc|godep-save]"

build:
	go build

test:
	go test

loc:
	@git archive --format=zip master > tmp-loc.zip
	@cloc tmp-loc.zip
	@rm tmp-loc.zip

godep-save:
	godep save
