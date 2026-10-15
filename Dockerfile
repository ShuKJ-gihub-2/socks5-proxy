FROM ginuerzh/gost:latest

# ユーザー名「myuser」パスワード「mypassword」で暗号化プロキシを起動
CMD ["sh", "-c", "/bin/gost -L=http://myuser:mypassword@:8080"]
