FROM ginuerzh/gost:latest

# 待ち受けるポートをRender指定の「10000」に変更します
CMD ["sh", "-c", "/bin/gost -L=http://admin:password@:10000"]
