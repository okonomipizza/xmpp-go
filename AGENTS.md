## プロジェクト概要

Go で実装された依存 0 かつSans I/O な XMPP ライブラリ

## スコープ

- **クライアント実装のみ** - サーバー (S2S) は対象外
- c2s (client-to-server) 接続のみサポート

## Agents

- Premature Optimization is the Root of All Evil
- If it hurts, do it more often
- Don't live with broken windows
- 妥協しないこと
- 忖度しないこと
- 常に日本語を利用すること
- 全角と半角の間には半角スペースを入れること
- 絵文字を使わないこと
- コメントは全て日本語にすること
- ログメッセージは全て英語にすること
- エラーメッセージは全て英語にすること
- テストメッセージは全て日本語にすること

## レビューについて
- 常に日本語を利用すること
- 厳しく行うこと
- 指摘内容を明確にすること
- 指摘は具体的にすること

## RFC について

- refs/ 以下にある RFC を参照すること

## テスト

### 方針
- Property-Based Testing を重視
- 任意のバイト列で panic しないことを保証
- RFC 記載のサンプルを conformance test として使用

### 命名規則
- `TestXxx` - 通常のユニットテスト
- `TestXxx_Property` - Property-Based Test
- `FuzzXxx` - Go 組み込み Fuzzing

### カバレッジ
- protocol パッケージは 90% 以上を目標
