# 3-way handshake の実装

やってみた

# 動かす方法

```
docker run --cap-add=NET_RAW --cap-add=NET_ADMIN \
  -v $(pwd):/app \
  -w /app \
  -it golang:1.25-trixie bash
```

コンテナ内で、必要なツールをインストールする

その後、パケットがファイアウォールに遮断されないようにする
```bash
apt-get update && apt-get install -y tcpdump netcat-traditional nftables
nft add table ip filter
nft add chain ip filter output { type filter hook output priority 0 \; }
nft add rule ip filter output tcp flags rst drop
```


別のターミナルで (`docker exec -it [CONTAINER] bash` とかで入る)、tcpdump もしておく

```bash
tcpdump -i lo -n tcp port 12345
```

あとは実行するだけ
```bash
go run .
```
