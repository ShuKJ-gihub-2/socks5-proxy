FROM ginuerzh/gost:latest
CMD ["sh", "-c", "/bin/gost -L=socks5://:10000"]
