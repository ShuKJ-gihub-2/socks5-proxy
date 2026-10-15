FROM ginuerzh/gost:latest
ENTRYPOINT ["/bin/gost"]
CMD ["-L=http://:10000"]
