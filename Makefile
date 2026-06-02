.PHONY: test test-unit test-pbt test-fuzz lint clean

# 全テスト実行 (Fuzzing 以外)
test:
	go test ./...

# ユニットテストのみ (_Property で終わるものを除く)
test-unit:
	go test -skip '_Property$$' ./...

# Property-Based Test のみ
test-pbt:
	go test -run '_Property' ./...

# Fuzzing (デフォルト 30 秒)
FUZZ_TIME ?= 30s
test-fuzz:
	go test -fuzz=FuzzParse -fuzztime=$(FUZZ_TIME) ./jid/...
	go test -fuzz=FuzzParserFeed -fuzztime=$(FUZZ_TIME) ./xmlstream/...
	go test -fuzz=FuzzConnectionReceive -fuzztime=$(FUZZ_TIME) ./protocol/...

# 静的解析
lint:
	staticcheck ./...

# キャッシュクリア
clean:
	go clean -testcache
