FROM ginuerzh/gost:latest

# ユーザー認証なしの純粋なSOCKS5として、Renderのポート10000番で起動します
CMD ["sh", "-c", "/bin/gost -L=socks5://:10000"]
