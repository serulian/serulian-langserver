PLATFORMS = \
	darwin-amd64 \
	linux-arm \
	linux-amd64 \
	windows-amd64

.PHONY: $(PLATFORMS) clean

all:
	for i in $(PLATFORMS); do \
		$(MAKE) $$i; \
	done

$(PLATFORMS): TARGET_OS=$$(echo $@ | cut -f1 -d-)
$(PLATFORMS): TARGET_ARCH=$$(echo $@ | cut -f2 -d-)
$(PLATFORMS):
	mkdir -p releases/$(TARGET_OS)-$(TARGET_ARCH)/
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build -o releases/$(TARGET_OS)-$(TARGET_ARCH)/serulian-langserver .

clean:
	$(RM) -rf releases