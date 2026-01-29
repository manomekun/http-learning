# 試す方法

```bash
# サーバー起動
go run step01/server.go

# 別ターミナルでテスト
echo 'Hello, TCP!' | nc localhost 8080
# → "Hello, TCP!" が返ってくれば成功！
```
