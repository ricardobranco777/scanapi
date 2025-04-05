FROM	docker.io/library/golang AS builder

WORKDIR	/go/src/scanapi
COPY	. .

RUN	make

FROM	scratch
COPY	--from=builder /go/src/scanapi/scanapi /usr/local/bin/scanapi

ENTRYPOINT ["/usr/local/bin/scanapi"]
