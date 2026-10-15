FROM ginuerzh/gost:latest

# プログラムが眠らないようにバックグラウンドで起動し、動き続けさせます
ENTRYPOINT ["/bin/gost"]
CMD ["-L=http://admin:password@:10000"]
