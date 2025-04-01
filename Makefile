ifneq (,$(filter $(OS),Windows_NT Windows))
	EXEEXT=.exe
endif

TARGET_WITH_EXT=tt${EXEEXT}
TARGET_NO_EXT=tt

.PHONY: all
all:: tt

.PHONY: tt
tt:
	go build -o $(TARGET_WITH_EXT)

.PHONY: linux
linux:
	GOOS=linux GOARCH=amd64 go build -o $(TARGET_NO_EXT)
