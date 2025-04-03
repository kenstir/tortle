ifneq (,$(filter $(OS),Windows_NT Windows))
	EXE=.exe
endif

LOCAL_TARGET=tt${EXE}
LINUX_TARGET=tt

.PHONY: all
all:: tt

.PHONY: tt
tt:
	go build -o $(LOCAL_TARGET)

.PHONY: linux
linux:
	GOOS=linux GOARCH=amd64 go build -o $(LINUX_TARGET)
